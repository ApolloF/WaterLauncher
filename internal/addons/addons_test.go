package addons

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// The test binary doubles as a fake add-on when WL_FAKE_ADDON is set.
func TestMain(m *testing.M) {
	if mode := os.Getenv("WL_FAKE_ADDON"); mode != "" {
		fakeAddon(mode)
		return
	}
	os.Exit(m.Run())
}

func fakeAddon(mode string) {
	if mode == "diesoon" {
		fmt.Fprintln(os.Stderr, "can't start")
		os.Exit(1)
	}
	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	write := func(v any) {
		b, _ := json.Marshal(v)
		out.Write(append(b, '\n'))
		out.Flush()
	}
	fmt.Fprintln(os.Stderr, "fake add-on here")
	for in.Scan() {
		var req struct {
			ID     *int64          `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if json.Unmarshal(in.Bytes(), &req) != nil {
			continue
		}
		switch req.Method {
		case "initialize":
			proto := 1
			if mode == "oldproto" {
				proto = 99
			}
			write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"protocol": proto, "name": "Fake", "version": "1.0"}})
		case "game.status":
			write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{
				"badges": []map[string]string{{"text": "DLSS 310.2", "tone": "ok"}},
			}})
		case "game.beforeLaunch":
			write(map[string]any{"jsonrpc": "2.0", "method": "progress", "params": map[string]string{"text": "patching"}})
			write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]string{"message": "done"}})
		case "game.runAction":
			write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": -32000, "message": "That didn't work"}})
		case "crash":
			os.Exit(3)
		case "hang":
			// never answers
		case "shutdown":
			return
		}
	}
}

func fakeManifest(t *testing.T, mode string) (*Manifest, string) {
	t.Helper()
	exe, _ := os.Executable()
	t.Setenv("WL_FAKE_ADDON", mode)
	m := &Manifest{ID: "fake", Name: "Fake", Publisher: "Tests", Exe: exe, Protocol: 1, Hooks: []string{HookStatus, HookBeforeLaunch}}
	sum, err := HashFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	return m, sum
}

type logs struct {
	mu    sync.Mutex
	lines []string
}

func (l *logs) f(format string, args ...any) {
	l.mu.Lock()
	l.lines = append(l.lines, fmt.Sprintf(format, args...))
	l.mu.Unlock()
}

func (l *logs) has(s string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, x := range l.lines {
		if strings.Contains(x, s) {
			return true
		}
	}
	return false
}

func TestHostCalls(t *testing.T) {
	m, sum := fakeManifest(t, "ok")
	lg := &logs{}
	h := NewHost("test", lg.f)
	defer h.StopAll()
	ctx := context.Background()

	var st struct {
		Badges []struct{ Text, Tone string } `json:"badges"`
	}
	if err := h.Call(ctx, m, sum, HookStatus, map[string]any{"game": map[string]any{"id": 1}}, &st, nil); err != nil {
		t.Fatal(err)
	}
	if len(st.Badges) != 1 || st.Badges[0].Text != "DLSS 310.2" {
		t.Errorf("status = %+v", st)
	}
	var progress []string
	var res struct{ Message string }
	if err := h.Call(ctx, m, sum, HookBeforeLaunch, nil, &res, func(s string) { progress = append(progress, s) }); err != nil {
		t.Fatal(err)
	}
	if res.Message != "done" || len(progress) != 1 || progress[0] != "patching" {
		t.Errorf("beforeLaunch = %+v, progress %v", res, progress)
	}
	err := h.Call(ctx, m, sum, "game.runAction", nil, nil, nil)
	if e, ok := err.(*Error); !ok || e.Message != "That didn't work" || e.Addon != "Fake" {
		t.Errorf("error = %#v", err)
	}
	if !h.Running("fake") {
		t.Error("should keep running between calls")
	}
	time.Sleep(200 * time.Millisecond)
	if !lg.has("fake add-on here") {
		t.Error("stderr not logged")
	}
}

func TestHostPinsHash(t *testing.T) {
	m, _ := fakeManifest(t, "ok")
	h := NewHost("test", func(string, ...any) {})
	if err := h.Call(context.Background(), m, "0000", HookStatus, nil, nil, nil); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Errorf("wrong hash = %v", err)
	}
}

func TestHostCrashAndTimeout(t *testing.T) {
	m, sum := fakeManifest(t, "ok")
	h := NewHost("test", func(string, ...any) {})
	defer h.StopAll()
	ctx := context.Background()
	// Crashes count toward the restart limit.
	for i := 0; i < maxRestarts; i++ {
		if err := h.Call(ctx, m, sum, "crash", nil, nil, nil); err == nil {
			t.Fatal("a crash must fail the call")
		}
		deadline := time.Now().Add(3 * time.Second)
		for h.Running("fake") && time.Now().Before(deadline) {
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := h.Call(ctx, m, sum, HookStatus, nil, nil, nil); err == nil || !strings.Contains(err.Error(), "stopped working") {
		t.Errorf("after %d crashes = %v", maxRestarts, err)
	}

	h2 := NewHost("test", func(string, ...any) {})
	defer h2.StopAll()
	tctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	if err := h2.Call(tctx, m, sum, "hang", nil, nil, nil); err == nil || !strings.Contains(err.Error(), "in time") {
		t.Errorf("hang = %v", err)
	}
}

func TestHostProtocolMismatch(t *testing.T) {
	m, sum := fakeManifest(t, "oldproto")
	h := NewHost("test", func(string, ...any) {})
	defer h.StopAll()
	if err := h.Call(context.Background(), m, sum, HookStatus, nil, nil, nil); err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Errorf("protocol mismatch = %v", err)
	}
}

func TestManifest(t *testing.T) {
	dir := t.TempDir()
	write := func(name, s string) string {
		d := filepath.Join(dir, name)
		os.MkdirAll(d, 0o755)
		p := filepath.Join(d, "addon.json")
		os.WriteFile(p, []byte(s), 0o644)
		return p
	}
	good := `{"id":"good","name":"Good","publisher":"Me","exe":"good.exe","protocol":1,"hooks":["game.status"],"permissions":["network"]}`
	write("a", good)
	write("b", `{"id":"Bad Id","name":"x","publisher":"y","exe":"x.exe","protocol":1,"hooks":["game.status"]}`)
	write("c", `{"id":"hooks","name":"x","publisher":"y","exe":"x.exe","protocol":1,"hooks":["game.format"]}`)
	write("d", `{"id":"perm","name":"x","publisher":"y","exe":"x.exe","protocol":1,"hooks":["game.status"],"permissions":["everything"]}`)
	write("e", `{"id":"script","name":"x","publisher":"y","exe":"x.bat","protocol":1,"hooks":["game.status"]}`)
	write("f", good) // same id twice
	write("g", `not json`)
	found, broken := Discover(dir)
	if len(found) != 1 || found[0].ID != "good" || found[0].ExePath() != filepath.Join(dir, "a", "good.exe") {
		t.Fatalf("found = %+v", found)
	}
	for _, k := range []string{"b", "c", "d", "e", "f", "g"} {
		if broken[k] == nil {
			t.Errorf("%s should be broken", k)
		}
	}

	// Install copies the manifest with an absolute program path.
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "good.exe"), []byte("MZ"), 0o644)
	os.WriteFile(filepath.Join(src, "addon.json"), []byte(good), 0o644)
	root := t.TempDir()
	m, err := Install(root, filepath.Join(src, "addon.json"))
	if err != nil {
		t.Fatal(err)
	}
	again, err := ReadManifest(filepath.Join(root, "good", "addon.json"))
	if err != nil || again.ExePath() != filepath.Join(src, "good.exe") || m.ID != "good" {
		t.Errorf("installed = %+v, %v", again, err)
	}
}

func TestTrust(t *testing.T) {
	p := filepath.Join(t.TempDir(), "trust.json")
	tr := OpenTrust(p)
	if tr.Get("x").Enabled {
		t.Fatal("new add-ons are off")
	}
	if err := tr.Set("x", Approval{Enabled: true, SHA256: "ab"}); err != nil {
		t.Fatal(err)
	}
	if a := OpenTrust(p).Get("x"); !a.Enabled || a.SHA256 != "ab" {
		t.Errorf("reloaded = %+v", a)
	}
}

func TestHostFailsFastWhenAddonDies(t *testing.T) {
	m, sum := fakeManifest(t, "diesoon")
	h := NewHost("test", func(string, ...any) {})
	defer h.StopAll()
	start := time.Now()
	err := h.Call(context.Background(), m, sum, HookStatus, nil, nil, nil)
	if err == nil {
		t.Fatal("must fail")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("took %v to notice the add-on died (%v)", d, err)
	}
}
