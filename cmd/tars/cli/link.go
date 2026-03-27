package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"tars/internal/paths"
	"tars/internal/shellpath"
)

func cmdLink() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Install tars on your PATH (symlink + shell / Windows user PATH)",
		Long: `Puts the tars command on your PATH for new terminals:

  1) Symlinks this executable to ~/.tars/bin/tars (tars.exe on Windows).
  2) Adds ~/.tars/bin to your shell startup file (bash/zsh/fish) or Windows user PATH.

Open a new terminal after running, or source your rc file. On Windows, symlink creation
may require Developer Mode or an elevated shell; PATH is still updated if the symlink fails.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			exe, err = filepath.EvalSymlinks(exe)
			if err != nil {
				return err
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			bin, err := paths.Bin()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(bin, 0o755); err != nil {
				return err
			}
			dstName := "tars"
			if runtime.GOOS == "windows" {
				dstName = "tars.exe"
			}
			dst := filepath.Join(bin, dstName)
			_ = os.Remove(dst)
			if err := os.Symlink(exe, dst); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not symlink %s -> %s: %v\n", dst, exe, err)
				fmt.Fprintln(os.Stderr, "  (PATH will still be updated; fix symlink or run tars from its folder.)")
			} else {
				fmt.Printf("Linked %s -> %s\n", dst, exe)
			}

			summary, err := shellpath.Ensure(home, bin)
			if err != nil {
				return err
			}
			fmt.Println(summary)
			if runtime.GOOS != "windows" {
				fmt.Printf("\nUse tars in this terminal now: export PATH=%q:$PATH\n", bin)
			}
			return nil
		},
	}
}
