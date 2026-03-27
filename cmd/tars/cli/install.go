package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tars/internal/agentconnect"
	"tars/internal/binlink"
	"tars/internal/catalog"
	"tars/internal/formula"
	"tars/internal/install"
	"tars/internal/paths"
	"tars/internal/registry"
	"tars/internal/tap"
	"tars/internal/toolsmd"
)

func cmdInstall() *cobra.Command {
	return &cobra.Command{
		Use:   "install [formula name or path]",
		Short: "Download, verify SHA256, and install a tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, _, err := tap.FindFormulaFile(args[0])
			if err != nil {
				return err
			}
			return runInstall(f)
		},
	}
}

func runInstall(f *formula.Formula) error {
	cacheDir, err := paths.Cache()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}
	base := filepath.Base(f.URL)
	if base == "" || base == "/" {
		base = f.Name + "-artifact"
	}
	cached := filepath.Join(cacheDir, fmt.Sprintf("%s-%s-%s", f.Name, f.Version, sanitizeFilePart(base)))

	fmt.Printf("==> Downloading %s\n", f.URL)
	if err := install.Download(f.URL, cached, f.SHA256); err != nil {
		return err
	}
	fmt.Println("==> SHA256 verified")

	prefix, err := paths.Installs()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(prefix, 0o755); err != nil {
		return err
	}

	fmt.Println("==> Installing")
	verDir, err := install.Install(f, cached, prefix)
	if err != nil {
		return err
	}

	binNames := f.Bin
	if len(binNames) == 0 {
		binNames = guessBinaries(verDir)
	}
	if err := binlink.Link(verDir, binNames); err != nil {
		return err
	}

	reg, err := registry.Open()
	if err != nil {
		return err
	}
	entry := registry.Entry{
		Name:        f.Name,
		Version:     f.Version,
		Tap:         f.Tap,
		InstallPath: verDir,
		ArtifactURL: f.URL,
		SHA256:      f.SHA256,
		InstalledAt: time.Now().UTC(),
	}
	if err := reg.Set(entry); err != nil {
		return err
	}

	usage := strings.TrimSpace(f.Usage)
	if usage == "" {
		usage = f.AgentUsageText()
	}
	rec := catalog.ToolRecord{
		Name:        f.Name,
		Version:     f.Version,
		Tap:         f.Tap,
		Description: strings.TrimSpace(f.Description),
		Usage:       usage,
		InstallPath: verDir,
		ArtifactURL: f.URL,
		SHA256:      f.SHA256,
		UpdatedAt:   time.Now().UTC(),
		Model:       f.Model,
	}
	if err := catalog.WriteTool(rec); err != nil {
		return err
	}
	toolsPath, err := toolsmd.Refresh()
	if err != nil {
		return err
	}
	if err := agentconnect.Apply(toolsPath, agentconnect.Options{}); err != nil {
		fmt.Fprintf(os.Stderr, "tars: connect agents (run 'tars connect' to retry): %v\n", err)
	}

	bin, _ := paths.Bin()
	fmt.Printf("==> Installed %s %s\n", f.Name, f.Version)
	fmt.Printf("    Prefix: %s\n", verDir)
	if bin != "" {
		fmt.Printf("    Add to PATH: %s\n", bin)
	}
	fmt.Printf("    Model catalog: ~/.tars/catalog/tools.json\n")
	fmt.Printf("    Agent tools doc: %s\n", toolsPath)
	return nil
}

func guessBinaries(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0o111 != 0 {
			names = append(names, e.Name())
		}
	}
	return names
}

func sanitizeFilePart(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '?', '*':
			return '-'
		}
		return r
	}, s)
	if len(s) > 120 {
		return s[:120]
	}
	return s
}
