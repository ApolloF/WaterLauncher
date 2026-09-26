package meta

import (
	"context"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAllowed(t *testing.T) {
	for raw, want := range map[string]bool{
		"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/1/logo.png": true,
		"https://images-2.gog-statics.com/x.jpg":                                        true,
		"https://cdn2.steamgriddb.com/grid/x.png":                                       true,
		"http://shared.akamai.steamstatic.com/x.png":                                    false, // not https
		"https://evil.example.com/x.png":                                                false,
		"https://gog-statics.com.evil.com/x.png":                                        false,
		"https://user:pw@api.gog.com/products/1":                                        false,
		"file:///C:/Windows/win.ini":                                                    false,
	} {
		u, _ := url.Parse(raw)
		if got := allowed(u); got != want {
			t.Errorf("allowed(%s) = %v, want %v", raw, got, want)
		}
	}
}

func solid(w, h int, c color.Color) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestAccent(t *testing.T) {
	if a := Accent(solid(64, 64, color.RGBA{200, 40, 40, 255})); a == "" || !strings.HasPrefix(a, "#e") {
		t.Errorf("red image accent = %q", a)
	}
	if a := Accent(solid(64, 64, color.RGBA{128, 128, 128, 255})); a != "" {
		t.Errorf("grey image accent = %q, want none", a)
	}
}

func TestCropCoverAndFit(t *testing.T) {
	c := cropCover(solid(1920, 620, color.White))
	if b := c.Bounds(); b.Dy() != 620 || b.Dx() != 413 {
		t.Errorf("crop = %v", b)
	}
	if b := fit(solid(3000, 900, color.White), 1920).Bounds(); b.Dx() != 1920 || b.Dy() != 576 {
		t.Errorf("fit = %v", b)
	}
}

func TestPlainText(t *testing.T) {
	if got := plainText(`<p>Hello &amp; <b>welcome</b></p><br>to the  game`); got != "Hello & welcome to the game" {
		t.Errorf("plainText = %q", got)
	}
	if yearOf("Aug 3, 2023") != 2023 || yearOf("Coming soon") != 0 {
		t.Error("yearOf")
	}
}

func TestArtHandler(t *testing.T) {
	dir := t.TempDir()
	name := strings.Repeat("a", 40) + ".jpg"
	if err := os.WriteFile(dir+`\`+name, []byte("jpegdata"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := ArtHandler(dir)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(299) }))
	for path, want := range map[string]int{
		"/art/" + name:             200,
		"/art/../secrets/sgdb.bin": 404,
		"/art/..%5Clibrary.json":   404,
		"/art/" + name + "x":       404,
		"/index.html":              299,
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "http://wails.localhost"+path, nil))
		if rec.Code != want {
			t.Errorf("%s: %d, want %d", path, rec.Code, want)
		}
	}
}

// TestRealFetch fetches Baldur's Gate 3 and Stardew Valley from Steam:
// WL_REAL_META=1 go test -run RealFetch -v ./internal/meta
func TestRealFetch(t *testing.T) {
	if os.Getenv("WL_REAL_META") == "" {
		t.Skip("set WL_REAL_META=1 to fetch from Steam")
	}
	c := NewClient(t.TempDir(), nil)
	for _, id := range []int{1086940, 413150} {
		m, err := c.Fetch(context.Background(), Request{SteamAppID: id})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%d: cover=%s hero=%s logo=%s accent=%s dualsense=%q year=%d genres=%v dev=%v\n  %s", id, m.Cover, m.Hero, m.Logo, m.Accent, m.DualSense, m.ReleaseYear, m.Genres, m.Developers, m.Description)
		if m.Cover == "" || m.Hero == "" {
			t.Errorf("%d: missing art", id)
		}
	}
}

func TestPruneArt(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-48 * time.Hour)
	files := map[string]bool{
		"1111111111111111111111111111111111111111.jpg":     true,  // in use
		"2222222222222222222222222222222222222222.png":     false, // unused
		"3333333333333333333333333333333333333333.jpg.tmp": false,
		"notes.txt": true, // not ours
	}
	for name := range files {
		p := filepath.Join(dir, name)
		_ = os.WriteFile(p, []byte("x"), 0o644)
		_ = os.Chtimes(p, old, old)
	}
	fresh := filepath.Join(dir, "4444444444444444444444444444444444444444.jpg")
	_ = os.WriteFile(fresh, []byte("x"), 0o644) // just written: kept
	n, _ := PruneArt(dir, map[string]bool{"/art/1111111111111111111111111111111111111111.jpg": true}, 24*time.Hour)
	if n != 2 {
		t.Errorf("removed %d, want 2", n)
	}
	for name, stays := range files {
		if _, err := os.Stat(filepath.Join(dir, name)); (err == nil) != stays {
			t.Errorf("%s: exists=%v, want %v", name, err == nil, stays)
		}
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a fresh file was removed")
	}
}
