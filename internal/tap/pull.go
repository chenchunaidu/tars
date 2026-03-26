package tap

import (
	"fmt"
	"os"
	"os/exec"
)

// PullUserTaps runs git pull --ff-only in each user-registered tap directory.
func PullUserTaps() error {
	taps, err := LoadList()
	if err != nil {
		return err
	}
	for _, t := range taps {
		cmd := exec.Command("git", "-C", t.Path, "pull", "--ff-only")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("tap %q: git pull: %w", t.Name, err)
		}
	}
	return nil
}
