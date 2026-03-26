package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent-tools",
	Short: "Package manager for agent tools (Homebrew-style)",
	Long: `agent-tools installs tools into ~/.agent-tools, verifies SHA256 checksums,
and maintains a shared catalog (~/.agent-tools/catalog/tools.json) for model/agent usage metadata.

Formulas come from the default core repository (homebrew-core analogue, cloned to
~/.agent-tools/taps/core) plus any extra taps added with "tap add". Override the
core URL with AGENT_TOOLS_CORE_URL.`,
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
		cmdHash(),
		cmdUpdate(),
	)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
