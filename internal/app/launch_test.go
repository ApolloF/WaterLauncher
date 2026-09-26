package app

import (
	"testing"

	"github.com/ApolloF/WaterLauncher/internal/library"
	"github.com/ApolloF/WaterLauncher/internal/pad"
)

func TestRoute(t *testing.T) {
	ds := pad.State{Connected: true, Kind: pad.PlayStation, DualSense: true}
	xbox := pad.State{Connected: true, Kind: pad.Xbox}
	xinputOnly := &library.Meta{Controller: "full"}
	for _, tc := range []struct {
		name string
		g    library.Game
		pad  pad.State
		want string
	}{
		{"store game", library.Game{LaunchURI: "steam://rungameid/1", Meta: xinputOnly}, ds, RouteStore},
		{"xinput game, DualSense", library.Game{Exe: `C:\g.exe`, Meta: xinputOnly}, ds, RouteSteamInput},
		{"xinput game, Xbox pad", library.Game{Exe: `C:\g.exe`, Meta: xinputOnly}, xbox, RouteDirect},
		{"xinput game, no pad", library.Game{Exe: `C:\g.exe`, Meta: xinputOnly}, pad.State{}, RouteDirect},
		{"DualSense game", library.Game{Exe: `C:\g.exe`, Meta: &library.Meta{Controller: "full", DualSense: "yes"}}, ds, RouteDirect},
		{"ships libScePad", library.Game{Exe: `C:\g.exe`, PadHint: "libScePad", Meta: xinputOnly}, ds, RouteDirect},
		{"unknown support", library.Game{Exe: `C:\g.exe`}, ds, RouteDirect},
		{"no controller support", library.Game{Exe: `C:\g.exe`, Meta: &library.Meta{}}, ds, RouteDirect},
		{"forced native", library.Game{Exe: `C:\g.exe`, PadMode: "native", Meta: xinputOnly}, ds, RouteDirect},
		{"forced Steam Input", library.Game{Exe: `C:\g.exe`, PadMode: "steam"}, xbox, RouteSteamInput},
	} {
		if got := route(tc.g, tc.pad); got != tc.want {
			t.Errorf("%s: route = %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestPlayArg(t *testing.T) {
	if id, ok := PlayArg([]string{`C:\WaterLauncher.exe`, "--play", "42"}); !ok || id != 42 {
		t.Errorf("PlayArg = %d, %v", id, ok)
	}
	for _, args := range [][]string{nil, {"--play"}, {"--play", "x"}, {"--play", "-1"}} {
		if _, ok := PlayArg(args); ok {
			t.Errorf("PlayArg(%q) accepted", args)
		}
	}
}
