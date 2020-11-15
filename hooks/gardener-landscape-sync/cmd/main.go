package main

import (
	"fmt"
	"os"

	cmd "github.com/danielfoehrkn/kubectlSwitch/hooks/gardener-landscape-sync/cmd/sync"
)

func main() {
	command := cmd.NewCommandStartSync()
	if err := command.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
