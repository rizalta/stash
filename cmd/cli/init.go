package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rizalta/stash/internal/vault"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create a new vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := filepath.Dir(vaultPath)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("creating vault directory: %w", err)
		}

		password, err := promptNewPassword()
		if err != nil {
			return err
		}

		if _, err := vault.Create(vaultPath, password); err != nil {
			return fmt.Errorf("creating vault: %w", err)
		}

		fmt.Printf("vault created at %s\n", vaultPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
