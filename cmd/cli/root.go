package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/rizalta/stash/internal/vault"
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
	rootCmd.PersistentFlags().StringVarP(&vaultPath, "vault", "v", vault.DefaultPath(), "path to vault file")
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
