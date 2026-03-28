package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"tars/internal/agentconnect"
	"tars/internal/paths"
	"tars/internal/toolsmd"
)

func cmdConnect() *cobra.Command {
	var copyTo string
	c := &cobra.Command{
		Use:   "connect (all | AGENT [AGENT...])",
		Short: "Regenerate ~/.tars/tools.md and wire chosen coding agents",
		Long: `Rebuilds ~/.tars/tools.md, then updates global agent instructions only for the agents you name.

  • cursor   — ~/.cursor/rules/tars-tools.mdc (alwaysApply)
  • claude   — ~/.claude/CLAUDE.md
  • gemini   — ~/.gemini/GEMINI.md
  • pi       — ~/.pi/agent/AGENTS.md

Use ` + "`tars connect all`" + ` to update every agent. Otherwise pass one or more agent names, e.g.
` + "`tars connect cursor`" + ` or ` + "`tars connect cursor claude`" + `.

` + "`tars install`" + ` / ` + "`tars uninstall`" + ` still run connect for all agents automatically.

Formulas should include a "description" (plus optional "model.summary") for best results in tools.md.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := toolsmd.Refresh()
			if err != nil {
				return err
			}
			fmt.Printf("Wrote agent tools doc: %s\n", out)
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			opts, err := agentconnect.OptionsFromConnectArgs(args)
			if err != nil {
				return err
			}
			if err := agentconnect.Apply(out, opts); err != nil {
				return err
			}
			if !opts.SkipCursor {
				fmt.Printf("Wrote Cursor global rule: %s\n", agentconnect.CursorRulePath(home))
			}
			if !opts.SkipClaude {
				fmt.Printf("Updated Claude Code global: %s\n", agentconnect.ClaudeGlobalPath(home))
			}
			if !opts.SkipGemini {
				fmt.Printf("Updated Gemini CLI global: %s\n", agentconnect.GeminiGlobalPath(home))
			}
			if !opts.SkipPi {
				fmt.Printf("Updated Pi global: %s\n", agentconnect.PiGlobalPath(home))
			}
			if copyTo != "" {
				dst := filepath.Join(copyTo, "tools.md")
				if err := copyFileConnect(out, dst); err != nil {
					return err
				}
				fmt.Printf("Copied tools.md to: %s\n", dst)
			}
			bin, _ := paths.Bin()
			if bin != "" {
				fmt.Printf("\nEnsure %s is on your PATH in terminals where you run tars-installed binaries.\n", bin)
			}
			return nil
		},
	}
	c.Flags().StringVar(&copyTo, "copy", "", "copy tools.md into this directory (e.g. current project root)")
	return c
}

func copyFileConnect(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
