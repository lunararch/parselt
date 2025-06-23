package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUIApp struct {
	app         fyne.App
	window      fyne.Window
	editor      *widget.Entry
	preview     *widget.RichText
	currentFile string
	fileLabel   *widget.Label
	splitPanel  *container.Split

	model       model
	mdProcessor *SharedMarkdownProcessor
}

func NewGUIApp() *GUIApp {
	myApp := app.NewWithID("com.parselt.editor")

	myApp.SetIcon(resourceParseltIconPng)

	myWindow := myApp.NewWindow("Parselt - Markdown Editor")
	myWindow.Resize(fyne.NewSize(1200, 800))
	myWindow.CenterOnScreen()

	myWindow.SetIcon(resourceParseltIconPng)

	m := initialModel("")

	return &GUIApp{
		app:         myApp,
		window:      myWindow,
		model:       m,
		mdProcessor: NewSharedMarkdownProcessor(),
	}
}

func (g *GUIApp) setupUI() {
	g.editor = widget.NewMultiLineEntry()
	g.editor.Wrapping = fyne.TextWrapWord
	g.editor.SetPlaceHolder("Start writing your markdown...")

	g.editor.TextStyle = fyne.TextStyle{
		Monospace: true,
	}

	g.preview = widget.NewRichText()
	g.preview.Wrapping = fyne.TextWrapWord

	g.fileLabel = widget.NewLabel("untitled.md")
	g.fileLabel.TextStyle = fyne.TextStyle{
		Bold: true,
	}

	editorContainer := container.NewBorder(
		widget.NewCard("Editor", "", nil), nil, nil, nil,
		container.NewScroll(g.editor),
	)

	previewContainer := container.NewBorder(
		widget.NewCard("Preview", "", nil), nil, nil, nil,
		container.NewScroll(g.preview),
	)

	g.splitPanel = container.NewHSplit(editorContainer, previewContainer)
	g.splitPanel.SetOffset(0.5)

	content := container.NewBorder(
		nil,
		g.fileLabel,
		nil,
		nil,
		g.splitPanel,
	)

	g.window.SetContent(content)
	g.setupEventHandlers()
	g.setupMenu()
}

func (g *GUIApp) setupMenu() {
	newItem := fyne.NewMenuItem("New", g.newFile)
	newItem.Icon = theme.DocumentCreateIcon()

	openItem := fyne.NewMenuItem("Open", g.openFile)
	openItem.Icon = theme.FolderOpenIcon()

	saveItem := fyne.NewMenuItem("Save", g.saveFile)
	saveItem.Icon = theme.DocumentSaveIcon()

	saveAsItem := fyne.NewMenuItem("Save As...", g.saveAsFile)

	quitItem := fyne.NewMenuItem("Quit", func() {
		g.app.Quit()
	})

	fileMenu := fyne.NewMenu("File", newItem, openItem, fyne.NewMenuItemSeparator(),
		saveItem, saveAsItem, fyne.NewMenuItemSeparator(), quitItem)

	toggleViewItem := fyne.NewMenuItem("Toggle Split View", g.toggleView)
	editorOnlyItem := fyne.NewMenuItem("Editor Only", func() {
		g.splitPanel.SetOffset(1.0)
	})
	previewOnlyItem := fyne.NewMenuItem("Preview Only", func() {
		g.splitPanel.SetOffset(0.0)
	})
	splitViewItem := fyne.NewMenuItem("Split View", func() {
		g.splitPanel.SetOffset(0.5)
	})

	viewMenu := fyne.NewMenu("View", toggleViewItem, fyne.NewMenuItemSeparator(),
		editorOnlyItem, previewOnlyItem, splitViewItem)

	aboutItem := fyne.NewMenuItem("About", g.showAbout)
	helpMenu := fyne.NewMenu("Help", aboutItem)

	mainMenu := fyne.NewMainMenu(fileMenu, viewMenu, helpMenu)
	g.window.SetMainMenu(mainMenu)
}

func (g *GUIApp) setupEventHandlers() {
	g.editor.OnChanged = func(content string) {
		g.updatePreview(content)
	}

	g.window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyS && (key.Physical.ScanCode == 0 || key.Physical.ScanCode == 1) {
			g.saveFile()
		}
	})
}

