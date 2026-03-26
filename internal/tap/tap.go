package tap

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"agent-tools/internal/core"
	"agent-tools/internal/formula"
	"agent-tools/internal/paths"
)

// Tap is a cloned formula repository (Homebrew tap analogue).
type Tap struct {
	Name string `json:"name"` // short name e.g. "acme/tools"
	URL  string `json:"url"`
	Path string `json:"path"`
}

func LoadList() ([]Tap, error) {
	p, err := paths.TapListFile()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var w struct {
		Taps []Tap `json:"taps"`
	}
	if err := json.Unmarshal(b, &w); err != nil {
		return nil, err
	}
	return w.Taps, nil
}

func SaveList(taps []Tap) error {
	p, err := paths.TapListFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	w := struct {
		Taps []Tap `json:"taps"`
	}{Taps: taps}
	b, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Add clones url into taps dir and registers the tap (third-party taps; core is implicit).
func Add(name, gitURL string) error {
	if strings.EqualFold(name, "core") {
		return fmt.Errorf(`tap name "core" is reserved for the default formula repo; set AGENT_TOOLS_CORE_URL instead`)
	}
	base, err := paths.Taps()
	if err != nil {
		return err
	}
	safe := sanitizeTapDir(name)
	dest := filepath.Join(base, safe)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("tap %q already exists at %s", name, dest)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("git", "clone", "--depth", "1", gitURL, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w (is git installed?)", err)
	}
	taps, err := LoadList()
	if err != nil {
		return err
	}
	taps = append(taps, Tap{Name: name, URL: gitURL, Path: dest})
	return SaveList(taps)
}

func sanitizeTapDir(s string) string {
	s = strings.ReplaceAll(s, string(filepath.Separator), "-")
	s = strings.ReplaceAll(s, "..", "")
	return s
}

var ErrFormulaNotFound = errors.New("formula not found in taps")

// DescribeTaps lists core (when enabled) plus user taps without cloning the core repo.
func DescribeTaps() ([]Tap, error) {
	var out []Tap
	if u := core.GitURL(); u != "" {
		d, err := core.Dir()
		if err != nil {
			return nil, err
		}
		out = append(out, Tap{Name: "core", URL: u, Path: d})
	}
	user, err := LoadList()
	if err != nil {
		return nil, err
	}
	for _, t := range user {
		if strings.EqualFold(t.Name, "core") {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

// AllTaps returns the implicit core tap first when present on disk, then user-added taps.
// Call core.Ensure() before this when formulas from core may be needed.
func AllTaps() ([]Tap, error) {
	var out []Tap
	if core.GitURL() != "" {
		if p, ok := core.TapPath(); ok {
			out = append(out, Tap{Name: "core", URL: core.GitURL(), Path: p})
		}
	}
	user, err := LoadList()
	if err != nil {
		return nil, err
	}
	for _, t := range user {
		if strings.EqualFold(t.Name, "core") {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

// FindFormula searches Formulas/*.json in core first, then other taps.
func FindFormula(toolName string) (*formula.Formula, string, error) {
	_ = core.Ensure()
	taps, err := AllTaps()
	if err != nil {
		return nil, "", err
	}
	for _, t := range taps {
		dir := filepath.Join(t.Path, "Formulas")
		for _, fn := range []string{toolName + ".json", strings.ToLower(toolName) + ".json"} {
			p := filepath.Join(dir, fn)
			if f, err := formula.LoadFile(p); err == nil {
				f.Tap = t.Name
				return f, p, nil
			}
		}
	}
	return nil, "", ErrFormulaNotFound
}

// FindFormulaFile resolves an explicit path or name in taps.
func FindFormulaFile(arg string) (*formula.Formula, string, error) {
	if strings.HasSuffix(arg, ".json") && fileExists(arg) {
		f, err := formula.LoadFile(arg)
		return f, arg, err
	}
	f, p, err := FindFormula(arg)
	return f, p, err
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// ListFormulas returns all .json formulas from core and registered taps.
func ListFormulas() ([]string, error) {
	_ = core.Ensure()
	taps, err := AllTaps()
	if err != nil {
		return nil, err
	}
	var out []string
	for _, t := range taps {
		dir := filepath.Join(t.Path, "Formulas")
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			out = append(out, filepath.Join(t.Name, strings.TrimSuffix(e.Name(), ".json")))
		}
	}
	return out, nil
}
