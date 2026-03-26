package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"agent-tools/internal/paths"
)

const (
	// DefaultCoreGitURL is the canonical formula repo (homebrew-core analogue).
	// Override with AGENT_TOOLS_CORE_URL.
	DefaultCoreGitURL = "https://github.com/agent-tools/homebrew-core.git"
)

// GitURL returns the git remote for the core formula repository.
// Set AGENT_TOOLS_CORE_URL to "" or "none" to disable the core tap.
func GitURL() string {
	v := strings.TrimSpace(os.Getenv("AGENT_TOOLS_CORE_URL"))
	if v == "" {
		return DefaultCoreGitURL
	}
	if strings.EqualFold(v, "none") || v == "-" {
		return ""
	}
	return v
}

// Dir is ~/.agent-tools/taps/core (single clone of homebrew-core-style repo).
func Dir() (string, error) {
	base, err := paths.Taps()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "core"), nil
}

// TapPath returns the path if the core repo exists on disk (clone completed).
func TapPath() (string, bool) {
	d, err := Dir()
	if err != nil {
		return "", false
	}
	if st, err := os.Stat(filepath.Join(d, ".git")); err != nil || !st.IsDir() {
		return "", false
	}
	return d, true
}

// Ensure clones the core repo on first use (like implicit homebrew/core).
func Ensure() error {
	url := GitURL()
	if url == "" {
		return nil
	}
	dest, err := Dir()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("core: git clone %s: %w", url, err)
	}
	return nil
}

// Update runs git pull inside the core tap directory.
func Update() error {
	url := GitURL()
	if url == "" {
		return nil
	}
	d, ok := TapPath()
	if !ok {
		return Ensure()
	}
	cmd := exec.Command("git", "-C", d, "pull", "--ff-only")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("core: git pull: %w", err)
	}
	return nil
}
