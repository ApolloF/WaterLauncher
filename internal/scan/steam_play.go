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

// steamID64 of account id 0; userdata folders are named by the 32-bit account id.
const steamIDBase = 76561197960265728

// SteamStat is what Steam recorded for one game.
type SteamStat struct {
	Playtime   int64 // seconds
	LastPlayed int64 // unix seconds
}

// steamStats reads playtime and last-played times from localconfig.vdf of
// the Steam account used on this PC.
func steamStats(root string) map[int]SteamStat {
	out := map[int]SteamStat{}
	if root == "" {
		return out
	}
	for _, acc := range steamAccounts(root) {
		n := vdf.ReadFile(filepath.Join(root, "userdata", acc, "config", "localconfig.vdf"))
		for _, store := range n.Kids() {
			for id, app := range store.Get("Software", "Valve", "Steam", "apps").Kids() {
				appID, err := strconv.Atoi(id)
				if err != nil || appID <= 0 {
					continue
				}
				minutes, _ := strconv.ParseInt(app.Value("Playtime"), 10, 64)
				last, _ := strconv.ParseInt(app.Value("LastPlayed"), 10, 64)
				s := out[appID]
				s.Playtime = max(s.Playtime, minutes*60)
				s.LastPlayed = max(s.LastPlayed, last)
				out[appID] = s
			}
		}
	}
	return out
}

// steamAccounts returns the userdata folder of the account Steam uses on
// this PC: the one signed in now, else the one that signs in
// automatically, else the last one used. All known accounts if unsure.
func steamAccounts(root string) []string {
	type user struct {
		acc, name string
		recent    bool
		ts        int64
	}
	var users []user
	for id64, u := range vdf.ReadFile(filepath.Join(root, "config", "loginusers.vdf")).Get("users").Kids() {
		n, err := strconv.ParseUint(id64, 10, 64)
		if err != nil || n <= steamIDBase {
			continue
		}
		acc := strconv.FormatUint(n-steamIDBase, 10)
		if !platform.IsDir(filepath.Join(root, "userdata", acc)) {
			continue
		}
		ts, _ := strconv.ParseInt(u.Value("Timestamp"), 10, 64)
		users = append(users, user{acc, u.Value("AccountName"), u.Value("MostRecent") == "1", ts})
	}
	sort.Slice(users, func(i, j int) bool { return users[i].acc < users[j].acc })

	if k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam\ActiveProcess`, registry.QUERY_VALUE); err == nil {
		v, _, err := k.GetIntegerValue("ActiveUser")
		k.Close()
		if acc := strconv.FormatUint(v, 10); err == nil && v != 0 && platform.IsDir(filepath.Join(root, "userdata", acc)) {
			return []string{acc}
		}
	}
	if k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE); err == nil {
		name, _, _ := k.GetStringValue("AutoLoginUser")
		k.Close()
		for _, u := range users {
			if name != "" && strings.EqualFold(u.name, name) {
				return []string{u.acc}
			}
		}
	}
	best := -1
	for i, u := range users {
		if u.recent {
			return []string{u.acc}
		}
		if u.ts > 0 && (best < 0 || u.ts > users[best].ts) {
			best = i
		}
	}
	if best >= 0 {
		return []string{users[best].acc}
	}
	var all []string
	es, _ := os.ReadDir(filepath.Join(root, "userdata"))
	for _, e := range es {
		if _, err := strconv.Atoi(e.Name()); err == nil && e.IsDir() && e.Name() != "0" {
			all = append(all, e.Name())
		}
	}
	return all
}
