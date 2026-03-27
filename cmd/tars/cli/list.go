package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"tars/internal/registry"
	"tars/internal/tap"
)

func cmdList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List installed tools or available formulas in taps",
		RunE: func(cmd *cobra.Command, args []string) error {
			avail, _ := cmd.Flags().GetBool("available")
			if avail {
				names, err := tap.ListFormulas()
				if err != nil {
					return err
				}
				sort.Strings(names)
				for _, n := range names {
					fmt.Println(n)
				}
				return nil
			}
			reg, err := registry.Open()
			if err != nil {
				return err
			}
			entries := reg.List()
			sort.Slice(entries, func(i, j int) bool {
				return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
			})
			for _, e := range entries {
				line := fmt.Sprintf("%s %s", e.Name, e.Version)
				if e.Tap != "" {
					line += " [" + e.Tap + "]"
				}
				fmt.Println(line)
			}
			return nil
		},
	}
	c.Flags().BoolP("available", "a", false, "list formulas from taps instead of installed tools")
	return c
}
