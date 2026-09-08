package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var vaultPath string

var rootCmd = &cobra.Command{
	Use:   "stash",
	Short: "a simple secrets manager",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&vaultPath, "vault", "v", defaultValuePath(), "path to vault file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
