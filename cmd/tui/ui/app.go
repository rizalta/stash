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
	width  int
	height int
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		am.width = msg.Width
		am.height = msg.Height

	case vaultOpenedMsg:
		am.vault = msg.vault
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
	content := ""
	switch am.screen {
	case screenUnlock:
		content = am.unlock.View().Content
	case screenList:
		content = am.list.View().Content
	}

	v := tea.NewView(centredView(am.width, am.height, content))
	v.AltScreen = true
	return v
}

func (am App) Close() error {
	if am.vault != nil {
		return am.vault.Close()
	}
	return nil
}
