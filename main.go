package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	editMode mode = iota
	previewMode
)

type model struct {
	filename string
	mode     mode
	width    int
	height   int
	content  string
}

func initialModel(filename string) model {
	return model{
		filename: filename,
		mode:     editMode,
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
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Markdown Editor - %dx%d\nPress q to quit", m.width, m.height)
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
