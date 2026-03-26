package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-tools/internal/tap"
)

func cmdTap() *cobra.Command {
	root := &cobra.Command{
		Use:   "tap",
		Short: "Manage extra formula taps (core is implicit; see AGENT_TOOLS_CORE_URL)",
	}
	root.AddCommand(
		&cobra.Command{
			Use:   "add <name> <git-url>",
			Short: "Clone a tap repository (publishers host Formulas/*.json here)",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := tap.Add(args[0], args[1]); err != nil {
					return err
				}
				fmt.Printf("Tapped %s\n", args[0])
				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List core + registered taps",
			RunE: func(cmd *cobra.Command, args []string) error {
				taps, err := tap.DescribeTaps()
				if err != nil {
					return err
				}
				for _, t := range taps {
					fmt.Printf("%s\t%s\n", t.Name, t.URL)
				}
				return nil
			},
		},
	)
	return root
}
