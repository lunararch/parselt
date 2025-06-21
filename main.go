package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var filename string
	var useGUI bool

	flag.BoolVar(&useGUI, "gui", false, "Launch GUI version")
	flag.Parse()

	args := flag.Args()
	if len(os.Args) > 1 && os.Args[1] == "-gui" {
		useGUI = true
	}

	if useGUI {
		gui := NewGUIApp()
		gui.Run()
		return
	}

	if len(args) > 0 {
		filename = args[0]
	} else if len(os.Args) > 1 && os.Args[1] != "-gui" {
		filename = os.Args[1]
	}

	if filename != "" {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			file, err := os.Create(filename)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				os.Exit(1)
			}
			file.Close()
		}
	}

	terminal := NewTerminalApp(filename)
	if err := terminal.Run(); err != nil {
		fmt.Printf("Error starting terminal app: %v\n", err)
		os.Exit(1)
	}
}
