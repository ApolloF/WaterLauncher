package vdf

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	n := Parse(strings.NewReader(`
"libraryfolders"
{
	// comment
	"0"
	{
		"path"		"C:\\Program Files (x86)\\Steam"
		"apps" { "413150" "123" }
	}
	"1" { "path" "D:\\SteamLibrary" }
}`))
	lf := n.Get("LibraryFolders")
	if lf == nil || len(lf.Kids()) != 2 {
		t.Fatalf("libraryfolders = %+v", lf)
	}
	if got := lf.Get("0").Value("PATH"); got != `C:\Program Files (x86)\Steam` {
		t.Errorf("path = %q", got)
	}
	if got := lf.Get("0", "apps").Value("413150"); got != "123" {
		t.Errorf("apps = %q", got)
	}
	if n.Get("missing", "deeper") != nil || n.Get("missing").Value("x") != "" {
		t.Error("missing keys must be nil-safe")
	}
}

func TestMalformed(t *testing.T) {
	for _, s := range []string{`"a" {`, `}}}}`, `"unterminated`, `{"k" "v"`} {
		_ = Parse(strings.NewReader(s))
	}
}

func FuzzParse(f *testing.F) {
	f.Add(`"AppState" { "appid" "1" "name" "x" }`)
	f.Fuzz(func(t *testing.T, s string) { _ = Parse(strings.NewReader(s)) })
}
