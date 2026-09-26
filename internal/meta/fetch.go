package meta

import (
	"context"
	"errors"
	"fmt"
	"image"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ApolloF/WaterLauncher/internal/library"
)

// Version marks what Fetch gathers; metadata from an older version is
// fetched again (2: backdrops; 3: full-size backdrops, round tiles, Steam's
// own art files when its store lists none).
const Version = 3

// reShotSize is the size Steam puts in a screenshot's file name.
var reShotSize = regexp.MustCompile(`\.\d+x\d+(\.jpg)`)

// Request says which game to fetch metadata for.
type Request struct {
	Title      string
	SteamAppID int
	GogID      string
	EpicApp    string        // namespace:item:appName, for games Steam doesn't have
	Keep       *library.Meta // previous metadata: user-chosen art is kept
}

// Fetch gathers metadata and art for one game. Missing pieces are simply
// left empty; an error means nothing at all could be fetched (and is
// ErrRateLimited when the caller should back off).
func (c *Client) Fetch(ctx context.Context, r Request) (*library.Meta, error) {
	m := &library.Meta{FetchedAt: time.Now().Unix(), Version: Version}
	var urls struct{ cover, hero, backdrop, logo, icon []string }
	var errs []error
	found := false

	if r.SteamAppID > 0 {
		if d, err := c.steamAppDetails(ctx, r.SteamAppID); err == nil {
			found = true
			m.Source = "Steam"
			m.Description = plainText(d.Short)
			m.Developers, m.Publishers = d.Developers, d.Publishers
			for _, g := range d.Genres {
				m.Genres = append(m.Genres, g.Description)
			}
			m.ReleaseDate = d.Release.Date
			m.ReleaseYear = yearOf(d.Release.Date)
			m.Controller = d.Controller
			for _, cat := range d.Categories {
				switch cat.ID {
				case catDualSense, catDualSenseBT:
					m.DualSense = "yes"
				case catDualShock:
					if m.DualSense == "" {
						m.DualSense = "dualshock"
					}
				}
			}
			if d.HeaderImage != "" {
				urls.hero = append(urls.hero, d.HeaderImage)
			}
			// The first screenshots, picked by the developer, make a sharp
			// backdrop that isn't the same picture as the game's tile. The
			// store lists them at 1920 wide; the file as uploaded (often
			// 4K) has the same name without the size.
			for _, s := range d.Screenshots[:min(2, len(d.Screenshots))] {
				if s.Full != "" {
					if orig := reShotSize.ReplaceAllString(s.Full, "$1"); orig != s.Full {
						urls.backdrop = append(urls.backdrop, orig)
					}
					urls.backdrop = append(urls.backdrop, s.Full)
				}
			}
		} else {
			errs = append(errs, err)
		}
		if a, err := c.steamArt(ctx, r.SteamAppID); err == nil {
			found = true
			if a.Cover != "" {
				urls.cover = append(urls.cover, a.Cover)
			}
			if a.Hero != "" {
				urls.hero = append([]string{a.Hero}, urls.hero...)
			}
			if a.Header != "" {
				urls.hero = append(urls.hero, a.Header)
			}
			urls.logo = append(urls.logo, a.LogoCandidates...)
		} else {
			errs = append(errs, err)
		}
		// Steam's CDN keeps an app's library art under fixed names, also
		// for games its store no longer lists: tried when nothing better
		// worked.
		base := fmt.Sprintf("%ssteam/apps/%d/", steamAssetBase, r.SteamAppID)
		urls.cover = append(urls.cover, base+"library_600x900_2x.jpg", base+"library_600x900.jpg")
		urls.hero = append(urls.hero, base+"library_hero.jpg", base+"header.jpg")
		urls.backdrop = append(urls.backdrop, base+"library_hero_2x.jpg")
	}
	if r.GogID != "" {
		if p, err := c.gogProduct(ctx, r.GogID); err == nil {
			found = true
			if m.Source == "" {
				m.Source = "GOG"
			}
			if m.Description == "" {
				m.Description = firstParagraph(plainText(p.Description.Lead + "\n" + p.Description.Full))
			}
			if m.ReleaseYear == 0 {
				m.ReleaseYear = yearOf(p.ReleaseDate)
			}
			if u := gogURL(p.Images.Background); u != "" {
				urls.hero = append(urls.hero, u)
			}
			if u := gogURL(p.Images.Icon); u != "" {
				urls.icon = append(urls.icon, u)
			}
		} else {
			errs = append(errs, err)
		}
	}
	// Epic's catalog, for games in an Epic library that Steam doesn't sell.
	if r.SteamAppID == 0 && r.EpicApp != "" {
		if e, err := c.epicGame(ctx, r.EpicApp); err == nil {
			found = true
			if m.Source == "" {
				m.Source = "Epic"
			}
			if m.Description == "" {
				m.Description = firstParagraph(plainText(e.Description))
			}
			if len(m.Developers) == 0 && e.Developer != "" {
				m.Developers = []string{e.Developer}
			}
			if m.ReleaseYear == 0 && len(e.ReleaseInfo) > 0 {
				m.ReleaseYear = yearOf(e.ReleaseInfo[0].DateAdded)
			}
			if u := e.image("DieselGameBoxTall", "OfferImageTall"); u != "" {
				urls.cover = append(urls.cover, u)
			}
			// The wide box art is 2560×1440: a backdrop as it is, and a hero.
			if u := e.image("DieselGameBox", "OfferImageWide", "DieselStoreFrontWide"); u != "" {
				urls.hero = append(urls.hero, u)
				urls.backdrop = append(urls.backdrop, u)
			}
			if u := e.image("DieselGameBoxLogo"); u != "" {
				urls.logo = append(urls.logo, u)
			}
		} else {
			errs = append(errs, err)
		}
	}
	// SteamGridDB fills whatever the stores didn't have.
	if len(urls.cover) == 0 || len(urls.hero) == 0 || len(urls.logo) == 0 {
		if id, err := c.sgdbGameID(ctx, r.SteamAppID, r.Title); err == nil {
			found = true
			if m.Source == "" {
				m.Source = "SteamGridDB"
			}
			for _, k := range []struct {
				kind string
				dst  *[]string
			}{{"grids", &urls.cover}, {"heroes", &urls.hero}, {"logos", &urls.logo}} {
				if len(*k.dst) > 0 {
					continue
				}
				if ims, err := c.sgdbArt(ctx, id, k.kind); err == nil && len(ims) > 0 {
					*k.dst = append(*k.dst, ims[0].URL)
				}
			}
		}
	}
	if !found {
		// Busy: try again later rather than settle for less.
		for _, err := range errs {
			if errors.Is(err, ErrRateLimited) {
				return nil, err
			}
		}
	}

	var heroImg, coverImg, logoImg image.Image
	m.Cover, coverImg = c.firstImage(ctx, urls.cover, Cover)
	m.Hero, heroImg = c.firstImage(ctx, urls.hero, Hero)
	// Failing screenshots, the middle of a large hero (Steam's 3840-wide
	// one) still fills the screen sharply; a small one is left to the hero.
	if len(urls.hero) > 0 {
		urls.backdrop = append(urls.backdrop, urls.hero[0])
	}
	m.Backdrop, _ = c.firstImage(ctx, urls.backdrop, Backdrop)
	m.Logo, logoImg = c.firstImage(ctx, urls.logo, Logo)
	m.Icon, _ = c.firstImage(ctx, urls.icon, Icon)
	if !found {
		// Only Steam's art files answered (a game its store no longer
		// lists): that's still worth keeping.
		if m.Cover == "" && m.Hero == "" {
			if len(errs) > 0 {
				return nil, errors.Join(errs...)
			}
			return nil, errNotFound
		}
		m.Source = "Steam"
	}
	if m.Cover == "" && heroImg != nil {
		m.Cover, _ = c.storeImage(cropCover(heroImg), Cover)
	}
	if t := makeTile(heroImg, logoImg, coverImg); t != nil {
		m.Tile, _ = c.storeImage(t, Tile)
	}
	switch {
	case coverImg != nil:
		m.Accent = Accent(coverImg)
	case heroImg != nil:
		m.Accent = Accent(heroImg)
	}
	if m.Accent == "" && heroImg != nil {
		m.Accent = Accent(heroImg)
	}
	keepOverrides(m, r.Keep)
	return m, nil
}

// firstImage stores the first candidate URL that yields a valid image.
func (c *Client) firstImage(ctx context.Context, urls []string, kind Kind) (string, image.Image) {
	for _, u := range urls {
		if art, img, err := c.saveImage(ctx, u, kind); err == nil {
			return art, img
		}
	}
	return "", nil
}

// keepOverrides carries over art the user picked themselves.
func keepOverrides(m, old *library.Meta) {
	if old == nil {
		return
	}
	for _, o := range old.ArtOverrides {
		switch o {
		case "cover":
			m.Cover = old.Cover
		case "hero":
			m.Hero = old.Hero
			if !slices.Contains(old.ArtOverrides, "backdrop") {
				m.Backdrop = "" // the user's hero, not a screenshot, behind the game
			}
			m.Tile = "" // made from the store's hero; the interface puts the user's together
		case "backdrop":
			m.Backdrop = old.Backdrop
		case "logo":
			m.Logo = old.Logo
			m.Tile = ""
		}
	}
	m.ArtOverrides = old.ArtOverrides
}

func firstParagraph(s string) string {
	if len(s) > 600 {
		if i := strings.LastIndex(s[:600], ". "); i > 200 {
			return s[:i+1]
		}
		return s[:600] + "…"
	}
	return s
}
