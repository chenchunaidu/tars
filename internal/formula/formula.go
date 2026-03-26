package formula

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Formula describes an installable agent tool (Homebrew-style manifest).
type Formula struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Revision    int    `json:"revision,omitempty"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	License     string `json:"license,omitempty"`

	// URL is the artifact URL (tar.gz, zip, or single binary).
	URL string `json:"url"`

	// SHA256 is required for integrity (Homebrew bottle-style).
	SHA256 string `json:"sha256"`

	// Bin lists executable names expected in the archive (or root for single file).
	Bin []string `json:"bin,omitempty"`

	// Install maps paths inside the archive to prefix-relative paths.
	Install []InstallMapping `json:"install,omitempty"`

	// Model holds agent/model-facing usage metadata stored in the shared catalog.
	Model *ModelMeta `json:"model,omitempty"`

	// Tap is set when loaded from a tap (not in file).
	Tap string `json:"-"`
}

type InstallMapping struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ModelMeta is the contract for how tools are invoked by agents or IDEs.
// Kept flexible with Raw for MCP JSON Schema or OpenAPI fragments.
type ModelMeta struct {
	Summary      string          `json:"summary,omitempty"`
	Invocation   string          `json:"invocation,omitempty"` // e.g. "cli", "mcp", "http"
	Parameters   json.RawMessage `json:"parameters,omitempty"`
	Examples     []string        `json:"examples,omitempty"`
	Requirements []string        `json:"requirements,omitempty"`
	Raw          json.RawMessage `json:"raw,omitempty"` // extension point
}

func LoadFile(path string) (*Formula, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

func Parse(data []byte) (*Formula, error) {
	var f Formula
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

func (f *Formula) Validate() error {
	if f == nil {
		return errors.New("nil formula")
	}
	f.Name = strings.TrimSpace(f.Name)
	f.Version = strings.TrimSpace(f.Version)
	f.URL = strings.TrimSpace(f.URL)
	f.SHA256 = strings.TrimSpace(f.SHA256)
	if f.Name == "" {
		return errors.New("formula: name is required")
	}
	if f.Version == "" {
		return errors.New("formula: version is required")
	}
	if f.URL == "" {
		return errors.New("formula: url is required")
	}
	if f.SHA256 == "" {
		return errors.New("formula: sha256 is required for security verification")
	}
	if !looksLikeSHA256(f.SHA256) {
		return fmt.Errorf("formula: sha256 must be 64 hex characters")
	}
	return nil
}

func looksLikeSHA256(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if c >= '0' && c <= '9' {
			continue
		}
		if c >= 'a' && c <= 'f' {
			continue
		}
		if c >= 'A' && c <= 'F' {
			continue
		}
		return false
	}
	return true
}
