package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/rizalta/stash/internal/vault"
)

type screen int

const (
	screenUnlock = iota
	screenList
)

type App struct {
	screen screen
	unlock unlockModel
	vault  *vault.Vault
}

func NewApp(ctx context.Context, vaultPath string) *App {
	return &App{
		screen: screenUnlock,
		unlock: newUnlockModel(ctx, vaultPath),
	}
}

func (am App) Init() tea.Cmd {
	return am.unlock.Init()
}

func (am App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if opened, ok := msg.(vaultOpenedMsg); ok {
		am.vault = opened.vault
		am.screen = screenList
	}

	switch am.screen {
	case screenUnlock:
		m, cmd := am.unlock.Update(msg)
		am.unlock = m.(unlockModel)
		return am, cmd

	case screenList:
		return am, nil
	}

	return am, nil
}

func (am App) View() tea.View {
	switch am.screen {
	case screenUnlock:
		return am.unlock.View()
	case screenList:
		return tea.NewView("unlocked")
	}

	return tea.NewView("")
}
