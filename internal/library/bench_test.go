package library

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// A big library: 500 games on this PC and 3,000 more owned on Steam.
func bigLibrary(b *testing.B) (*Store, []Found) {
	s, _ := Open(filepath.Join(b.TempDir(), "library.json"))
	var found []Found
	for i := 0; i < 500; i++ {
		t := fmt.Sprintf("Game %d", i)
		found = append(found, Found{Key: fmt.Sprintf(`c:\games\%d`, i), Title: t, SortTitle: t, Source: "steam", SteamAppID: 10 + i})
	}
	now := time.Now()
	s.ApplyScan(found, now)
	var owned []Owned
	for i := 0; i < 3000; i++ {
		owned = append(owned, Owned{Store: "steam", ID: strconv.Itoa(100000 + i), Title: fmt.Sprintf("Owned %d", i)})
	}
	s.ApplyOwned("steam", owned, now)
	return s, found
}

func BenchmarkApplyScanWithOwned(b *testing.B) {
	s, found := bigLibrary(b)
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ApplyScan(found, now)
	}
}

func BenchmarkGames(b *testing.B) {
	s, _ := bigLibrary(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Games()
	}
}
