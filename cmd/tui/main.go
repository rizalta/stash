package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/rizalta/stash/cmd/tui/ui"
	"github.com/rizalta/stash/internal/vault"
)

func main() {
	var vaultPath string
	def := vault.DefaultPath()

	flag.StringVar(&vaultPath, "v", def, "path to vault file")
	flag.StringVar(&vaultPath, "vault", def, "path to vault file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p := tea.NewProgram(ui.NewApp(ctx, vaultPath))
	if _, err := p.Run(); err != nil {
		log.Fatalf("error running the program: %v", err)
	}
}
