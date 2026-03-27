package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tars/internal/formula"
)

func cmdPublish() *cobra.Command {
	root := &cobra.Command{
		Use:   "publish",
		Short: "Validate formulas for publishing (Homebrew-style PR workflow)",
	}
	root.AddCommand(
		&cobra.Command{
			Use:   "validate <formula.json>",
			Short: "Validate formula JSON and required security fields",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				f, err := formula.LoadFile(args[0])
				if err != nil {
					return err
				}
				b, err := json.MarshalIndent(f, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				fmt.Println("OK: formula is valid (sha256 present).")
				return nil
			},
		},
		&cobra.Command{
			Use:   "init <name>",
			Short: "Write a template formula.json for a new tool",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				name := args[0]
				tmpl := formula.Formula{
					Name:        name,
					Version:     "0.0.1",
					Description: "One-line summary (Homebrew-style).",
					Usage: "A few lines for coding agents: how to invoke this tool, typical flags, " +
						"and when the user should reach for it (shown in ~/.tars/tools.md).",
					URL:    "https://example.com/releases/" + name + "-0.0.1.tar.gz",
					SHA256: "REPLACE_WITH_SHA256_OF_RELEASE_ARTIFACT",
					Bin:    []string{name},
					Model: &formula.ModelMeta{
						Summary:    "Optional extra one-liner; merged into tools.md if usage is empty.",
						Invocation: "cli",
						Examples:   []string{name + " --help"},
					},
				}
				b, err := json.MarshalIndent(tmpl, "", "  ")
				if err != nil {
					return err
				}
				out := name + ".json"
				if err := os.WriteFile(out, b, 0o644); err != nil {
					return err
				}
				fmt.Printf("Wrote %s — fill url, sha256, and model metadata, then open a PR to your tap repo under Formulas/\n", out)
				return nil
			},
		},
	)
	return root
}
