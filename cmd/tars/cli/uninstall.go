package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"tars/internal/agentconnect"
	"tars/internal/binlink"
	"tars/internal/catalog"
	"tars/internal/registry"
	"tars/internal/toolsmd"
)

func cmdUninstall() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <name>",
		Short: "Remove an installed tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			reg, err := registry.Open()
			if err != nil {
				return err
			}
			e, ok := reg.Get(name)
			if !ok {
				return fmt.Errorf("%s: %w", name, registry.ErrNotFound)
			}
			binNames := guessBinariesFromPrefix(e.InstallPath)
			_ = binlink.Unlink(binNames)
			if e.InstallPath != "" {
				_ = os.RemoveAll(e.InstallPath)
				parent := filepath.Dir(e.InstallPath)
				_ = os.Remove(parent) // remove empty name dir if possible
			}
			if err := reg.Remove(name); err != nil {
				return err
			}
			if err := catalog.RemoveTool(name); err != nil {
				return err
			}
			toolsPath, err := toolsmd.Refresh()
			if err != nil {
				return err
			}
			if err := agentconnect.Apply(toolsPath, agentconnect.Options{}); err != nil {
				fmt.Fprintf(os.Stderr, "tars: connect agents (run 'tars connect all' to retry): %v\n", err)
			}
			fmt.Printf("Uninstalled %s\n", name)
			return nil
		},
	}
}

func guessBinariesFromPrefix(dir string) []string {
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
