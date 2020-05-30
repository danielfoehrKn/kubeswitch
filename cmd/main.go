package main

import (
	"fmt"
	"os"

	"github.com/danielfoehrkn/kubectlSwitch/cmd/switcher"
)

func main() {
	command := switcher.NewCommandStartSwitcher()
	if err := command.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
