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
	width     int
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
	case tea.WindowSizeMsg:
		um.width = msg.Width
		um.input.SetWidth(min(maxContentWidth, um.width) - 8)
		return um, nil

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
	title := titleStyle.Render("Unlock your Vault")
	input := inputBoxStyle.Render(um.input.View())

	content := lipgloss.JoinVertical(lipgloss.Left, title, input)

	if um.errMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, errorStyle.Render(um.errMsg))
	}
	content = lipgloss.JoinVertical(lipgloss.Left, content, helpStyle.Render("Enter to unlock · ctrl+c to quit"))

	return tea.NewView(content)
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
