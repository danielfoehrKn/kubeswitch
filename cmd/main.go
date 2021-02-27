// Copyright 2021 Daniel Foehr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"

	"github.com/danielfoehrkn/kubeswitch/cmd/switcher"
)

func main() {
	rootCommand := switcher.NewCommandStartSwitcher()

	// if first argument is not found, assume it is a context name
	// hence call default subcommand
	if len(os.Args) > 1 {
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
	}

	if err := rootCommand.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
