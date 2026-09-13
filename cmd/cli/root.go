package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	rootCmd.PersistentFlags().StringVarP(&vaultPath, "vault", "v", defaultVaultPath(), "path to vault file")
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func defaultVaultPath() string {
	path := filepath.Join(".stash", "vault.db")
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, path)
}
