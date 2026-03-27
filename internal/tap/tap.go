package tap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"tars/internal/core"
	"tars/internal/formula"
	"tars/internal/paths"
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
		return fmt.Errorf(`tap name "core" is reserved for the default formula repo; set TARS_CORE_URL instead`)
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

// formulaFirstLetterDir picks the Homebrew-style letter bucket under Formulas/:
// a–z for names starting with an ASCII letter, "0" for a leading digit, "@" otherwise.
func formulaFirstLetterDir(toolName string) string {
	toolName = strings.TrimSpace(toolName)
	if toolName == "" {
		return "@"
	}
	r, w := utf8.DecodeRuneInString(toolName)
	if r == utf8.RuneError && w == 1 {
		return "@"
	}
	if r < utf8.RuneSelf {
		c := unicode.ToLower(r)
		switch {
		case c >= 'a' && c <= 'z':
			return string(c)
		case c >= '0' && c <= '9':
			return "0"
		}
		return "@"
	}
	l := unicode.ToLower(r)
	if l >= 'a' && l <= 'z' {
		return string(l)
	}
	if unicode.IsDigit(r) {
		return "0"
	}
	return "@"
}

func formulaJSONPaths(dir, toolName string) []string {
	low := strings.ToLower(toolName)
	fn1 := toolName + ".json"
	fn2 := low + ".json"
	sub := formulaFirstLetterDir(toolName)

	seen := map[string]struct{}{}
	var paths []string
	add := func(p string) {
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}
	add(filepath.Join(dir, fn1))
	if fn2 != fn1 {
		add(filepath.Join(dir, fn2))
	}
	add(filepath.Join(dir, sub, fn1))
	if fn2 != fn1 {
		add(filepath.Join(dir, sub, fn2))
	}
	return paths
}

func collectFormulaJSONFiles(formulasDir string) ([]string, error) {
	if _, err := os.Stat(formulasDir); os.IsNotExist(err) {
		return nil, nil
	}
	var out []string
	err := filepath.WalkDir(formulasDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".json") {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

// FindFormula searches Formulas (flat or Formulas/<letter>/ like Homebrew) in core first, then other taps.
func FindFormula(toolName string) (*formula.Formula, string, error) {
	_ = core.Ensure()
	taps, err := AllTaps()
	if err != nil {
		return nil, "", err
	}
	for _, t := range taps {
		dir := filepath.Join(t.Path, "Formulas")
		for _, p := range formulaJSONPaths(dir, toolName) {
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
		paths, err := collectFormulaJSONFiles(dir)
		if err != nil {
			continue
		}
		for _, p := range paths {
			name := strings.TrimSuffix(filepath.Base(p), ".json")
			out = append(out, filepath.Join(t.Name, name))
		}
	}
	return out, nil
}
