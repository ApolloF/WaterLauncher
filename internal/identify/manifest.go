// Package identify works out which game a scanned folder is: its proper
// title and Steam or GOG id, from the ids a copy carries or, failing that,
// its name, using the community Ludusavi manifest (PCGamingWiki data on
// ~50,000 games, with their Steam and GOG ids and install folder names).
package identify

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ApolloF/WaterLauncher/internal/platform"
)

const (
	manifestURL  = "https://raw.githubusercontent.com/mtkennerly/ludusavi-manifest/master/data/manifest.yaml"
	refreshAfter = 7 * 24 * time.Hour
	indexName    = "manifest-index-v1.gob.gz"
	maxManifest  = 256 << 20
)

// Entry is one game in the manifest.
type Entry struct {
	Name        string
	SteamID     int
	GogID       string
	InstallDirs []string
	Aliases     []string // other names that point here
}

// Manager loads the manifest index from cache and refreshes it weekly.
type Manager struct {
	dir  string
	mu   sync.Mutex
	idx  *Index
	busy bool
}

// NewManager uses dir for the cached index (created on demand).
func NewManager(dir string) *Manager { return &Manager{dir: dir} }

func (m *Manager) indexFile() string { return filepath.Join(m.dir, indexName) }
func (m *Manager) etagFile() string  { return filepath.Join(m.dir, "manifest.etag") }

// Index returns the loaded index, reading the cache on first use; nil until one exists.
func (m *Manager) Index() *Index {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.idx == nil {
		if es, err := readIndex(m.indexFile()); err == nil {
			m.idx = build(es)
		}
	}
	return m.idx
}

// Stale reports whether the cached index is missing or older than a week.
func (m *Manager) Stale() bool {
	fi, err := os.Stat(m.indexFile())
	return err != nil || time.Since(fi.ModTime()) > refreshAfter
}

// Refresh downloads the manifest when it changed and rebuilds the index.
// It reports whether a new index was built.
func (m *Manager) Refresh(ctx context.Context) (bool, error) {
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return false, nil
	}
	m.busy = true
	m.mu.Unlock()
	defer func() { m.mu.Lock(); m.busy = false; m.mu.Unlock() }()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return false, err
	}
	if tag, err := os.ReadFile(m.etagFile()); err == nil && platform.IsFile(m.indexFile()) {
		req.Header.Set("If-None-Match", strings.TrimSpace(string(tag)))
	}
	req.Header.Set("User-Agent", "WaterLauncher")
	c := &http.Client{Timeout: 3 * time.Minute}
	resp, err := c.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		now := time.Now()
		_ = os.Chtimes(m.indexFile(), now, now)
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}
	es, err := Parse(io.LimitReader(resp.Body, maxManifest))
	if err != nil {
		return false, err
	}
	if len(es) < 10000 {
		return false, fmt.Errorf("game database looks incomplete (%d entries)", len(es))
	}
	_ = os.MkdirAll(m.dir, 0o755)
	if err := writeIndex(m.indexFile(), es); err != nil {
		return false, err
	}
	if tag := resp.Header.Get("ETag"); tag != "" {
		_ = os.WriteFile(m.etagFile(), []byte(tag), 0o644)
	}
	m.mu.Lock()
	m.idx = build(es)
	m.mu.Unlock()
	return true, nil
}

func readIndex(p string) ([]Entry, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	var es []Entry
	return es, gob.NewDecoder(zr).Decode(&es)
}

func writeIndex(p string, es []Entry) error {
	tmp := p + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	if err := gob.NewEncoder(zw).Encode(es); err != nil {
		f.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Parse reads the manifest YAML. The file is machine-generated with a fixed
// two-space layout, so a line scanner does the job at a fraction of the
// time and memory a YAML decoder needs.
func Parse(r io.Reader) ([]Entry, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	var (
		out     []Entry
		cur     *Entry
		alias   string
		section string
	)
	aliases := map[string][]string{} // canonical name → alias names
	finish := func() {
		if cur == nil {
			return
		}
		if alias != "" {
			aliases[alias] = append(aliases[alias], cur.Name)
		} else {
			out = append(out, *cur)
		}
		cur, alias = nil, ""
	}
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line == "---" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		ind := len(line) - len(strings.TrimLeft(line, " "))
		t := strings.TrimSpace(line)
		switch {
		case ind == 0:
			finish()
			cur = &Entry{Name: unquote(strings.TrimSuffix(t, ":"))}
			section = ""
		case cur == nil:
		case ind == 2:
			k, v, _ := strings.Cut(t, ":")
			section = k
			if k == "alias" {
				alias = unquote(strings.TrimSpace(v))
			}
		case ind == 4 && section == "installDir":
			if dir := unquote(strings.TrimSuffix(strings.TrimSuffix(t, " {}"), ":")); dir != "" {
				cur.InstallDirs = append(cur.InstallDirs, dir)
			}
		case ind == 4 && (section == "steam" || section == "gog"):
			k, v, ok := strings.Cut(t, ":")
			if !ok || k != "id" {
				continue
			}
			v = strings.TrimSpace(v)
			if section == "steam" {
				if id, err := strconv.Atoi(v); err == nil && id > 0 {
					cur.SteamID = id
				}
			} else if _, err := strconv.ParseInt(v, 10, 64); err == nil {
				cur.GogID = v
			}
		}
	}
	finish()
	if err := sc.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Aliases = aliases[out[i].Name]
	}
	return out, nil
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' {
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'")
	}
	return s
}
