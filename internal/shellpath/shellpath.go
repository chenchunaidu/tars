package shellpath

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Marker must stay on its own line so we detect existing tars PATH setup.
const Marker = "# tars link PATH (managed by tars; safe to remove)"

// Ensure adds ~/.tars/bin to the login shell config (Unix) or user PATH (Windows).
// home is the user's home directory; binDir must be the absolute path to ~/.tars/bin.
func Ensure(home, binDir string) (summary string, err error) {
	if runtime.GOOS == "windows" {
		return ensureWindowsUserPath(binDir)
	}
	return ensureUnixShellRc(home, binDir)
}

func ensureUnixShellRc(home, binDir string) (string, error) {
	_ = binDir // snippet uses $HOME/.tars/bin for portability
	rc := rcFile(home)
	if strings.Contains(os.Getenv("SHELL"), "fish") {
		if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil {
			return "", err
		}
		snippet := Marker + "\nfish_add_path $HOME/.tars/bin\n"
		changed, err := appendIfMissing(rc, snippet)
		if err != nil {
			return "", err
		}
		if changed {
			return fmt.Sprintf("Added %s to PATH in %s (open a new terminal, or run: source %s)", "$HOME/.tars/bin", rc, rc), nil
		}
		return fmt.Sprintf("PATH already includes ~/.tars/bin in %s", rc), nil
	}

	snippet := Marker + "\nexport PATH=\"$HOME/.tars/bin:$PATH\"\n"
	changed, err := appendIfMissing(rc, snippet)
	if err != nil {
		return "", err
	}
	if changed {
		return fmt.Sprintf("Added %s to PATH in %s (open a new terminal, or run: source %s)", "$HOME/.tars/bin", rc, rc), nil
	}
	return fmt.Sprintf("PATH already configured in %s", rc), nil
}

func rcFile(home string) string {
	sh := os.Getenv("SHELL")
	switch {
	case strings.Contains(sh, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
	case strings.Contains(sh, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(sh, "bash"):
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".bash_profile")
		}
		return filepath.Join(home, ".bashrc")
	default:
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".zshrc")
		}
		return filepath.Join(home, ".bashrc")
	}
}

func appendIfMissing(rcPath, snippet string) (bool, error) {
	data, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if bytes.Contains(data, []byte(Marker)) {
		return false, nil
	}
	var out bytes.Buffer
	out.Write(data)
	if len(data) > 0 && !bytes.HasSuffix(data, []byte{'\n'}) {
		out.WriteByte('\n')
	}
	out.WriteString(snippet)
	return true, os.WriteFile(rcPath, out.Bytes(), 0o644)
}

func ensureWindowsUserPath(binDir string) (string, error) {
	binDir = filepath.Clean(binDir)
	// Single-quoted PowerShell literal; escape ' as ''
	safe := strings.ReplaceAll(binDir, "'", "''")
	ps := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$d = [System.IO.Path]::GetFullPath('%s')
$cur = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($null -eq $cur) { $cur = '' }
$found = $false
foreach ($x in ($cur -split ';')) {
  if ([string]::IsNullOrWhiteSpace($x)) { continue }
  try {
    if ([System.IO.Path]::GetFullPath($x) -ieq $d) { $found = $true; break }
  } catch { }
}
if ($found) { Write-Output 'already'; exit 0 }
$new = if ($cur -eq '') { $d } else { $cur.TrimEnd(';') + ';' + $d }
[Environment]::SetEnvironmentVariable('Path', $new, 'User')
Write-Output 'added'
`, safe)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", ps)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("update Windows user PATH: %w", err)
	}
	s := strings.TrimSpace(string(out))
	if s == "added" {
		return fmt.Sprintf("Added %s to your user PATH (open a new Command Prompt or PowerShell window)", binDir), nil
	}
	return fmt.Sprintf("%s is already on your user PATH", binDir), nil
}
