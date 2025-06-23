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
	textarea    textarea.Model
	viewport    viewport.Model
	filename    string
	mode        mode
	width       int
	height      int
	showHelp    bool
	content     string
	renderedMD  string
	keys        keyMap
	mdProcessor *SharedMarkdownProcessor
}

type TerminalApp struct {
	model model
}

func NewTerminalApp(filename string) *TerminalApp {
	m := initialModel(filename)
	return &TerminalApp{
		model: m,
	}
}

func (t *TerminalApp) Run() error {
	p := tea.NewProgram(t.model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func initialModel(filename string) model {
	ta := textarea.New()
	ta.Placeholder = "Start writing your markdown ..."
	ta.Focus()

	vp := viewport.New(0, 0)

	m := model{
		textarea:    ta,
		viewport:    vp,
		filename:    filename,
		mode:        editMode,
		keys:        keys,
		mdProcessor: NewSharedMarkdownProcessor(),
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

		m.viewport.Width = msg.Width - 6
		m.viewport.Height = msg.Height - verticalMargins

		if m.mode == previewMode && m.content != "" {
			m.renderedMD = m.RenderMarkdown(m.content)
			m.viewport.SetContent(m.renderedMD)
		}

		if m.mode == previewMode && m.content != "" {
			m.renderedMD = m.RenderMarkdown(m.content)
			m.viewport.SetContent(m.renderedMD)
		}

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

	title := titleStyle.Render("Parselt")
	if m.filename != "" {
		title = titleStyle.Render(fmt.Sprintf("Parselt - %s", filepath.Base(m.filename)))
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
	htmlContent := m.mdProcessor.ConvertMarkdownToHTML(content)
	return m.htmlToTerminal(htmlContent)
}

func (m model) htmlToTerminal(html string) string {
	text := m.mdProcessor.UnescapeHTML(html)

	lines := strings.Split(text, "\n")
	var formatted []string
	var inCodeBlock bool
	var codeBlockContent []string
	var codeBlockLang string

	availableWidth := m.width - 8
	if availableWidth < 40 {
		availableWidth = 40
	}

	h1TagRe := regexp.MustCompile(`<h1[^>]*>`)
	h2TagRe := regexp.MustCompile(`<h2[^>]*>`)
	h3TagRe := regexp.MustCompile(`<h3[^>]*>`)
	h4TagRe := regexp.MustCompile(`<h4[^>]*>`)

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)

		if line == "" {
			if inCodeBlock {
				codeBlockContent = append(codeBlockContent, "")
			} else {
				formatted = append(formatted, "")
			}
			continue
		}

		if strings.Contains(line, "<pre><code") {
			inCodeBlock = true
			codeBlockContent = []string{}
			codeBlockLang = m.mdProcessor.ExtractCodeLanguage(line)
			content := m.mdProcessor.RemoveHTMLTags(line)
			if strings.TrimSpace(content) != "" {
				codeBlockContent = append(codeBlockContent, content)
			}
			continue
		}

		if strings.Contains(line, "</code></pre>") {
			inCodeBlock = false
			content := strings.ReplaceAll(line, "</code></pre>", "")
			content = m.mdProcessor.RemoveHTMLTags(content)
			if strings.TrimSpace(content) != "" {
				codeBlockContent = append(codeBlockContent, content)
			}

			codeHeader := "Code"
			if codeBlockLang != "" {
				codeHeader = strings.ToUpper(codeBlockLang)
			}

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00FF00")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 1)

			blockStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF41")).
				Background(lipgloss.Color("#1a1a1a")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#555555")).
				Padding(1, 2).
				Margin(1, 0)

			formatted = append(formatted, headerStyle.Render("┌─ "+codeHeader+" ─┐"))
			codeContent := m.wrapCodeBlock(codeBlockContent, availableWidth-6)
			formatted = append(formatted, blockStyle.Render(codeContent))
			continue
		}

		if inCodeBlock {
			content := m.mdProcessor.RemoveHTMLTags(line)
			codeBlockContent = append(codeBlockContent, content)
			continue
		}

		if h1TagRe.MatchString(line) {
			if content := m.mdProcessor.ExtractHeaderContent(line, 1); content != "" {
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FF0000")).
					Background(lipgloss.Color("#2A0A0A")).
					Padding(0, 2).
					Render("▶ " + strings.ToUpper(content) + " ◀")
				formatted = append(formatted, styled, "")
			}
		} else if h2TagRe.MatchString(line) {
			if content := m.mdProcessor.ExtractHeaderContent(line, 2); content != "" {
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#00FFFF")).
					Render("▶▶ " + content)
				underline := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#00FFFF")).
					Render(strings.Repeat("═", len(content)+3))
				formatted = append(formatted, styled, underline, "")
			}
		} else if h3TagRe.MatchString(line) {
			if content := m.mdProcessor.ExtractHeaderContent(line, 3); content != "" {
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FFFF00")).
					Render("▶▶▶ " + content)
				formatted = append(formatted, styled)
			}
		} else if h4TagRe.MatchString(line) {
			if content := m.mdProcessor.ExtractHeaderContent(line, 4); content != "" {
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#96CEB4")).
					Render("◦ " + content)
				formatted = append(formatted, styled)
			}
		} else if strings.Contains(line, "<ol>") || strings.Contains(line, "</ol>") {
			continue
		} else if strings.Contains(line, "<ul>") || strings.Contains(line, "</ul>") {
			continue
		} else if strings.Contains(line, "<li>") {
			content := strings.ReplaceAll(line, "<li>", "")
			content = strings.ReplaceAll(content, "</li>", "")

			content = m.processInlineFormatting(content)
			content = m.mdProcessor.RemoveHTMLTags(content)
			content = strings.TrimSpace(content)

			if content != "" {
				leadingSpaces := len(originalLine) - len(strings.TrimLeft(originalLine, " \t"))
				indent := ""
				bullet := "•"

				if leadingSpaces >= 4 {
					indent = "    "
					bullet = "◦"
				} else if leadingSpaces >= 2 {
					indent = "  "
					bullet = "▪"
				}

				if strings.Contains(originalLine, "1.") || strings.Contains(originalLine, "2.") {
					numRe := regexp.MustCompile(`(\d+)\.`)
					if matches := numRe.FindStringSubmatch(content); len(matches) > 1 {
						bullet = matches[1] + "."
						content = numRe.ReplaceAllString(content, "")
						content = strings.TrimSpace(content)
					}
				}

				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFEAA7")).
					Render(indent + bullet + " " + content)
				formatted = append(formatted, styled)
			}
		} else if strings.Contains(line, "<p>") {
			content := strings.ReplaceAll(line, "<p>", "")
			content = strings.ReplaceAll(content, "</p>", "")
			content = m.processInlineFormatting(content)
			content = m.mdProcessor.RemoveHTMLTags(content)
			content = strings.TrimSpace(content)

			if content != "" {
				if len(content) > availableWidth {
					content = m.wrapTextPreservingCode(content, availableWidth, 0)
				}

				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#E6E6E6")).
					Render(content)
				formatted = append(formatted, styled, "")
			}
		} else if strings.Contains(line, "<blockquote>") {
			content := strings.ReplaceAll(line, "<blockquote>", "")
			content = strings.ReplaceAll(content, "</blockquote>", "")
			content = m.processInlineFormatting(content)
			content = m.mdProcessor.RemoveHTMLTags(content)
			content = strings.TrimSpace(content)

			if content != "" {
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#888888")).
					BorderLeft(true).
					BorderStyle(lipgloss.ThickBorder()).
					BorderForeground(lipgloss.Color("#666666")).
					PaddingLeft(2).
					Italic(true).
					Render("❝ " + content)
				formatted = append(formatted, styled)
			}
		} else {
			content := m.processInlineFormatting(line)
			cleanLine := m.mdProcessor.RemoveHTMLTags(content)
			cleanLine = strings.TrimSpace(cleanLine)
			if cleanLine != "" {
				if len(cleanLine) > availableWidth {
					cleanLine = m.wrapTextPreservingCode(cleanLine, availableWidth, 0)
				}

				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#CCCCCC")).
					Render(cleanLine)
				formatted = append(formatted, styled)
			}
		}
	}
	return strings.Join(formatted, "\n")
}

func (m model) wrapText(text string, width int, indent int) string {
	if width <= 0 {
		return text
	}

	if strings.Contains(text, "\x1b[") {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string
	indentStr := strings.Repeat(" ", indent)

	for i, word := range words {
		if strings.Contains(word, "`") && strings.Count(currentLine+" "+word, "`")%2 == 1 {
			if currentLine != "" {
				currentLine += " " + word
			} else {
				currentLine = word
			}
			continue
		}

		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		linePrefix := ""
		if i == 0 || currentLine == "" {
			linePrefix = ""
		} else {
			linePrefix = indentStr
		}

		if len(linePrefix+testLine) <= width {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, linePrefix+currentLine)
				currentLine = word
			} else {
				lines = append(lines, linePrefix+word)
				currentLine = ""
			}
		}
	}

	if currentLine != "" {
		linePrefix := indentStr
		if len(lines) == 0 {
			linePrefix = ""
		}
		lines = append(lines, linePrefix+currentLine)
	}
	return strings.Join(lines, "\n")
}

func (m model) wrapCodeBlock(codeLines []string, maxWidth int) string {
	if maxWidth <= 20 {
		return strings.Join(codeLines, "\n")
	}

	var wrappedLines []string

	for _, line := range codeLines {
		if len(line) <= maxWidth {
			wrappedLines = append(wrappedLines, line)
		} else {
			wrapped := m.wrapCodeLine(line, maxWidth)
			wrappedLines = append(wrappedLines, wrapped...)
		}
	}

	return strings.Join(wrappedLines, "\n")
}

func (m model) wrapCodeLine(line string, maxWidth int) []string {
	if len(line) <= maxWidth {
		return []string{line}
	}

	leadingSpaces := 0
	for _, char := range line {
		if char == ' ' || char == '\t' {
			if char == '\t' {
				leadingSpaces += 4
			} else {
				leadingSpaces++
			}
		} else {
			break
		}
	}

	contIndent := strings.Repeat(" ", leadingSpaces+2)

	var lines []string
	remaining := line

	for len(remaining) > maxWidth {
		breakPoint := m.findCodeBreakPoint(remaining, maxWidth)

		if breakPoint <= leadingSpaces {
			breakPoint = maxWidth - 3 // Leave room for "..."
			lines = append(lines, remaining[:breakPoint]+"...")
			remaining = contIndent + "..." + remaining[breakPoint:]
		} else {
			lines = append(lines, remaining[:breakPoint])
			remaining = contIndent + strings.TrimLeft(remaining[breakPoint:], " ")
		}

		maxWidth = maxWidth - len(contIndent)
		if maxWidth < 20 {
			lines = append(lines, remaining)
			break
		}
	}

	if remaining != "" {
		lines = append(lines, remaining)
	}

	return lines
}

func (m model) findCodeBreakPoint(line string, maxWidth int) int {
	if maxWidth >= len(line) {
		return len(line)
	}

	breakChars := []string{" ", ",", ";", ".", ")", "}", "]", ">", "|", "&", "+", "-", "="}

	for i := maxWidth - 1; i > maxWidth/2; i-- {
		if i >= len(line) {
			continue
		}

		char := string(line[i])
		for _, breakChar := range breakChars {
			if char == breakChar {
				if breakChar == " " {
					return i // Break before space
				} else {
					return i + 1 // Break after punctuation
				}
			}
		}
	}
	return maxWidth / 2
}

func (m model) wrapTextPreservingCode(text string, width int, indent int) string {
	if width <= 0 {
		return text
	}

	if strings.Contains(text, "`") {
		return m.wrapTextWithInlineCode(text, width, indent)
	}

	return m.wrapText(text, width, indent)
}

func (m model) wrapTextWithInlineCode(text string, width int, indent int) string {
	if width <= 0 {
		return text
	}

	segments := m.splitTextPreservingCode(text)

	var lines []string
	var currentLine string
	indentStr := strings.Repeat(" ", indent)

	for i, segment := range segments {
		isCode := strings.HasPrefix(segment.text, "`") && strings.HasSuffix(segment.text, "`")

		testLine := currentLine
		if testLine != "" && !isCode {
			testLine += " "
		}
		testLine += segment.text

		linePrefix := ""
		if i == 0 || currentLine == "" {
			linePrefix = ""
		} else {
			linePrefix = indentStr
		}

		if len(linePrefix+testLine) <= width || isCode {
			if currentLine == "" {
				currentLine = segment.text
			} else if isCode {
				currentLine += segment.text // No space before inline code
			} else {
				currentLine += " " + segment.text
			}
		} else {
			// Line too long, break here
			if currentLine != "" {
				lines = append(lines, linePrefix+currentLine)
				currentLine = segment.text
			} else {
				// Single segment longer than width
				lines = append(lines, linePrefix+segment.text)
				currentLine = ""
			}
		}
	}

	if currentLine != "" {
		linePrefix := indentStr
		if len(lines) == 0 {
			linePrefix = ""
		}
		lines = append(lines, linePrefix+currentLine)
	}

	return strings.Join(lines, "\n")
}

type textSegment struct {
	text   string
	isCode bool
}

func (m model) splitTextPreservingCode(text string) []textSegment {
	var segments []textSegment
	var current strings.Builder
	inCode := false

	for _, char := range text {
		if char == '`' {
			if inCode {
				current.WriteRune(char)
				segments = append(segments, textSegment{
					text:   current.String(),
					isCode: true,
				})
				current.Reset()
				inCode = false
			} else {
				if current.Len() > 0 {
					segments = append(segments, textSegment{
						text:   current.String(),
						isCode: false,
					})
					current.Reset()
				}
				current.WriteRune(char)
				inCode = true
			}
		} else if char == ' ' && !inCode {
			if current.Len() > 0 {
				segments = append(segments, textSegment{
					text:   current.String(),
					isCode: false,
				})
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		segments = append(segments, textSegment{
			text:   current.String(),
			isCode: inCode,
		})
	}

	return segments
}

func (m model) processInlineFormatting(content string) string {
	codeRe := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)
	content = codeRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			codeContent := strings.TrimSpace(matches[1])
			styled := lipgloss.NewStyle().
				Background(lipgloss.Color("#333333")).
				Foreground(lipgloss.Color("#00FF00")).
				Render("`" + codeContent + "`")
			return styled
		}
		return match
	})

	strongRe := regexp.MustCompile(`<strong[^>]*>(.*?)</strong>`)
	content = strongRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := strongRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			boldContent := strings.TrimSpace(matches[1])
			styled := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render(boldContent)
			return styled
		}
		return match
	})

	emRe := regexp.MustCompile(`<em[^>]*>(.*?)</em>`)
	content = emRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := emRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			italicContent := strings.TrimSpace(matches[1])
			styled := lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("#DDDDDD")).
				Render(italicContent)
			return styled
		}
		return match
	})

	return content
}
