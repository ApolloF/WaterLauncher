package lnk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArgsAndWorkDir(t *testing.T) {
	l, err := ReadFile(filepath.Join("testdata", "args.lnk"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(filepath.Base(l.Target), "notepad.exe") {
		t.Errorf("target = %q", l.Target)
	}
	if l.Args != `-windowed --profile "My Profile"` {
		t.Errorf("args = %q", l.Args)
	}
	if !strings.EqualFold(l.WorkDir, os.Getenv("WINDIR")) {
		t.Errorf("workdir = %q", l.WorkDir)
	}
	if l.Name != "Play Notepad" {
		t.Errorf("name = %q", l.Name)
	}
}

func TestUnicodeTarget(t *testing.T) {
	l, err := ReadFile(filepath.Join("testdata", "unicode.lnk"))
	if err != nil {
		t.Fatal(err)
	}
	want := `C:\WLFixture\Café Ünïcode Game\bin\game.exe`
	if l.Target != want {
		t.Errorf("target = %q, want %q", l.Target, want)
	}
	if l.WorkDir != `C:\WLFixture\Café Ünïcode Game\bin` {
		t.Errorf("workdir = %q", l.WorkDir)
	}
}

func TestRejectsGarbage(t *testing.T) {
	for _, b := range [][]byte{nil, []byte("hello"), make([]byte, 200)} {
		if _, err := Parse(b); err == nil {
			t.Errorf("Parse(%d bytes) accepted garbage", len(b))
		}
	}
}

// Truncated links must fail cleanly, never panic.
func TestTruncated(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "unicode.lnk"))
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < len(b); n += 7 {
		_, _ = Parse(b[:n])
	}
}

func FuzzParse(f *testing.F) {
	for _, name := range []string{"args.lnk", "unicode.lnk"} {
		if b, err := os.ReadFile(filepath.Join("testdata", name)); err == nil {
			f.Add(b)
		}
	}
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = Parse(b) })
}
