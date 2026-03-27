package paths

import (
	"os"
	"path/filepath"
)

// Root is ~/.tars
func Root() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tars"), nil
}

func Installs() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "installs"), nil
}

// Bin is where symlinks to installed executables live (add to PATH).
func Bin() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "bin"), nil
}

func Taps() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "taps"), nil
}

func Cache() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "cache", "downloads"), nil
}

// CatalogDir holds the merged model-facing index (readable by any agent/model).
func CatalogDir() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "catalog"), nil
}

func RegistryFile() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "registry.json"), nil
}

func TapListFile() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "taps.json"), nil
}

// ToolsMarkdown is the standalone agent-facing doc listing connected tools (~/.tars/tools.md).
func ToolsMarkdown() (string, error) {
	r, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(r, "tools.md"), nil
}
