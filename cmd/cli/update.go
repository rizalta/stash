package main

import (
	"errors"
	"fmt"

	"github.com/rizalta/stash/internal/vault"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update secret on vault",
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

		title, err := promptFieldWithDefault("title", e.Title)
		if err != nil {
			return err
		}

		url, err := promptFieldWithDefault("url", e.URL)
		if err != nil {
			return err
		}

		username, err := promptFieldWithDefault("username", e.Username)
		if err != nil {
			return err
		}

		password, err := promptField("password [leave blank to keep unchanged]: ")
		if err != nil {
			return err
		}
		if password == "" {
			password = e.Password
		}

		notes, err := promptFieldWithDefault("notes", e.Notes)
		if err != nil {
			return err
		}

		ue := vault.Entry{
			Title:    title,
			URL:      url,
			Username: username,
			Password: password,
			Notes:    notes,
		}

		if err := v.Update(id, ue); err != nil {
			return fmt.Errorf("updating secret: %w", err)
		}

		fmt.Println("secret updated")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
