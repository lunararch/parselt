package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	editMode mode = iota
	previewMode
)

type keyMap struct {
	quit    key.Binding
	save    key.Binding
	preview key.Binding
	edit    key.Binding
	help    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.save, k.preview, k.edit},
		{k.help, k.quit},
	}
}

var keys = keyMap{
	quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
	save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	preview: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "preview"),
	),
	edit: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "edit"),
	),
	help: key.NewBinding(
		key.WithKeys("ctrl+h", "?"),
		key.WithHelp("ctrl+h", "help"),
	),
}

type model struct {
	filename string
	mode     mode
	width    int
	height   int
	showHelp bool
	content  string
	keys     keyMap
}

func initialModel(filename string) model {
	return model{
		filename: filename,
		mode:     editMode,
		keys:     keys,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.preview):
			m.mode = previewMode
			return m, nil

		case key.Matches(msg, m.keys.edit):
			m.mode = editMode
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	modeText := "EDIT"
	if m.mode == previewMode {
		modeText = "PREVIEW"
	}

	help := "ctrl+s: save • ctrl+p: preview • ctrl+e: edit • ctrl+h: help • ctrl+q: quit"

	if m.showHelp {
		help = m.helpView()
	}

	return fmt.Sprintf("Markdown Editor - %s\n\n%s", modeText, help)
}

func (m model) helpView() string {
	helpText := "Markdown Editor Help\n\n"
	helpText += "Keyboard Shortcuts:\n"
	for _, binding := range m.keys.FullHelp() {
		for _, b := range binding {
			helpText += fmt.Sprintf("  %s: %s\n", b.Keys()[0], b.Help())
		}
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(helpText)
}

func main() {
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	m := initialModel(filename)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}
