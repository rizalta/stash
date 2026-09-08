package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := openVaultFromPrompt()
		if err != nil {
			return err
		}

		entries := v.List()

		if len(entries) == 0 {
			fmt.Println("no secrets stored")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		defer func() {
			_ = w.Flush()
		}()

		_, _ = fmt.Fprintln(w, "ID\tTitle")
		for _, e := range entries {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", e.ID, e.Title)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