func (g *GUIApp) updatePreview(content string) {
	if content == "" {
		g.preview.ParseMarkdown("")
		return
	}

	htmlContent := g.mdProcessor.ConvertMarkdownToHTML(content)
	markdownForFyne := g.htmlToMarkdown(htmlContent)
	g.preview.ParseMarkdown(markdownForFyne)
}

func (g *GUIApp) htmlToMarkdown(html string) string {
	text := g.mdProcessor.UnescapeHTML(html)

	lines := strings.Split(text, "\n")
	var result []string
	var inCodeBlock bool
	var codeBlockContent []string
	var codeBlockLang string

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		if line == "" {
			if inCodeBlock {
				codeBlockContent = append(codeBlockContent, "")
			} else {
				result = append(result, "")
			}
			continue
		}

		if strings.Contains(line, "<pre><code") {
			inCodeBlock = true
			codeBlockContent = []string{}
			codeBlockLang = g.mdProcessor.ExtractCodeLanguage(line)
			content := g.mdProcessor.RemoveHTMLTags(line)
			if strings.TrimSpace(content) != "" {
				codeBlockContent = append(codeBlockContent, content)
			}
			continue
		}

		if strings.Contains(line, "</code></pre>") {
			inCodeBlock = false
			content := strings.ReplaceAll(line, "</code></pre>", "")
			content = g.mdProcessor.RemoveHTMLTags(content)
			if strings.TrimSpace(content) != "" {
				codeBlockContent = append(codeBlockContent, content)
			}

			if codeBlockLang != "" {
				result = append(result, "```"+codeBlockLang)
			} else {
				result = append(result, "```")
			}
			result = append(result, strings.Join(codeBlockContent, "\n"))
			result = append(result, "```")
			result = append(result, "")
			continue
		}

		if inCodeBlock {
			content := g.mdProcessor.RemoveHTMLTags(line)
			codeBlockContent = append(codeBlockContent, content)
			continue
		}

		if strings.Contains(line, "<h1") {
			if content := g.mdProcessor.ExtractHeaderContent(line, 1); content != "" {
				result = append(result, "# "+content)
				result = append(result, "")
			}
		} else if strings.Contains(line, "<h2") {
			if content := g.mdProcessor.ExtractHeaderContent(line, 2); content != "" {
				result = append(result, "## "+content)
				result = append(result, "")
			}
		} else if strings.Contains(line, "<h3") {
			if content := g.mdProcessor.ExtractHeaderContent(line, 3); content != "" {
				result = append(result, "### "+content)
				result = append(result, "")
			}
		} else if strings.Contains(line, "<h4") {
			if content := g.mdProcessor.ExtractHeaderContent(line, 4); content != "" {
				result = append(result, "#### "+content)
				result = append(result, "")
			}
		} else if strings.Contains(line, "<ol>") || strings.Contains(line, "</ol>") {
			continue
		} else if strings.Contains(line, "<ul>") || strings.Contains(line, "</ul>") {
			continue
		} else if strings.Contains(line, "<li>") {
			content := strings.ReplaceAll(line, "<li>", "")
			content = strings.ReplaceAll(content, "</li>", "")

			content = g.processInlineFormatting(content)
			content = g.mdProcessor.RemoveHTMLTags(content)
			content = strings.TrimSpace(content)

			if content != "" {
				leadingSpaces := len(originalLine) - len(strings.TrimLeft(originalLine, " \t"))
				indent := ""
				bullet := "- "

				if leadingSpaces >= 4 {
					indent = "    "
				} else if leadingSpaces >= 2 {
					indent = "  "
				}

				numRe := regexp.MustCompile(`(\d+)\.`)
				if matches := numRe.FindStringSubmatch(content); len(matches) > 1 {
					bullet = matches[1] + ". "
					content = numRe.ReplaceAllString(content, "")
					content = strings.TrimSpace(content)
				}

				result = append(result, indent+bullet+content)
			}
		} else if strings.Contains(line, "<p>") {
			content := strings.ReplaceAll(line, "<p>", "")
			content = strings.ReplaceAll(content, "</p>", "")
			content = g.processInlineFormatting(content)
			content = g.mdProcessor.RemoveHTMLTags(content)
			content = strings.TrimSpace(content)

			if content != "" {
				result = append(result, content)
				result = append(result, "")
			}
		} else if strings.Contains(line, "<blockquote>") {
			content := strings.ReplaceAll(line, "<blockquote>", "")
			content = strings.ReplaceAll(content, "</blockquote>", "")
			content = g.processInlineFormatting(content)
			content = g.mdProcessor.RemoveHTMLTags(content)
			content = strings.TrimSpace(content)

			if content != "" {
				result = append(result, "> "+content)
			}
		} else {
			content := g.processInlineFormatting(line)
			cleanLine := g.mdProcessor.RemoveHTMLTags(content)
			cleanLine = strings.TrimSpace(cleanLine)
			if cleanLine != "" {
				result = append(result, cleanLine)
			}
		}
	}

	var cleanResult []string
	for i, line := range result {
		if line == "" && i > 0 && i < len(result)-1 && result[i-1] == "" {
			continue // Skip consecutive empty lines
		}
		cleanResult = append(cleanResult, line)
	}
	return strings.Join(cleanResult, "\n")
}

