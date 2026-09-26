package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ApolloF/WaterLauncher/internal/identify"
	"github.com/ApolloF/WaterLauncher/internal/meta"
	"github.com/ApolloF/WaterLauncher/internal/platform"
	"github.com/ApolloF/WaterLauncher/internal/scan"
)

// TestMatchAudit runs well-known games, named the way stores, installers
// and folders name them, through identification and the Steam store
// search, with the real game database, and lists what finds no Steam app
// (no metadata or art without a SteamGridDB key):
//
//	WL_MATCH_AUDIT=1 go test -run MatchAudit -v ./internal/app
func TestMatchAudit(t *testing.T) {
	if os.Getenv("WL_MATCH_AUDIT") == "" {
		t.Skip("set WL_MATCH_AUDIT=1 to query the Steam store")
	}
	ix := identify.NewManager(platform.CacheDir("manifest")).Index()
	if ix == nil {
		t.Skip("no game database in the cache yet (run WaterLauncher once)")
	}
	w := &metaWorker{client: meta.NewClient(t.TempDir(), nil)}
	store := func(title string, src scan.Source) scan.Candidate {
		return scan.Candidate{Title: title, Source: src, How: "store"}
	}
	folder := func(dir string) scan.Candidate {
		return scan.Candidate{Title: scan.CleanTitle(filepath.Base(dir)), Dir: dir, Source: scan.Folder}
	}
	installer := func(name, dir string) scan.Candidate {
		return scan.Candidate{Title: scan.CleanTitle(name), Dir: dir, Source: scan.Installer, TitleTrusted: true}
	}
	cases := []scan.Candidate{
		store("Grand Theft Auto V", scan.Epic),
		store("Red Dead Redemption 2", scan.Epic),
		store("Cyberpunk 2077", scan.GOG),
		store("The Witcher 3: Wild Hunt - Complete Edition", scan.GOG),
		store("Hogwarts Legacy", scan.Epic),
		store("Alan Wake 2", scan.Epic),
		store("Control Ultimate Edition", scan.Epic),
		store("Death Stranding Director's Cut", scan.Epic),
		store("Marvel's Spider-Man Remastered", scan.Epic),
		store("Horizon Zero Dawn™ Complete Edition", scan.Epic),
		store("Borderlands 3", scan.Epic),
		store("Rocket League®", scan.Epic),
		store("Fall Guys", scan.Epic),
		store("Hades", scan.Epic),
		store("Celeste", scan.Epic),
		store("Assassin's Creed Valhalla", scan.Ubisoft),
		store("Far Cry® 6", scan.Ubisoft),
		store("Tom Clancy's Rainbow Six® Siege", scan.Ubisoft),
		store("Forza Horizon 5", scan.Xbox),
		store("Microsoft Flight Simulator", scan.Xbox),
		store("Halo: The Master Chief Collection", scan.Xbox),
		store("Starfield", scan.Xbox),
		store("Sea of Thieves", scan.Xbox),
		store("EA SPORTS FC™ 24", scan.EA),
		store("Battlefield™ 2042", scan.EA),
		store("The Sims™ 4", scan.EA),
		store("Apex Legends™", scan.EA),
		store("Mass Effect™ Legendary Edition", scan.EA),
		store("Diablo IV", scan.BattleNet),
		store("Overwatch 2", scan.BattleNet),
		store("Call of Duty®", scan.BattleNet),
		folder(`D:\Games\Elden.Ring.v1.10-FitGirl`),
		folder(`D:\Games\Red Dead Redemption 2 [FitGirl Repack]`),
		folder(`D:\Games\Cyberpunk.2077.Phantom.Liberty-RUNE`),
		folder(`D:\Games\Hogwarts.Legacy.Deluxe.Edition-EMPRESS`),
		folder(`D:\Games\The Witcher 3 Wild Hunt GOTY`),
		folder(`D:\Games\GTA V`),
		folder(`D:\Games\Sekiro Shadows Die Twice`),
		folder(`D:\Games\DARK SOULS III`),
		folder(`D:\Games\Spider-Man Miles Morales`),
		folder(`D:\Games\God of War`),
		folder(`D:\Games\Ghost of Tsushima DIRECTOR'S CUT`),
		folder(`D:\Games\HITMAN 3`),
		folder(`D:\Games\Resident Evil 4`),
		folder(`D:\Games\Batman Arkham Knight`),
		folder(`D:\Games\Star Wars Jedi Fallen Order`),
		folder(`D:\Games\Mortal Kombat 11 Ultimate`),
		folder(`D:\Games\TEKKEN 8`),
		folder(`D:\Games\Street Fighter 6`),
		folder(`D:\Games\Monster Hunter World`),
		folder(`D:\Games\Palworld`),
		folder(`D:\Games\Lethal Company`),
		folder(`D:\Games\Black Myth Wukong`),
		folder(`D:\Games\STALKER 2 Heart of Chornobyl`),
		folder(`D:\Games\Kingdom Come Deliverance II`),
		folder(`D:\Games\DOOM Eternal`),
		folder(`D:\Games\Metro Exodus Enhanced Edition`),
		folder(`D:\Games\Horizon Forbidden West Complete Edition`),
		folder(`D:\Games\Final Fantasy VII Remake Intergrade`),
		folder(`D:\Games\Persona 5 Royal`),
		folder(`D:\Games\Yakuza Like a Dragon`),
		folder(`D:\Games\Sonic Frontiers`),
		folder(`D:\Games\Need for Speed Heat`),
		folder(`D:\Games\Assassins Creed Mirage`),
		folder(`D:\Games\FarCry5`),
		folder(`D:\Games\Baldurs.Gate.3.v4.1.1-GOG`),
		folder(`D:\Games\Stardew_Valley`),
		folder(`D:\Games\HollowKnight`),
		folder(`D:\Games\Hades II`),
		folder(`D:\Games\Portal 2`),
		installer("Elden Ring Shadow of the Erdtree Deluxe Edition", `C:\Games\ELDEN RING`),
		installer("Grand Theft Auto V version 1.0.3179", `C:\Games\Grand Theft Auto V`),
		installer("Cyberpunk 2077 - Ultimate Edition", `C:\Games\Cyberpunk 2077`),
		installer("Resident Evil Village Gold Edition", `C:\Games\RE Village`),
		installer("Dying Light 2 Stay Human", `C:\Games\Dying Light 2`),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	failed := 0
	for _, c := range cases {
		m := ix.Identify(c)
		how, app, name := m.How, m.SteamAppID, m.Title
		if app == 0 && m.GogID == "" {
			h, ok, err := w.searchStore(ctx, m.Title)
			if err != nil {
				t.Logf("search %q: %v", m.Title, err)
			}
			if !ok {
				t.Logf("MISS %-45q → %q (%s)", c.Title, m.Title, m.How)
				failed++
				continue
			}
			how, app, name = "store search", h.AppID, h.Name
		}
		t.Logf("ok   %-45q → %q app %d (%s)", c.Title, name, app, how)
	}
	t.Logf("%d of %d without a Steam app", failed, len(cases))
}
