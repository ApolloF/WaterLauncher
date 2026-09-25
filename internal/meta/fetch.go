package meta

import (
	"context"
	"errors"
	"image"
	"strings"
	"time"

	"github.com/ApolloF/WaterLauncher/internal/library"
)

// Request says which game to fetch metadata for.
type Request struct {
	Title      string
	SteamAppID int
	GogID      string
	Keep       *library.Meta // previous metadata: user-chosen art is kept
}

// Fetch gathers metadata and art for one game. Missing pieces are simply
// left empty; an error means nothing at all could be fetched (and is
// ErrRateLimited when the caller should back off).
func (c *Client) Fetch(ctx context.Context, r Request) (*library.Meta, error) {
	m := &library.Meta{FetchedAt: time.Now().Unix()}
	var urls struct{ cover, hero, logo, icon []string }
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
		for _, err := range errs {
			if errors.Is(err, ErrRateLimited) {
				return nil, err
			}
		}
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
		return nil, errNotFound
	}

	var heroImg, coverImg image.Image
	m.Cover, coverImg = c.firstImage(ctx, urls.cover, Cover)
	m.Hero, heroImg = c.firstImage(ctx, urls.hero, Hero)
	m.Logo, _ = c.firstImage(ctx, urls.logo, Logo)
	m.Icon, _ = c.firstImage(ctx, urls.icon, Icon)
	if m.Cover == "" && heroImg != nil {
		m.Cover, _ = c.storeImage(cropCover(heroImg), Cover)
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
		case "logo":
			m.Logo = old.Logo
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
