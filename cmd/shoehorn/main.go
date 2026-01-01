package main

import (
	"os"

	"github.com/imbabamba/shoehorn-cli/cmd/shoehorn/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
