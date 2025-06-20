package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
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
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2)

	editorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#04B575")).
			Padding(0, 1)
)

type model struct {
	textarea   textarea.Model
	viewport   viewport.Model
	filename   string
	mode       mode
	width      int
	height     int
	showHelp   bool
	content    string
	renderedMD string
	keys       keyMap
}

func initialModel(filename string) model {
	ta := textarea.New()
	ta.Placeholder = "Start writing your markdown ..."
	ta.Focus()

	vp := viewport.New(0, 0)

	m := model{
		textarea: ta,
		viewport: vp,
		filename: filename,
		mode:     editMode,
		keys:     keys,
	}

	if filename != "" {
		if content, err := os.ReadFile(filename); err == nil {
			m.content = string(content)
			m.textarea.SetValue(m.content)
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 3
		verticalMargins := headerHeight + footerHeight

		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - verticalMargins)

		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - verticalMargins

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.save):
			return m, m.saveFile()

		case key.Matches(msg, m.keys.preview):
			m.mode = previewMode
			m.content = m.textarea.Value()
			m.renderedMD = m.RenderMarkdown(m.content)
			m.viewport.SetContent(m.renderedMD)
			return m, nil

		case key.Matches(msg, m.keys.edit):
			m.mode = editMode
			m.textarea.Focus()
			return m, nil

		case key.Matches(msg, m.keys.help):
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	if m.mode == editMode {
		m.textarea, tiCmd = m.textarea.Update(msg)
	} else {
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd)
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

	if m.mode == editMode {
		content = editorStyle.Render(m.textarea.View())
	} else {
		content = previewStyle.Render(m.viewport.View())
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

func (m model) saveFile() tea.Cmd {
	return func() tea.Msg {
		content := m.textarea.Value()

		filename := m.filename
		if filename == "" {
			filename = "untitled.md"
		}

		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %v", err)
		}

		return fmt.Sprintf("Saved to %s", filename)
	}
}

func (m model) RenderMarkdown(content string) string {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)

	var buf strings.Builder
	if err := md.Convert([]byte(content), &buf); err != nil {
		return content
	}

	htmlOutput := buf.String()
	fmt.Printf("DEBUG - HTML output:\n%s\n", htmlOutput)

	return m.htmlToTerminal(buf.String())
}

func (m model) htmlToTerminal(html string) string {
	// First, let's handle HTML entities
	text := strings.ReplaceAll(html, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", `"`)
	text = strings.ReplaceAll(text, "&#39;", "'")

	lines := strings.Split(text, "\n")
	var formatted []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			formatted = append(formatted, "")
			continue
		}

		// Process HTML tags and apply styling - use regex to handle attributes
		if matched, _ := regexp.MatchString(`<h1[^>]*>`, line); matched {
			// Extract content between h1 tags
			re := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				content := strings.TrimSpace(matches[1])
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FF0000")).
					Render("▶ " + strings.ToUpper(content))
				formatted = append(formatted, styled)
			}
		} else if matched, _ := regexp.MatchString(`<h2[^>]*>`, line); matched {
			re := regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				content := strings.TrimSpace(matches[1])
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#00FFFF")).
					Render("▶▶ " + content)
				formatted = append(formatted, styled)
			}
		} else if matched, _ := regexp.MatchString(`<h3[^>]*>`, line); matched {
			re := regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				content := strings.TrimSpace(matches[1])
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FFFF00")).
					Render("▶▶▶ " + content)
				formatted = append(formatted, styled)
			}
		} else if strings.Contains(line, "<p>") {
			content := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line, "<p>", ""), "</p>", ""))
			if content != "" {
				// Regular paragraph text - keep it simple and readable
				formatted = append(formatted, content)
			}
		} else {
			// Strip any remaining HTML tags for fallback
			re := regexp.MustCompile(`<[^>]*>`)
			cleanLine := re.ReplaceAllString(line, "")
			cleanLine = strings.TrimSpace(cleanLine)
			if cleanLine != "" {
				formatted = append(formatted, cleanLine)
			}
		}
	}

	return strings.Join(formatted, "\n")
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
