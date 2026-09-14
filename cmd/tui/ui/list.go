package ui

import (
	"context"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/rizalta/stash/internal/storage"
	"github.com/rizalta/stash/internal/vault"
)

type listModel struct {
	ctx    context.Context
	vault  *vault.Vault
	list   list.Model
	errMsg string
}

type entriesLoadingFailedMsg struct {
	err error
}

type entriesLoadedMsg struct {
	entries []storage.EntryMeta
}

type entryItem struct {
	meta storage.EntryMeta
}

func (e entryItem) FilterValue() string { return e.meta.Title }
func (e entryItem) Title() string       { return e.meta.Title }
func (e entryItem) Description() string { return "" }

func newListModel(ctx context.Context, v *vault.Vault) listModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	l := list.New([]list.Item{}, delegate, 40, 20)
	l.Title = "stash"

	return listModel{
		ctx:   ctx,
		vault: v,
		list:  l,
	}
}

func (lm listModel) Init() tea.Cmd {
	return lm.loadEntries()
}

func (lm listModel) loadEntries() tea.Cmd {
	return func() tea.Msg {
		entries, err := lm.vault.List(lm.ctx)
		if err != nil {
			return entriesLoadingFailedMsg{err}
		}
		return entriesLoadedMsg{entries}
	}
}

func (lm listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case entriesLoadingFailedMsg:
		lm.errMsg = "error loading secrets: " + msg.err.Error()
		return lm, nil

	case entriesLoadedMsg:
		items := make([]list.Item, len(msg.entries))
		for i, e := range msg.entries {
			items[i] = entryItem{e}
		}
		lm.errMsg = ""

		cmd := lm.list.SetItems(items)
		return lm, cmd

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return lm, tea.Quit
		}
	}

	var cmd tea.Cmd
	lm.list, cmd = lm.list.Update(msg)

	return lm, cmd
}

func (lm listModel) View() tea.View {
	var content string
	if lm.errMsg != "" {
		content = lm.errMsg
	} else {
		content = lm.list.View()
	}

	v := tea.NewView(content)
	v.AltScreen = true

	return v
}
