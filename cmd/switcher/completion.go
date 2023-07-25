// Copyright 2021 The Kubeswitch authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package switcher

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	setName string

	completionCmd = &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Short:                 "generate completion script",
		Long:                  "load the completion script for switch into the current shell",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			if setName != "" {
				root.Use = setName
			}
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			}
			return fmt.Errorf("unsupported shell type: %s", args[0])
		},
	}
)

func init() {
	completionCmd.Flags().StringVarP(&setName, "cmd", "c", "", "generate completion for the specified command")

	rootCommand.AddCommand(completionCmd)
}
