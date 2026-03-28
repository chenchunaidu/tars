package formula

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// PlatformDefault is the platforms map key used when the host GOOS/GOARCH has no entry.
const PlatformDefault = "default"

// PlatformArtifact is one release asset (url + sha256) for a given GOOS_GOARCH or "default".
type PlatformArtifact struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// Formula describes an installable agent tool (Homebrew-style manifest).
type Formula struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Revision    int    `json:"revision,omitempty"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	License     string `json:"license,omitempty"`

	// URL is the artifact URL when platforms is omitted (tar.gz, zip, or single binary).
	URL string `json:"url,omitempty"`

	// SHA256 is required for integrity when platforms is omitted.
	SHA256 string `json:"sha256,omitempty"`

	// Platforms maps "linux_amd64", "darwin_arm64", "windows_amd64", etc., and optional "default".
	// When non-empty, install picks the host key from goos/goarch, else "default", else errors.
	Platforms map[string]PlatformArtifact `json:"platforms,omitempty"`

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
	f.Platforms = normalizePlatforms(f.Platforms)
	f.Name = strings.TrimSpace(f.Name)
	f.Version = strings.TrimSpace(f.Version)
	f.URL = strings.TrimSpace(f.URL)
	f.SHA256 = strings.TrimSpace(f.SHA256)
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

func normalizePlatforms(m map[string]PlatformArtifact) map[string]PlatformArtifact {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]PlatformArtifact, len(m))
	for k, v := range m {
		nk := strings.ToLower(strings.TrimSpace(k))
		if nk == "" {
			continue
		}
		out[nk] = PlatformArtifact{
			URL:    strings.TrimSpace(v.URL),
			SHA256: strings.TrimSpace(v.SHA256),
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// PlatformKey is the lookup key for Platforms: lowercase goos + "_" + lowercase goarch (Go's names).
func PlatformKey(goos, goarch string) string {
	return strings.ToLower(goos) + "_" + strings.ToLower(goarch)
}

// ResolveArtifact returns the download URL and SHA256 for goos/goarch.
// When platforms is set, Selected is the map key used ("linux_amd64", "default", etc.).
// When using legacy url/sha256 only, Selected is empty.
func (f *Formula) ResolveArtifact(goos, goarch string) (url, sha256, selected string, err error) {
	if f == nil {
		return "", "", "", errors.New("nil formula")
	}
	if len(f.Platforms) > 0 {
		key := PlatformKey(goos, goarch)
		if a, ok := f.Platforms[key]; ok && a.URL != "" {
			return a.URL, a.SHA256, key, nil
		}
		if a, ok := f.Platforms[PlatformDefault]; ok && a.URL != "" {
			return a.URL, a.SHA256, PlatformDefault, nil
		}
		return "", "", "", fmt.Errorf(
			`formula: no platforms entry for %q and no %q fallback`,
			key, PlatformDefault,
		)
	}
	if f.URL == "" {
		return "", "", "", errors.New("formula: url is required when platforms is empty")
	}
	if f.SHA256 == "" {
		return "", "", "", errors.New("formula: sha256 is required when platforms is empty")
	}
	return f.URL, f.SHA256, "", nil
}

func (f *Formula) Validate() error {
	if f == nil {
		return errors.New("nil formula")
	}
	if f.Name == "" {
		return errors.New("formula: name is required")
	}
	if f.Version == "" {
		return errors.New("formula: version is required")
	}
	if len(f.Platforms) > 0 {
		for key, a := range f.Platforms {
			if a.URL == "" {
				return fmt.Errorf("formula: platforms[%q] missing url", key)
			}
			if a.SHA256 == "" {
				return fmt.Errorf("formula: platforms[%q] missing sha256", key)
			}
			if !looksLikeSHA256(a.SHA256) {
				return fmt.Errorf("formula: platforms[%q] sha256 must be 64 hex characters", key)
			}
		}
		return nil
	}
	if f.URL == "" {
		return errors.New("formula: url is required when platforms is empty")
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

// AgentUsageText returns the text to show agents: description plus model.summary.
func (f *Formula) AgentUsageText() string {
	if f == nil {
		return ""
	}
	var parts []string
	if d := strings.TrimSpace(f.Description); d != "" {
		parts = append(parts, d)
	}
	if f.Model != nil {
		if s := strings.TrimSpace(f.Model.Summary); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n\n")
}