func (g *GUIApp) processInlineFormatting(content string) string {
	codeRe := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)
	content = codeRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			codeContent := strings.TrimSpace(matches[1])
			return "`" + codeContent + "`"
		}
		return match
	})

	strongRe := regexp.MustCompile(`<strong[^>]*>(.*?)</strong>`)
	content = strongRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := strongRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			boldContent := strings.TrimSpace(matches[1])
			return "**" + boldContent + "**"
		}
		return match
	})

	emRe := regexp.MustCompile(`<em[^>]*>(.*?)</em>`)
	content = emRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := emRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			italicContent := strings.TrimSpace(matches[1])
			return "*" + italicContent + "*"
		}
		return match
	})

	return content
}

func (g *GUIApp) newFile() {
	g.editor.SetText("")
	g.currentFile = ""
	g.fileLabel.SetText("untitled.md")
	g.window.SetTitle("Parselt - Markdown Editor")
}

func (g *GUIApp) openFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}

		g.editor.SetText(string(data))
		g.currentFile = reader.URI().Path()
		g.fileLabel.SetText(filepath.Base(g.currentFile))
		g.window.SetTitle(fmt.Sprintf("Parselt - %s", filepath.Base(g.currentFile)))
	}, g.window)
}

func (g *GUIApp) saveFile() {
	if g.currentFile == "" {
		g.saveAsFile()
		return
	}

	err := os.WriteFile(g.currentFile, []byte(g.editor.Text), 0644)
	if err != nil {
		dialog.ShowError(err, g.window)
		return
	}

	dialog.ShowInformation("Saved", fmt.Sprintf("File saved to %s", g.currentFile), g.window)
}

func (g *GUIApp) saveAsFile() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = writer.Write([]byte(g.editor.Text))
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}

		g.currentFile = writer.URI().Path()
		g.fileLabel.SetText(filepath.Base(g.currentFile))
		g.window.SetTitle(fmt.Sprintf("Parselt - %s", filepath.Base(g.currentFile)))

		dialog.ShowInformation("Saved", fmt.Sprintf("File saved to %s", g.currentFile), g.window)
	}, g.window)
}

func (g *GUIApp) toggleView() {
	if g.splitPanel.Offset > 0.75 {
		g.splitPanel.SetOffset(0.0) // Show preview only
	} else if g.splitPanel.Offset < 0.25 {
		g.splitPanel.SetOffset(0.5) // Show both
	} else {
		g.splitPanel.SetOffset(1.0) // Show editor only
	}
}

func (g *GUIApp) showAbout() {
	dialog.ShowInformation("About Parselt",
		"Parselt - Markdown Editor\n\nA simple and elegant markdown editor built with Go and Fyne.\n\nReusing the terminal app's rendering engine for consistency!",
		g.window)
}

func (g *GUIApp) Run() {
	g.setupUI()

	if len(os.Args) > 2 && os.Args[1] == "-gui" {
		filename := os.Args[2]
		if content, err := os.ReadFile(filename); err == nil {
			g.editor.SetText(string(content))
			g.currentFile = filename
			g.fileLabel.SetText(filepath.Base(filename))
			g.window.SetTitle(fmt.Sprintf("Parselt - %s", filepath.Base(filename)))
		}
	}

	g.window.ShowAndRun()
}
