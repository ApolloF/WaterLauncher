package scan

import (
	"path/filepath"
	"strconv"

	"github.com/ApolloF/gamekit/steam"
	"github.com/ApolloF/gamekit/vdf"
)

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
	for _, acc := range SteamAccounts(root) {
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

// SteamAccounts returns the userdata folder of the account Steam uses on
// this PC: the one signed in now, else the one that signs in
// automatically, else the last one used. All known accounts if unsure.
func SteamAccounts(root string) []string {
	if root == "" {
		return nil
	}
	return steam.Accounts(root, steam.HostInfo())
}
