package main

import (
	"fmt"
	"os"

	"github.com/danielfoehrkn/kubectlSwitch/cmd/switcher"
)

func main() {
	rootCommand := switcher.NewCommandStartSwitcher()

	// if first argument is not found, assume it is a context name
	// hence call default subcommand
	cmd, _, err := rootCommand.Find(os.Args[1:])
	if err != nil || cmd == nil {
		args := append([]string{"set-context"}, os.Args[1:]...)
		rootCommand.SetArgs(args)
	}

	// cobra somehow does  not recognize - as a valid command
	if os.Args[1] == "-" {
		args := append([]string{"set-previous-context"}, os.Args[1:]...)
		rootCommand.SetArgs(args)
	}

	if err := rootCommand.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
