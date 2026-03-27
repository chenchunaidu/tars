package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"tars/internal/core"
	"tars/internal/tap"
)

func cmdUpdate() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Fetch latest formula definitions (core + tapped repos)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := core.Update(); err != nil {
				return err
			}
			if err := tap.PullUserTaps(); err != nil {
				return err
			}
			fmt.Println("==> Formula sources updated.")
			return nil
		},
	}
}
