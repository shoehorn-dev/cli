package main

import (
	"os"

	"github.com/imbabamba/shoehorn-cli/cmd/shoehorn/commands"
	_ "github.com/imbabamba/shoehorn-cli/cmd/shoehorn/commands/get" // register get subcommands
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
