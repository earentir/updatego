package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// PrintShellHints prints export lines for the user's shell.
func PrintShellHints(goFullPath string) {
	fmt.Println("Add the following to your shell profile:")
	home, err := os.UserHomeDir()
	gopath := filepath.Join(home, "go")
	if err != nil {
		gopath = "$HOME/go"
	}
	fmt.Printf("export GOROOT=%s\n", goFullPath)
	fmt.Printf("export PATH=\"$GOROOT/bin:$PATH\"\n")
	fmt.Printf("export GOPATH=%s\n", gopath)
}
