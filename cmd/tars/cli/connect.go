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
	var noCursor, noClaude, noGemini, noPi bool
	c := &cobra.Command{
		Use:   "connect",
		Short: "Regenerate ~/.tars/tools.md and wire Cursor, Claude, Gemini CLI, and Pi",
		Long: `Rebuilds ~/.tars/tools.md and updates global agent instructions:

  • Cursor: ~/.cursor/rules/tars-tools.mdc (alwaysApply)
  • Claude Code: ~/.claude/CLAUDE.md
  • Gemini CLI: ~/.gemini/GEMINI.md
  • Pi coding agent: ~/.pi/agent/AGENTS.md

Each Markdown target gets the same managed <!-- tars-connect --> block: read tools.md when
the task may involve tars-installed CLIs; otherwise continue normally.

Use --no-cursor, --no-claude, --no-gemini, or --no-pi to skip one. Formulas should include a
"usage" field (or "description" plus optional "model.summary") for best results in tools.md.`,
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
			opts := agentconnect.Options{
				SkipCursor: noCursor, SkipClaude: noClaude,
				SkipGemini: noGemini, SkipPi: noPi,
			}
			if err := agentconnect.Apply(out, opts); err != nil {
				return err
			}
			if !noCursor {
				fmt.Printf("Wrote Cursor global rule: %s\n", agentconnect.CursorRulePath(home))
			}
			if !noClaude {
				fmt.Printf("Updated Claude Code global: %s\n", agentconnect.ClaudeGlobalPath(home))
			}
			if !noGemini {
				fmt.Printf("Updated Gemini CLI global: %s\n", agentconnect.GeminiGlobalPath(home))
			}
			if !noPi {
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
	c.Flags().BoolVar(&noCursor, "no-cursor", false, "do not write ~/.cursor/rules/tars-tools.mdc")
	c.Flags().BoolVar(&noClaude, "no-claude", false, "do not merge the block into ~/.claude/CLAUDE.md")
	c.Flags().BoolVar(&noGemini, "no-gemini", false, "do not merge the block into ~/.gemini/GEMINI.md")
	c.Flags().BoolVar(&noPi, "no-pi", false, "do not merge the block into ~/.pi/agent/AGENTS.md")
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
