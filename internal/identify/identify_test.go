package identify

import (
	"os"
	"strings"
	"testing"

	"github.com/ApolloF/WaterLauncher/internal/scan"
)

const sample = `---
"Baldur's Gate 3":
  cloud:
    steam: true
  files:
    "<winLocalAppData>/Larian Studios/Baldur's Gate 3/PlayerProfiles":
      tags:
        - save
  installDir:
    Baldurs Gate 3: {}
  steam:
    id: 1086940
Hades:
  gog:
    id: 1330167364
  installDir:
    Hades: {}
  steam:
    id: 1145360
The Sims 4:
  installDir:
    The Sims 4: {}
  steam:
    id: 1222670
"-KLAUS-":
  alias: Klaus
Klaus:
  installDir:
    KLAUS: {}
  steam:
    id: 431330
"Ori and the Blind Forest: Definitive Edition":
  steam:
    id: 387290
Stardew Valley:
  installDir:
    Stardew Valley: {}
  steam:
    id: 413150
`

func index(t *testing.T) *Index {
	t.Helper()
	es, err := Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	return build(es)
}

func TestParse(t *testing.T) {
	es, err := Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 6 {
		t.Fatalf("got %d entries", len(es))
	}
	if es[0].Name != "Baldur's Gate 3" || es[0].SteamID != 1086940 || es[0].InstallDirs[0] != "Baldurs Gate 3" {
		t.Errorf("bg3 = %+v", es[0])
	}
	if es[1].GogID != "1330167364" {
		t.Errorf("hades = %+v", es[1])
	}
	for _, e := range es {
		if e.Name == "Klaus" && (len(e.Aliases) != 1 || e.Aliases[0] != "-KLAUS-") {
			t.Errorf("klaus aliases = %v", e.Aliases)
		}
	}
}

func TestIdentify(t *testing.T) {
	ix := index(t)
	for _, tc := range []struct {
		name  string
		c     scan.Candidate
		title string
		steam int
		conf  int
	}{
		{"emulator appid wins", scan.Candidate{Title: "Baldurs Gate 3", Source: scan.Installer, SteamAppID: 1086940, AppIDFrom: "steam_emu.ini"}, "Baldur's Gate 3", 1086940, 95},
		{"repack by title", scan.Candidate{Title: "The Sims 4", Source: scan.Installer, TitleTrusted: true}, "The Sims 4", 1222670, 85},
		{"alias", scan.Candidate{Title: "-KLAUS-", Source: scan.Folder}, "Klaus", 431330, 85},
		{"ampersand", scan.Candidate{Title: "Ori & the Blind Forest: Definitive Edition", Source: scan.Folder}, "Ori and the Blind Forest: Definitive Edition", 387290, 85},
		{"edition stripped", scan.Candidate{Title: "Hades Deluxe Edition", Source: scan.Folder}, "Hades", 1145360, 75},
		{"folder name", scan.Candidate{Title: "Kl4us Game", Dir: `D:\Games\KLAUS`, Source: scan.Folder}, "Klaus", 431330, 70},
		{"store title kept", scan.Candidate{Title: "Stardew Valley", Source: scan.Steam, SteamAppID: 413150}, "Stardew Valley", 413150, 100},
		{"unknown folder", scan.Candidate{Title: "Homebrew Thing", Dir: `D:\Games\Homebrew Thing`, Source: scan.Folder}, "Homebrew Thing", 0, 40},
	} {
		m := ix.Identify(tc.c)
		if m.Title != tc.title || m.SteamAppID != tc.steam || m.Confidence != tc.conf {
			t.Errorf("%s: got %+v", tc.name, m)
		}
	}
}

func TestNilIndex(t *testing.T) {
	var ix *Index
	m := ix.Identify(scan.Candidate{Title: "Baldurs Gate 3", SteamAppID: 1086940, AppIDFrom: "steam_emu.ini"})
	if m.Title != "Baldurs Gate 3" || m.SteamAppID != 1086940 {
		t.Errorf("got %+v", m)
	}
}

// TestRealManifest parses a downloaded manifest: WL_MANIFEST=path go test -run RealManifest ./internal/identify
func TestRealManifest(t *testing.T) {
	p := os.Getenv("WL_MANIFEST")
	if p == "" {
		t.Skip("set WL_MANIFEST to a manifest.yaml")
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	es, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	ix := build(es)
	t.Logf("%d entries, %d names, %d steam ids", len(es), ix.Len(), len(ix.bySteam))
	for _, c := range []scan.Candidate{
		{Title: "Baldurs Gate 3", SteamAppID: 1086940, AppIDFrom: "steam_emu.ini", Source: scan.Installer},
		{Title: "The Sims 4", Source: scan.Installer, TitleTrusted: true},
	} {
		t.Logf("%q → %+v", c.Title, ix.Identify(c))
	}
}
