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
	SkipCursor  bool
	SkipClaude  bool
	SkipGemini  bool
	SkipPi      bool
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
description: Tars — use tools.md when the task involves tars-installed CLIs
alwaysApply: true
---

# Tars installed tools

When a session starts or the user asks about **command-line tools**, **CLIs**, or binaries under **~/.tars**, decide whether this applies:

- **If** the question may involve tools installed with **tars** (or paths like **~/.tars/bin**), **read** **`)
	b.WriteString(toolsAbs)
	b.WriteString(`** first when you need names, flags, or usage. Prefer that file over guessing.
- **If** that file is missing, says there are no tools, or the task is clearly unrelated (no local CLI/tars angle), **continue normally** without requiring it.

Regenerate the doc with: `)
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
	b.WriteString("When you begin or the user asks about **local CLIs** / **command-line tools** installed with **tars**:\n\n")
	b.WriteString("1. **If** **")
	b.WriteString(toolsAbs)
	b.WriteString("** exists and the task may involve those tools (or **~/.tars** generally), **read it** for usage and paths.\n")
	b.WriteString("2. **Otherwise** (file missing, empty for this purpose, or question unrelated), **continue as usual** — do not block on this file.\n\n")
	b.WriteString("Ensure **~/.tars/bin** is on PATH when running those binaries. Refresh: `tars connect`.\n\n")
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
