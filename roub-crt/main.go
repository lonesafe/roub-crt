package main

import (
	"os"

	"roub-crt/cmd"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "gui" {
		// GUI mode requires GUI build tag and display libraries
		// Build with: go build -tags gui
		// or run directly: ./roub-crt gui
		cmd.RunGUI()
	} else {
		cmd.Execute()
	}
}
