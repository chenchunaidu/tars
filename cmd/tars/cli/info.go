package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tars/internal/paths"
	"tars/internal/registry"
	"tars/internal/tap"
)

func cmdInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show formula or installed tool details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			reg, err := registry.Open()
			if err != nil {
				return err
			}
			if e, ok := reg.Get(name); ok {
				fmt.Printf("Installed: %s %s\n", e.Name, e.Version)
				fmt.Printf("  Path: %s\n", e.InstallPath)
				fmt.Printf("  URL:  %s\n", e.ArtifactURL)
				fmt.Printf("  SHA256 (verified): %s\n", e.SHA256)
				if e.Tap != "" {
					fmt.Printf("  Tap: %s\n", e.Tap)
				}
				return nil
			}
			f, path, err := tap.FindFormula(name)
			if err != nil {
				return err
			}
			fmt.Printf("Formula: %s (from %s)\n", f.Name, path)
			b, _ := json.MarshalIndent(f, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
}

func cmdCatalog() *cobra.Command {
	return &cobra.Command{
		Use:   "catalog",
		Short: "Print path to the merged model catalog JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := paths.CatalogDir()
			if err != nil {
				return err
			}
			p := d + "/tools.json"
			fmt.Println(p)
			if _, err := os.Stat(p); os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "(file not created yet — install a tool first)")
			}
			return nil
		},
	}
}
