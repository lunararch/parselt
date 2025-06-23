package main

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type SharedMarkdownProcessor struct{}

func NewSharedMarkdownProcessor() *SharedMarkdownProcessor {
	return &SharedMarkdownProcessor{}
}

func (smp *SharedMarkdownProcessor) ConvertMarkdownToHTML(content string) string {
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

	return buf.String()
}

func (smp *SharedMarkdownProcessor) UnescapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", `"`)
	text = strings.ReplaceAll(text, "&#39;", "'")
	return text
}

func (smp *SharedMarkdownProcessor) ExtractCodeLanguage(line string) string {
	langRe := regexp.MustCompile(`class="language-([^"]*)"`)
	if matches := langRe.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (smp *SharedMarkdownProcessor) RemoveHTMLTags(text string) string {
	anyTagRe := regexp.MustCompile(`<[^>]*>`)
	return anyTagRe.ReplaceAllString(text, "")
}

func (smp *SharedMarkdownProcessor) ExtractHeaderContent(line string, level int) string {
	var re *regexp.Regexp
	switch level {
	case 1:
		re = regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	case 2:
		re = regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`)
	case 3:
		re = regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`)
	case 4:
		re = regexp.MustCompile(`<h4[^>]*>(.*?)</h4>`)
	default:
		return ""
	}

	if matches := re.FindStringSubmatch(line); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (smp *SharedMarkdownProcessor) ProcessInlineCode(content string) (string, []string) {
	codeRe := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)
	var codeSnippets []string

	content = codeRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			codeContent := strings.TrimSpace(matches[1])
			codeSnippets = append(codeSnippets, codeContent)
			return "{{CODE_PLACEHOLDER}}"
		}
		return match
	})

	return content, codeSnippets
}

func (smp *SharedMarkdownProcessor) ProcessInlineBold(content string) (string, []string) {
	strongRe := regexp.MustCompile(`<strong[^>]*>(.*?)</strong>`)
	var boldSnippets []string

	content = strongRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := strongRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			boldContent := strings.TrimSpace(matches[1])
			boldSnippets = append(boldSnippets, boldContent)
			return "{{BOLD_PLACEHOLDER}}"
		}
		return match
	})

	return content, boldSnippets
}

func (smp *SharedMarkdownProcessor) ProcessInlineItalic(content string) (string, []string) {
	emRe := regexp.MustCompile(`<em[^>]*>(.*?)</em>`)
	var italicSnippets []string

	content = emRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := emRe.FindStringSubmatch(match)
		if len(matches) > 1 {
			italicContent := strings.TrimSpace(matches[1])
			italicSnippets = append(italicSnippets, italicContent)
			return "{{ITALIC_PLACEHOLDER}}"
		}
		return match
	})

	return content, italicSnippets
}
