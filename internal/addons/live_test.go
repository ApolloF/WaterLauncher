package addons

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestLiveAddon runs a real add-on (ADDON_EXE, with ADDON_ARGS) against a
// game folder (ADDON_GAME), read-only: status, actions and beforeLaunch.
func TestLiveAddon(t *testing.T) {
	exe := os.Getenv("ADDON_EXE")
	if exe == "" {
		t.Skip("ADDON_EXE not set")
	}
	m := &Manifest{ID: "live", Name: "Live", Publisher: "Tests", Exe: exe, Args: []string{"--addon"}, Protocol: 1,
		Hooks: []string{HookStatus, HookActions, HookBeforeLaunch}}
	sum, err := HashFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	h := NewHost("test", t.Logf)
	defer h.StopAll()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	game := map[string]any{"game": map[string]any{"id": 1, "title": "Live game", "dir": os.Getenv("ADDON_GAME"), "source": "installer"}}
	for _, method := range []string{HookStatus, HookActions, HookBeforeLaunch} {
		var out json.RawMessage
		start := time.Now()
		err := h.Call(ctx, m, sum, method, game, &out, func(s string) { t.Logf("  progress: %s", s) })
		t.Logf("%s (%v): %s %v", method, time.Since(start).Round(time.Millisecond), out, err)
		if err != nil {
			t.Error(err)
		}
	}
}
