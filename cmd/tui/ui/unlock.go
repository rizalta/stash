package ui

import (
	"context"
	"errors"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/rizalta/stash/internal/vault"
)

type vaultOpenedMsg struct {
	vault *vault.Vault
}

type vaultOpenFailedMsg struct {
	err error
}

type unlockModel struct {
	ctx       context.Context
	vaultPath string
	input     textinput.Model
	errMsg    string
}

func newUnlockModel(ctx context.Context, vaultPath string) unlockModel {
	ti := textinput.New()
	ti.Placeholder = "master password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	return unlockModel{
		ctx:       ctx,
		vaultPath: vaultPath,
		input:     ti,
	}
}

func (um unlockModel) Init() tea.Cmd {
	return textinput.Blink
}

func (um unlockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return um, tea.Quit

		case "enter":
			password := []byte(um.input.Value())
			um.errMsg = ""
			return um, attemptOpen(um.ctx, um.vaultPath, password)

		default:
			um.input, cmd = um.input.Update(msg)
			return um, cmd
		}

	case vaultOpenFailedMsg:
		if errors.Is(msg.err, vault.ErrWrongPassword) {
			um.errMsg = "incorrect password"
		} else {
			um.errMsg = "error: " + msg.err.Error()
		}
		um.input.SetValue("")
		return um, nil

	case vaultOpenedMsg:
		return um, nil
	}

	um.input, cmd = um.input.Update(msg)

	return um, cmd
}

func (um unlockModel) View() tea.View {
	content := lipgloss.JoinVertical(lipgloss.Left, "Unlock your Vault", um.input.View())
	if um.errMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, um.errMsg)
	}
	v := tea.NewView(content)
	v.AltScreen = true

	return v
}

func attemptOpen(ctx context.Context, vaultPath string, password []byte) tea.Cmd {
	return func() tea.Msg {
		v, err := vault.Open(ctx, vaultPath, password)
		if err != nil {
			return vaultOpenFailedMsg{err}
		}

		return vaultOpenedMsg{v}
	}
}
