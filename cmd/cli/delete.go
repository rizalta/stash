package main

import (
	"errors"
	"fmt"

	"github.com/rizalta/stash/internal/vault"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete secret from vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		v, err := openVaultFromPrompt()
		if err != nil {
			return err
		}

		if err := v.Delete(id); err != nil {
			if errors.Is(err, vault.ErrEntryNotFound) {
				return ErrSecretNotFound
			}

			return fmt.Errorf("deleting secret: %w", err)
		}

		fmt.Println("secret deleted")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
