// +build !gui

package cmd

import (
	"fmt"
	"os"
)

func RunGUI() {
	fmt.Println("GUI mode is not available.")
	fmt.Println("To enable GUI mode:")
	fmt.Println("1. Install GUI dependencies (OpenGL, X11):")
	fmt.Println("   Ubuntu/Debian: sudo apt-get install libgl1-mesa-dev xorg-dev")
	fmt.Println("   Fedora/RHEL:   sudo dnf install mesa-libGL-devel gtk3-devel")
	fmt.Println("")
	fmt.Println("2. Build with GUI tag:")
	fmt.Println("   go build -tags gui")
	fmt.Println("")
	fmt.Println("3. Run the GUI:")
	fmt.Println("   ./roub-crt gui")
	fmt.Println("")
	fmt.Println("Currently running in CLI mode. Use --help for available commands.")
	os.Exit(0)
}
