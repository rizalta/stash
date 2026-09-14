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
	ctx    context.Context
	screen screen
	unlock unlockModel
	list   listModel
	vault  *vault.Vault
}

func NewApp(ctx context.Context, vaultPath string) *App {
	return &App{
		ctx:    ctx,
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
		am.list = newListModel(am.ctx, am.vault)
		am.screen = screenList
		return am, am.list.Init()
	}

	switch am.screen {
	case screenUnlock:
		m, cmd := am.unlock.Update(msg)
		am.unlock = m.(unlockModel)
		return am, cmd

	case screenList:
		m, cmd := am.list.Update(msg)
		am.list = m.(listModel)
		return am, cmd
	}

	return am, nil
}

func (am App) View() tea.View {
	switch am.screen {
	case screenUnlock:
		return am.unlock.View()
	case screenList:
		return am.list.View()
	}

	return tea.NewView("")
}

func (am App) Close() error {
	return am.vault.Close()
}
