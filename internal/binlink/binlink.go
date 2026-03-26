package binlink

import (
	"os"
	"path/filepath"

	"agent-tools/internal/paths"
)

// Link creates symlinks in ~/.agent-tools/bin for each name pointing to installDir/name.
func Link(installDir string, names []string) error {
	bin, err := paths.Bin()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(bin, 0o755); err != nil {
		return err
	}
	for _, n := range names {
		src := filepath.Join(installDir, n)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dst := filepath.Join(bin, n)
		_ = os.Remove(dst)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// Unlink removes symlinks for names from ~/.agent-tools/bin.
func Unlink(names []string) error {
	bin, err := paths.Bin()
	if err != nil {
		return err
	}
	for _, n := range names {
		dst := filepath.Join(bin, n)
		_ = os.Remove(dst)
	}
	return nil
}
