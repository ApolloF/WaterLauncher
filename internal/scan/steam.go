package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ApolloF/WaterLauncher/internal/platform"
	"github.com/ApolloF/WaterLauncher/internal/vdf"
	"golang.org/x/sys/windows/registry"
)

// steamTools are app ids Steam installs that aren't games.
var steamTools = map[int]bool{
	228980:  true, // Steamworks Common Redistributables
	250820:  true, // SteamVR
	431960:  true, // Wallpaper Engine
	1070560: true, // Steam Linux Runtime
	1391110: true, // Steam Linux Runtime - Soldier
	1628350: true, // Steam Linux Runtime - Sniper
	1826330: true, // Proton EasyAntiCheat Runtime
	1493710: true, // Proton Experimental
	1161040: true, // Proton BattlEye Runtime
}

var steamToolNames = []string{"proton", "steam linux runtime", "steamworks", "redistributable", "steamvr"}

// State flags of an appmanifest (EAppState).
const (
	stateUpdateRequired = 2
	stateFullyInstalled = 4
	stateUpdateRunning  = 256
	stateUpdatePaused   = 512
	stateUpdateStarted  = 1024
	stateUninstalling   = 2048
)

// SteamDir returns Steam's install folder ("" if Steam isn't installed).
func SteamDir() string {
	for _, k := range []struct {
		root       registry.Key
		path, name string
	}{
		{registry.CURRENT_USER, `Software\Valve\Steam`, "SteamPath"},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, "InstallPath"},
		{registry.LOCAL_MACHINE, `SOFTWARE\Valve\Steam`, "InstallPath"},
	} {
		key, err := registry.OpenKey(k.root, k.path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		v, _, err := key.GetStringValue(k.name)
		key.Close()
		if err != nil || strings.TrimSpace(v) == "" {
			continue
		}
		if d := filepath.Clean(filepath.FromSlash(strings.TrimSpace(v))); platform.IsDir(filepath.Join(d, "steamapps")) {
			return platform.RealCase(d)
		}
	}
	if d := filepath.Join(platform.ProgramFilesX86, "Steam"); platform.IsDir(filepath.Join(d, "steamapps")) {
		return d
	}
	return ""
}

// SteamLibraries returns Steam's own folder plus every library folder listed
// in its libraryfolders.vdf.
func SteamLibraries(root string) []string {
	if !filepath.IsAbs(root) {
		return nil
	}
	libs := []string{filepath.Clean(root)}
	seen := map[string]bool{platform.Key(root): true}
	var extra []string
	for _, lib := range vdf.ReadFile(filepath.Join(root, "steamapps", "libraryfolders.vdf")).Get("libraryfolders").Kids() {
		dir := filepath.Clean(lib.Value("path"))
		if k := platform.Key(dir); filepath.IsAbs(dir) && !seen[k] {
			extra = append(extra, dir)
			seen[k] = true
		}
	}
	sort.Strings(extra)
	return append(libs, extra...)
}

// steamCandidates lists the games Steam installed, across all libraries.
func steamCandidates(root string) []Candidate {
	var out []Candidate
	seen := map[int]bool{}
	stats := steamStats(root)
	for _, lib := range SteamLibraries(root) {
		files, _ := filepath.Glob(filepath.Join(lib, "steamapps", "appmanifest_*.acf"))
		for _, file := range files {
			st := vdf.ReadFile(file).Get("AppState")
			id, err := strconv.Atoi(st.Value("appid"))
			if err != nil || id <= 0 || seen[id] || steamTools[id] {
				continue
			}
			name := st.Value("name")
			if isSteamTool(name) {
				continue
			}
			common := filepath.Join(lib, "steamapps", "common")
			installDir := st.Value("installdir")
			if !filepath.IsLocal(installDir) {
				continue
			}
			dir := filepath.Join(common, installDir)
			flags, _ := strconv.Atoi(st.Value("StateFlags"))
			playable := flags&stateFullyInstalled != 0 ||
				flags&(stateUpdateRequired|stateUpdateRunning|stateUpdatePaused|stateUpdateStarted) != 0
			if !playable || flags&stateUninstalling != 0 || !platform.IsDir(dir) {
				continue
			}
			size, _ := strconv.ParseInt(st.Value("SizeOnDisk"), 10, 64)
			if name == "" {
				name = installDir
			}
			seen[id] = true
			out = append(out, Candidate{
				Title: name, TitleTrusted: true, Dir: dir, Source: Steam, SteamAppID: id,
				LaunchURI: "steam://rungameid/" + strconv.Itoa(id), How: "Steam library", SizeBytes: size,
				StorePlaytime: stats[id].Playtime, StoreLastPlayed: stats[id].LastPlayed,
			})
		}
	}
	return out
}

func isSteamTool(name string) bool {
	l := strings.ToLower(name)
	for _, t := range steamToolNames {
		if strings.Contains(l, t) {
			return true
		}
	}
	return false
}

// SteamWatchDirs are the folders whose changes mean Steam installed or removed something.
func SteamWatchDirs() []string {
	var out []string
	for _, lib := range SteamLibraries(SteamDir()) {
		if d := filepath.Join(lib, "steamapps"); platform.IsDir(d) {
			out = append(out, d)
		}
	}
	return out
}

// dirSize totals file sizes below dir, stopping after limit entries.
func dirSize(dir string, limit int) int64 {
	var total int64
	n := 0
	_ = filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if n++; n > limit {
			return filepath.SkipAll
		}
		if !d.IsDir() {
			if fi, err := d.Info(); err == nil {
				total += fi.Size()
			}
		}
		return nil
	})
	return total
}
