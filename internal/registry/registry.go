package registry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tars/internal/paths"
)

// Entry is one installed tool.
type Entry struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Tap         string    `json:"tap,omitempty"`
	InstallPath string    `json:"install_path"`
	ArtifactURL string    `json:"artifact_url"`
	SHA256      string    `json:"sha256"`
	InstalledAt time.Time `json:"installed_at"`
}

// Registry tracks installed tools on disk.
type Registry struct {
	mu     sync.Mutex
	path   string
	byName map[string]Entry
}

func Open() (*Registry, error) {
	p, err := paths.RegistryFile()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return nil, err
	}
	r := &Registry{path: p, byName: map[string]Entry{}}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, err
	}
	var wrapper struct {
		Tools []Entry `json:"tools"`
	}
	if err := json.Unmarshal(b, &wrapper); err != nil {
		return nil, err
	}
	for _, e := range wrapper.Tools {
		r.byName[e.Name] = e
	}
	return r, nil
}

func (r *Registry) Get(name string) (Entry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.byName[name]
	return e, ok
}

func (r *Registry) Set(e Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byName[e.Name] = e
	return r.persistLocked()
}

func (r *Registry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byName, name)
	return r.persistLocked()
}

func (r *Registry) List() []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Entry, 0, len(r.byName))
	for _, e := range r.byName {
		out = append(out, e)
	}
	return out
}

func (r *Registry) persistLocked() error {
	wrapper := struct {
		Tools []Entry `json:"tools"`
	}{}
	for _, e := range r.byName {
		wrapper.Tools = append(wrapper.Tools, e)
	}
	b, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

// ErrNotFound when uninstalling missing tool.
var ErrNotFound = errors.New("tool not in registry")
