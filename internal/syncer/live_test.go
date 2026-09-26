package syncer

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLive talks to the Syncer on this PC when SYNCER_LIVE is set.
func TestLive(t *testing.T) {
	if os.Getenv("SYNCER_LIVE") == "" {
		t.Skip("SYNCER_LIVE not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	c, err := Dial(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	st, err := c.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status: %+v", st)
	fs, err := c.Games(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fs {
		t.Logf("  %-30s sync=%v backup=%v state=%s conflicts=%d backedUp=%s", f.Label, f.Sync, f.Backup, f.State, f.Conflicts, f.BackedUp.Format(time.DateTime))
	}
	if len(fs) > 0 {
		gs, err := c.GameStatus(ctx, Game{Title: fs[0].Label})
		if err != nil || !gs.Known {
			t.Errorf("gameStatus(%q) = %+v, %v", fs[0].Label, gs, err)
		}
		if f := fs[0]; f.Sync {
			r, err := c.SyncNow(ctx, []string{f.ID}, 15*time.Second)
			t.Logf("syncNow: %+v %v", r, err)
		}
	}
}
