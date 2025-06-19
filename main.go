package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "help"),
	),
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Background(lipgloss.Color("#FFFDF5"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	previewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Background(lipgloss.Color("#874BFD")).
			Padding(1, 2)

	editorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Background(lipgloss.Color("#04B575")).
			Padding(0, 1)
)

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
	var content string

	title := titleStyle.Render("Markdown Editor")
	if m.filename != "" {
		title = titleStyle.Render(fmt.Sprintf("Markdown Editor - %s", filepath.Base(m.filename)))
	}

	modeText := "EDIT"
	if m.mode == previewMode {
		modeText = "PREVIEW"
	}
	status := statusStyle.Render(fmt.Sprintf(" %s ", modeText))

	header := lipgloss.JoinHorizontal(lipgloss.Left, title, " ", status)

	// Content area (placeholder)
	if m.mode == editMode {
		content = editorStyle.Render("Editor placeholder")
	} else {
		content = previewStyle.Render("Preview placeholder")
	}

	help := helpStyle.Render("ctrl+s: save • ctrl+p: preview • ctrl+e: edit • ctrl+h: help • ctrl+q: quit")
	if m.showHelp {
		help = m.helpView()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		help,
	)
}

func (m model) helpView() string {
	help := `
Keyboard Shortcuts:

  ctrl+s    Save file
  ctrl+p    Switch to preview mode
  ctrl+e    Switch to edit mode  
  ctrl+h    Toggle this help
  ctrl+q    Quit application

Navigation (Preview Mode):
  ↑/k       Scroll up
  ↓/j       Scroll down
  g         Go to top
  G         Go to bottom
  
Edit Mode:
  All standard text editing shortcuts work
  Tab       Insert 2 spaces (for indentation)
`
	return helpStyle.Render(help)
}

func main() {
	var filename string

	if len(os.Args) > 1 {
		filename = os.Args[1]

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			file, err := os.Create(filename)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				os.Exit(1)
			}
			file.Close()
		}
	}

	m := initialModel(filename)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}
