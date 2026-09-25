package scan

import (
	"io/fs"
	"math"
	"path/filepath"
	"regexp"
	"strings"
)

// Executables that never start a game.
var reExcluded = regexp.MustCompile(`(?i)^(unins\d*|uninstall.*|setup.*|.*setup|dxsetup|dxwebsetup|vc_?redist.*|vcredist.*|dotnet.*|ndp\d+.*|oalinst|physx.*|.*crash.*|.*report.*|.*updater?|.*installer.*|.*prereq.*|.*redist.*|unitycrashhandler\d*|.*cefsubprocess|.*webhelper|shadercompileworker|.*helper|touchup|cleanup|activation.*|quicksfv|7za?|unrar|.*benchmark.*|.*_be|.*_eac|.*server|.*dedicated.*|dlc-.*|language-.*|.*-changer|easetup|eadm.*|register.*|.*patcher|.*config.*|.*settings|.*editor|.*sdk.*|.*modmanager|.*mod_manager|dowser|ue4prereqsetup.*|ueprereqsetup.*|crs-.*|start_protected_game)$`)

func excludedExe(p string) bool {
	return reExcluded.MatchString(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))
}

var reShipping = regexp.MustCompile(`(?i)-(win64|wingdk)-shipping$`)

var exeSkipDirs = map[string]bool{
	"_commonredist": true, "commonredist": true, "redist": true, "redistributables": true, "directx": true,
	"__installer": true, "installers": true, "installer": true, "vcredist": true, "dotnetfx": true, "physx": true,
	"$recycle.bin": true, "uninstall": true, "support": true, "engine": true, "tools": true, "_language packs": true,
	"dotnetcore": true, "prerequisites": true, "prereqs": true, "easyanticheat": true, "battleye": true,
}

// PickExe chooses the executable that most likely starts the game in dir.
// Unlike picking the process the game ends up running as, launching
// prefers Unreal's small bootstrapper in the root over the Shipping binary,
// since the bootstrapper passes the right arguments along.
func PickExe(dir, title string) string {
	root := filepath.Clean(dir)
	depth0 := strings.Count(root, `\`)
	type exe struct {
		path  string
		size  int64
		depth int
	}
	var exes []exe
	n := 0
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if n++; n > 20000 {
			return filepath.SkipAll
		}
		depth := strings.Count(filepath.Clean(p), `\`) - depth0
		if d.IsDir() {
			if p != root && (exeSkipDirs[strings.ToLower(d.Name())] || depth > 4) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(p), ".exe") || excludedExe(p) {
			return nil
		}
		if fi, err := d.Info(); err == nil {
			exes = append(exes, exe{p, fi.Size(), depth - 1})
		}
		return nil
	})
	names := []string{}
	for _, s := range []string{Normalize(filepath.Base(root)), Normalize(title)} {
		if len(s) >= 3 {
			names = append(names, s)
		}
	}
	best, bestScore := "", math.Inf(-1)
	for _, e := range exes {
		if s := exeScore(root, e.path, e.size, e.depth, names); s > bestScore {
			best, bestScore = e.path, s
		}
	}
	return best
}

func exeScore(root, path string, size int64, depth int, names []string) float64 {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	alnum := Normalize(base)
	lower := strings.ToLower(base)
	rel := strings.ToLower(strings.TrimPrefix(path, root))
	score := math.Log2(float64(size)/1048576+1)*6 - float64(depth)*4
	nameScore := 0.0
	for _, g := range names {
		switch {
		case alnum == g:
			nameScore = math.Max(nameScore, 40)
		case len(alnum) >= 4 && (strings.HasPrefix(g, alnum) || strings.HasPrefix(alnum, g)):
			nameScore = math.Max(nameScore, 20)
		}
	}
	score += nameScore
	dir := filepath.Dir(path)
	// Unity: the player sits next to <Name>_Data.
	if isDirPath(filepath.Join(dir, base+"_Data")) {
		score += 35
	}
	// Unreal: a root <Name>.exe next to <Name>\Binaries\Win64\<Name>-Win64-Shipping.exe.
	if depth == 0 && isDirPath(filepath.Join(dir, base, "Binaries")) {
		score += 60
	}
	if reShipping.MatchString(base) {
		score += 25
	}
	if strings.Contains(rel, `\binaries\win64\`) || strings.HasPrefix(rel, `\bin\x64\`) || strings.HasPrefix(rel, `\x64\`) || strings.HasPrefix(rel, `\bin64\`) {
		score += 12
	}
	if strings.Contains(lower, "launcher") {
		score -= 10
	}
	if strings.HasSuffix(lower, "32") || strings.Contains(rel, "x86") || strings.Contains(rel, "win32") {
		score -= 15
	}
	if strings.Contains(lower, "dx11") || strings.Contains(lower, "dx9") || strings.Contains(lower, "_fpb") || strings.Contains(lower, "vulkan") {
		score -= 3 // alternative renderer builds: keep the main one first
	}
	return score
}

func isDirPath(p string) bool { return isDir(p) }
