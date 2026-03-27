package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"tars/internal/security"
)

func cmdHash() *cobra.Command {
	return &cobra.Command{
		Use:   "hash <file>",
		Short: "Print SHA256 of a file (for filling formula sha256 fields)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := security.FileSHA256(args[0])
			if err != nil {
				return err
			}
			fmt.Println(h)
			return nil
		},
	}
}
