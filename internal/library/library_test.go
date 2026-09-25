package library

import (
	"path/filepath"
	"testing"
	"time"
)

func TestApplyScanKeepsUserChoices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_800_000_000, 0)
	found := []Found{
		{Key: `c:\games\hades`, Title: "Hades", SortTitle: "hades", Source: "folder", Dir: `C:\Games\Hades`, Exe: `C:\Games\Hades\Hades.exe`, Confidence: 85},
		{Key: `c:\games\bg3`, Title: "Baldur's Gate 3", SortTitle: "baldur's gate 3", Source: "installer", SteamAppID: 1086940, Dir: `C:\Games\BG3`},
	}
	if added, _ := s.ApplyScan(found, now); added != 2 {
		t.Fatalf("added %d", added)
	}
	games := s.Games()
	if games[0].Title != "Baldur's Gate 3" || games[1].Title != "Hades" {
		t.Fatalf("order: %s, %s", games[0].Title, games[1].Title)
	}
	hades := games[1]
	if _, err := s.Update(hades.ID, func(g *Game) {
		g.Favorite, g.CustomTitle, g.PadMode, g.Playtime = true, "Hades (modded)", "steam", 3600
		g.UserExe, g.Exe = true, `C:\Games\Hades\x64\Hades.exe`
	}); err != nil {
		t.Fatal(err)
	}
	// Rescan: Hades is found again, BG3 is gone.
	found[0].Exe = `C:\Games\Hades\Other.exe`
	if _, removed := s.ApplyScan(found[:1], now.Add(time.Hour)); removed != 1 {
		t.Errorf("removed %d", removed)
	}
	g, _ := s.Get(hades.ID)
	if !g.Favorite || g.CustomTitle != "Hades (modded)" || g.PadMode != "steam" || g.Playtime != 3600 || g.Exe != `C:\Games\Hades\x64\Hades.exe` {
		t.Errorf("user choices lost: %+v", g)
	}
	if g.AddedAt != now.Unix() {
		t.Errorf("addedAt changed")
	}
	bg3, _ := s.Get(games[0].ID)
	if bg3.Installed {
		t.Error("removed game still installed")
	}

	// Round-trip through disk.
	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(s2.Games()) != 2 {
		t.Fatalf("reloaded %d games", len(s2.Games()))
	}
	if g2, _ := s2.Get(hades.ID); !g2.Favorite || g2.Playtime != 3600 {
		t.Errorf("reloaded = %+v", g2)
	}
	// New ids never reuse old ones.
	s2.ApplyScan([]Found{{Key: `c:\games\new`, Title: "New"}}, now)
	for _, g := range s2.Games() {
		if g.Title == "New" && g.ID <= bg3.ID {
			t.Errorf("id reused: %d", g.ID)
		}
	}
}

func TestBrokenFileStartsFresh(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.json")
	if err := writeAtomic(path, []byte("{not json")); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil || len(s.Games()) != 0 {
		t.Fatalf("got %v, %d games", err, len(s.Games()))
	}
}
