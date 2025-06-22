# Parselt - Markdown Editor



A dual-interface markdown editor built with Go, featuring both terminal and GUI applications with live preview functionality.



## Features



### Core Functionality

- **Live Preview** - Real-time markdown rendering as you type

- **Dual Interface** - Choose between terminal TUI or native GUI

- **File Management** - Create, open, save, and save-as operations

- **Cross-Platform** - Works on Windows, macOS, and Linux



### Markdown Support

- **Headers** - H1 through H4 with custom styling

- **Lists** - Both ordered (numbered) and unordered (bullet) lists

- **Nested Lists** - Multiple levels of indentation

- **Inline Formatting** - Bold, italic, and inline code

- **Code Blocks** - Fenced code blocks with language syntax highlighting

- **Blockquotes** - Quote formatting with visual indicators

- **Tables** - GitHub Flavored Markdown (GFM) table support

- **Task Lists** - Checkbox lists

- **Strikethrough** - Text strikethrough formatting



## Installation



### Prerequisites

- Go 1.19 or later

- Git



### Build from Source

```bash

git clone https://github.com/lunararch/parselt.git

cd parselt

go mod tidy

go build -o parselt

```



## Usage



### Terminal Version (TUI)

```bash

# Start with a new file

./parselt



# Open an existing file

./parselt myfile.md



# Create and open a new file

./parselt newfile.md

```



#### Terminal Keyboard Shortcuts

- `Ctrl+S` - Save file

- `Ctrl+P` - Switch to preview mode

- `Ctrl+E` - Switch to edit mode

- `Ctrl+H` - Toggle help

- `Ctrl+Q` - Quit application



### GUI Version

```bash

# Start GUI with a new file

./parselt -gui



# Open an existing file in GUI

./parselt -gui myfile.md

```



#### GUI Features

- **Split View** - Editor and preview side-by-side

- **Menu Bar** - File, View, and Help menus

- **Keyboard Shortcuts** - Ctrl+S to save, Ctrl+O to open

- **View Modes** - Editor only, preview only, or split view



## Supported Markdown Features



### Headers

```markdown

# Header 1

## Header 2

### Header 3

#### Header 4

```



### Lists

```markdown

# Unordered Lists

- Item 1

- Item 2

  - Nested item

  - Another nested item

    - Deep nested item



# Ordered Lists

1. First item

2. Second item

   1. Nested numbered item

   2. Another nested numbered item

```



### Text Formatting

```markdown

**Bold text**

*Italic text*

`Inline code`

```



### Code Blocks

````markdown

```go

func main() {

    fmt.Println("Hello, World!")

}

```



```bash

echo "Terminal commands"

ls -la

```



```javascript

const greeting = "Hello, World!";

console.log(greeting);

```

````



### Blockquotes

```markdown

> This is a blockquote

> It can span multiple lines

```



### Complex Example

```markdown

# Project Overview

This project consists of two main components:

1. A terminal-based editor

2. A GUI-based editor



## Features

- **Live Preview** - Real-time rendering

- **Multiple Formats** - Support for:

  - Headers and subheaders

  - Lists (ordered and unordered)

  - Code blocks with syntax highlighting

  - Inline `code` formatting



### Code Example

```go

func main() {

    fmt.Println("Welcome to Parselt!")

}

```



> Note: Both interfaces share the same markdown rendering engine

```



## Technical Details



### Architecture

- **main.go** - Application entry point and routing

- **terminal.go** - Terminal User Interface using Bubble Tea

- **gui.go** - Graphical User Interface using Fyne

- **Shared Rendering** - Both interfaces use the same Goldmark-based renderer



### Dependencies

- [Goldmark](https://github.com/yuin/goldmark) - Markdown parser and renderer

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework

- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

- [Fyne](https://fyne.io/) - Cross-platform GUI framework



### Rendering Pipeline

1. User types markdown in editor

2. Goldmark converts markdown to HTML

3. HTML is processed and styled for the target interface:

   - Terminal: HTML → Styled terminal output using Lipgloss

   - GUI: HTML → Clean markdown for Fyne's RichText widget



## Contributing



1. Fork the repository

2. Create a feature branch (`git checkout -b feature/amazing-feature`)

3. Commit your changes (`git commit -m 'Add some amazing feature'`)

4. Push to the branch (`git push origin feature/amazing-feature`)

5. Open a Pull Request



## License



This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.



## Acknowledgments



- Built with [Go](https://golang.org/)

- Terminal UI powered by [Charm](https://charm.sh/) libraries

- GUI powered by [Fyne](https://fyne.io/)

- Markdown parsing by [Goldmark](https://github.com/yuin/goldmark)