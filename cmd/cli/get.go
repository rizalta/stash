package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rizalta/stash/internal/vault"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get secret from vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		v, err := openVaultFromPrompt()
		if err != nil {
			return err
		}

		e, err := v.Get(id)
		if err != nil {
			if errors.Is(err, vault.ErrEntryNotFound) {
				return ErrSecretNotFound
			}
			return fmt.Errorf("getting secret: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		defer func() {
			_ = w.Flush()
		}()

		_, _ = fmt.Fprintf(w, "Title:\t%s\n", e.Title)
		_, _ = fmt.Fprintf(w, "URL:\t%s\n", e.URL)
		_, _ = fmt.Fprintf(w, "Username:\t%s\n", e.Username)
		_, _ = fmt.Fprintf(w, "Password:\t%s\n", e.Password)
		_, _ = fmt.Fprintf(w, "Notes:\t%s\n", e.Notes)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
