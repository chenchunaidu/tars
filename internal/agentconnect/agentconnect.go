package agentconnect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	cursorRuleFile = "tars-tools.mdc"
	sectionBegin   = "<!-- tars-connect:begin -->"
	sectionEnd     = "<!-- tars-connect:end -->"
)

// Options control which global agent files are written.
type Options struct {
	SkipCursor bool
	SkipClaude bool
	SkipGemini bool
	SkipPi     bool
}

// Apply writes global agent instructions so assistants read ~/.tars/tools.md when relevant.
// toolsMD must be the absolute path to ~/.tars/tools.md.
func Apply(toolsMD string, opts Options) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	toolsMD = filepath.Clean(toolsMD)
	var errs []string
	if !opts.SkipCursor {
		if err := writeCursorRule(home, toolsMD); err != nil {
			errs = append(errs, "cursor: "+err.Error())
		}
	}
	if !opts.SkipClaude {
		if err := mergeGlobalMarkdown(ClaudeGlobalPath(home), toolsMD); err != nil {
			errs = append(errs, "claude: "+err.Error())
		}
	}
	if !opts.SkipGemini {
		if err := mergeGlobalMarkdown(GeminiGlobalPath(home), toolsMD); err != nil {
			errs = append(errs, "gemini: "+err.Error())
		}
	}
	if !opts.SkipPi {
		if err := mergeGlobalMarkdown(PiGlobalPath(home), toolsMD); err != nil {
			errs = append(errs, "pi: "+err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

func writeCursorRule(home, toolsAbs string) error {
	dir := filepath.Join(home, ".cursor", "rules")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString(`---
description: Tars — read tools.md each session; match tools, then consult their docs
alwaysApply: true
---

# Tars installed tools

Use a **two-step** workflow for binaries installed via **tars**.

## Step 1 — Read the catalog

- **Always read** **`)
	b.WriteString(toolsAbs)
	b.WriteString(`** near the **start of the session** or **before** you act on work that could involve a local CLI. You do **not** need the user to mention command-line tools, CLIs, or **~/.tars** first.
- Use it to see **which tools exist**, paths such as **~/.tars/bin**, and each tool’s **description** (plus optional notes) **tars** recorded.
- **If** the file is **missing**, says there are no tools, or **no listed tool matches** the current task, **stop here** and **continue normally** — do not block on tars.

## Step 2 — Docs before invocation

- When a **listed tool matches** the task, **do not guess** flags, subcommands, or behavior.
- **Consult that tool’s documentation** next, in order: what **tools.md** already states (primarily **description**); then the binary’s own help (e.g. **`)
	b.WriteString("<tool-name>")
	b.WriteString(` --help** / **-h**), or a **docs**-style subcommand when that applies (e.g. **`)
	b.WriteString("<tool-name>")
	b.WriteString(` docs**).
- Prefer those sources over memory.

Regenerate **tools.md** and refresh this rule with: `)
	b.WriteString("`tars connect`")
	b.WriteString(`.

`)
	body := b.String()
	p := filepath.Join(dir, cursorRuleFile)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func mergeGlobalMarkdown(mdPath, toolsAbs string) error {
	block := buildTarsSectionBlock(toolsAbs)
	old, err := os.ReadFile(mdPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(old)
	if strings.Contains(content, sectionBegin) && strings.Contains(content, sectionEnd) {
		content = replaceMarkedBlock(content, block)
	} else {
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if content != "" {
			content += "\n"
		}
		content += block + "\n"
	}
	if err := os.MkdirAll(filepath.Dir(mdPath), 0o755); err != nil {
		return err
	}
	tmp := mdPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, mdPath)
}

func buildTarsSectionBlock(toolsAbs string) string {
	var b strings.Builder
	b.WriteString(sectionBegin)
	b.WriteString("\n\n## Tars (CLI tools)\n\n")
	b.WriteString("**Step 1:** Near the start of a session or before work that might use a local CLI, **read** **")
	b.WriteString(toolsAbs)
	b.WriteString("** (always — the user need not mention CLIs or **~/.tars**). If missing, empty, or no tool matches the task, **continue normally**.\n\n")
	b.WriteString("**Step 2:** If a listed tool **does** match, **do not guess** — use what **tools.md** says (description first), then **")
	b.WriteString("<tool-name>")
	b.WriteString(" --help** / **-h** or **")
	b.WriteString("<tool-name>")
	b.WriteString(" docs** when applicable.\n\n")
	b.WriteString("Keep **~/.tars/bin** on PATH. Refresh: `tars connect`.\n\n")
	b.WriteString(sectionEnd)
	return b.String()
}

func replaceMarkedBlock(s, newBlock string) string {
	start := strings.Index(s, sectionBegin)
	end := strings.Index(s, sectionEnd)
	if start == -1 || end == -1 || end < start {
		return strings.TrimRight(s, "\n") + "\n\n" + newBlock + "\n"
	}
	end += len(sectionEnd)
	return s[:start] + newBlock + s[end:]
}

// CursorRulePath returns ~/.cursor/rules/tars-tools.mdc for display.
func CursorRulePath(home string) string {
	return filepath.Join(home, ".cursor", "rules", cursorRuleFile)
}

// ClaudeGlobalPath returns ~/.claude/CLAUDE.md for display.
func ClaudeGlobalPath(home string) string {
	return filepath.Join(home, ".claude", "CLAUDE.md")
}

// GeminiGlobalPath returns ~/.gemini/GEMINI.md (Gemini CLI global context).
func GeminiGlobalPath(home string) string {
	return filepath.Join(home, ".gemini", "GEMINI.md")
}

// PiGlobalPath returns ~/.pi/agent/AGENTS.md (Pi coding agent global instructions).
func PiGlobalPath(home string) string {
	return filepath.Join(home, ".pi", "agent", "AGENTS.md")
}
