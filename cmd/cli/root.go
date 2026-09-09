package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var vaultPath string

var ErrSecretNotFound = errors.New("secret not found")

var rootCmd = &cobra.Command{
	Use:          "stash",
	Short:        "a simple secrets manager",
	SilenceUsage: true,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(&vaultPath, "vault", "v", defaultValuePath(), "path to vault file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func defaultValuePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "stash.json"
	}

	return filepath.Join(home, ".stash", "vault.json")
}
