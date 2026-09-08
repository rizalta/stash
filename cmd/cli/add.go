package main

import (
	"fmt"

	"github.com/rizalta/stash/internal/vault"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add secret to vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := openVaultFromPrompt()
		if err != nil {
			return err
		}

		title, err := promptField("title: ")
		if err != nil {
			return err
		}

		url, err := promptField("url: ")
		if err != nil {
			return err
		}

		username, err := promptField("username: ")
		if err != nil {
			return err
		}

		password, err := promptField("password: ")
		if err != nil {
			return err
		}

		e := vault.Entry{
			Title:    title,
			URL:      url,
			Username: username,
			Password: password,
		}

		id, err := v.Add(e)
		if err != nil {
			return fmt.Errorf("adding entry: %w", err)
		}

		fmt.Printf("added secret id: %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
