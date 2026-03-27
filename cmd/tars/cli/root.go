package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tars/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "tars",
	Short: "Package manager for agent tools (Homebrew-style)",
	Long: `tars installs tools into ~/.tars, verifies SHA256 checksums,
and maintains a shared catalog (~/.tars/catalog/tools.json) plus a standalone
~/.tars/tools.md for coding agents. "tars connect" (and install/uninstall) refresh
that doc and update global instructions for Cursor, Claude Code, Gemini CLI, and Pi
so assistants check tools.md when relevant.

Formulas come from the default core repository (homebrew-core analogue, cloned to
~/.tars/taps/core) plus any extra taps added with "tap add". Override the
core URL with NXTOOLS_CORE_URL.`,
	Version: version.Version,
	Example: `  tars --version
  tars update && tars install ripgrep
  tars list
  tars list --available
  tars connect --copy .
  tars help install`,
}

func Execute() error {
	rootCmd.AddCommand(
		cmdInstall(),
		cmdUninstall(),
		cmdList(),
		cmdInfo(),
		cmdTap(),
		cmdPublish(),
		cmdCatalog(),
		cmdConnect(),
		cmdHash(),
		cmdUpdate(),
	)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
