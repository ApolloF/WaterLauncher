package addons

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Approval is the user's decision about one add-on.
type Approval struct {
	Enabled  bool      `json:"enabled"`
	SHA256   string    `json:"sha256"` // the program as approved
	Approved time.Time `json:"approved"`
}

// Trust keeps approvals in a JSON file. Safe for concurrent use.
type Trust struct {
	path string
	mu   sync.Mutex
	m    map[string]Approval
}

// OpenTrust reads the approvals file (none when missing or damaged).
func OpenTrust(path string) *Trust {
	t := &Trust{path: path, m: map[string]Approval{}}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &t.m)
	}
	return t
}

// Get returns an add-on's approval.
func (t *Trust) Get(id string) Approval {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.m[id]
}

// Set stores an add-on's approval (a zero one forgets it).
func (t *Trust) Set(id string, a Approval) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if a == (Approval{}) {
		delete(t.m, id)
	} else {
		t.m[id] = a
	}
	b, err := json.MarshalIndent(t.m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(t.path), 0o755); err != nil {
		return err
	}
	tmp := t.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, t.path)
}
