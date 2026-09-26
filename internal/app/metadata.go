package app

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/ApolloF/WaterLauncher/internal/library"
	"github.com/ApolloF/WaterLauncher/internal/logx"
	"github.com/ApolloF/WaterLauncher/internal/meta"
	"github.com/ApolloF/WaterLauncher/internal/platform"
	"github.com/ApolloF/WaterLauncher/internal/scan"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// EventMetaState reports metadata fetching progress to the interface.
const EventMetaState = "meta:state"

// MetaState is what the metadata worker is doing.
type MetaState struct {
	Running bool `json:"running"`
	Done    int  `json:"done"`
	Total   int  `json:"total"`
}

func init() { application.RegisterEvent[MetaState](EventMetaState) }

const (
	sgdbSecret   = "steamgriddb"
	metaMaxAge   = 30 * 24 * time.Hour
	metaRetryGap = 24 * time.Hour
)

// metaWorker fetches metadata and art in the background, one game at a
// time, most recently played first.
type metaWorker struct {
	c      *Core
	client *meta.Client

	mu      sync.Mutex
	queue   []int64
	queued  map[int64]bool
	failed  map[int64]time.Time
	state   MetaState
	wake    chan struct{}
	changed *time.Timer
}

func newMetaWorker(c *Core) *metaWorker {
	w := &metaWorker{c: c, queued: map[int64]bool{}, failed: map[int64]time.Time{}, wake: make(chan struct{}, 1)}
	w.client = meta.NewClient(platform.CacheDir("art"), func() string { return platform.LoadSecret(sgdbSecret) })
	return w
}

// queueMissing queues installed games without (fresh) metadata, and
// owned games that aren't installed when those are shown.
func (w *metaWorker) queueMissing() {
	hasKey := platform.LoadSecret(sgdbSecret) != ""
	showOwned := w.c.Settings.Get().ShowOwned
	games := w.c.Lib.Games()
	sort.SliceStable(games, func(i, j int) bool {
		if games[i].Installed != games[j].Installed {
			return games[i].Installed // installed games first
		}
		a := max(games[i].LastPlayed, games[i].StoreLastPlayed)
		b := max(games[j].LastPlayed, games[j].StoreLastPlayed)
		return a > b
	})
	now := time.Now()
	w.mu.Lock()
	for _, g := range games {
		if !(g.Installed || showOwned && g.Owned) || g.Hidden || w.queued[g.ID] {
			continue
		}
		if t, ok := w.failed[g.ID]; ok && now.Sub(t) < metaRetryGap {
			continue
		}
		fresh := g.Meta != nil && now.Sub(time.Unix(g.Meta.FetchedAt, 0)) < metaMaxAge
		if fresh && (g.Meta.Cover != "" || !hasKey) {
			continue
		}
		w.queue = append(w.queue, g.ID)
		w.queued[g.ID] = true
		w.state.Total++
	}
	w.mu.Unlock()
	w.poke()
}

// queueNow puts one game at the front of the queue, even if it has metadata.
func (w *metaWorker) queueNow(id int64) {
	w.mu.Lock()
	delete(w.failed, id)
	if !w.queued[id] {
		w.queue = append([]int64{id}, w.queue...)
		w.queued[id] = true
		w.state.Total++
	}
	w.mu.Unlock()
	w.poke()
}

func (w *metaWorker) poke() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

func (w *metaWorker) run(ctx context.Context) {
	for {
		id, ok := w.next()
		if !ok {
			w.setState(func(s *MetaState) { *s = MetaState{} })
			select {
			case <-ctx.Done():
				return
			case <-w.wake:
				continue
			}
		}
		w.setState(func(s *MetaState) { s.Running = true })
		err := w.fetch(ctx, id)
		if ctx.Err() != nil {
			return
		}
		if errors.Is(err, meta.ErrRateLimited) {
			logx.Printf("metadata: rate limited, pausing a minute")
			w.mu.Lock()
			w.queue = append([]int64{id}, w.queue...)
			w.queued[id] = true
			w.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
			}
			continue
		}
		w.setState(func(s *MetaState) { s.Done++ })
	}
}

func (w *metaWorker) next() (int64, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.queue) == 0 {
		return 0, false
	}
	id := w.queue[0]
	w.queue = w.queue[1:]
	delete(w.queued, id)
	return id, true
}

func (w *metaWorker) fetch(ctx context.Context, id int64) error {
	g, ok := w.c.Lib.Get(id)
	if !ok {
		return nil
	}
	appID := g.SteamAppID
	if appID == 0 {
		appID = g.MetaAppID
	}
	// Unknown to the game database: an exact title match on the Steam store
	// still gives art and details.
	if appID == 0 && g.GogID == "" {
		if hits, err := w.client.SearchSteam(ctx, g.DisplayTitle()); err == nil {
			want := scan.Normalize(g.DisplayTitle())
			for _, h := range hits {
				if scan.Normalize(h.Name) == want {
					appID = h.AppID
					_, _ = w.c.Lib.Update(id, func(g *library.Game) { g.MetaAppID = h.AppID })
					break
				}
			}
		} else if errors.Is(err, meta.ErrRateLimited) {
			return err
		}
	}
	m, err := w.client.Fetch(ctx, meta.Request{Title: g.DisplayTitle(), SteamAppID: appID, GogID: g.GogID, Keep: g.Meta})
	if err != nil {
		if !errors.Is(err, meta.ErrRateLimited) {
			w.mu.Lock()
			w.failed[id] = time.Now()
			w.mu.Unlock()
			logx.Printf("metadata for %q: %v", g.DisplayTitle(), err)
		}
		return err
	}
	_, err = w.c.Lib.Update(id, func(g *library.Game) { g.Meta = m })
	w.libraryChanged()
	return err
}

// libraryChanged tells the interface to reload, at most a few times a second.
func (w *metaWorker) libraryChanged() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.changed != nil {
		return
	}
	w.changed = time.AfterFunc(600*time.Millisecond, func() {
		w.mu.Lock()
		w.changed = nil
		w.mu.Unlock()
		w.c.emit(EventLibraryChanged, "metadata")
	})
}

func (w *metaWorker) setState(fn func(*MetaState)) {
	w.mu.Lock()
	fn(&w.state)
	if !w.state.Running && len(w.queue) == 0 {
		w.state = MetaState{}
	}
	s := w.state
	w.mu.Unlock()
	w.c.emit(EventMetaState, s)
}

func (w *metaWorker) State() MetaState {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state
}
