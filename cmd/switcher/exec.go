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

	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/exec"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/spf13/cobra"
)

var (
	execCmd = &cobra.Command{
		Use:                   "exec wildcard-search -- COMMAND [args...]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"e"},
		Short:                 "Execute any command towards the matching contexts from the wildcard search",
		Long:                  `Execute any command to all the matching cluster contexts given by the search parameter. Eg: switch exec "*-dev-?" -- kubectl get namespaces"`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			var comps []string
			if len(args) == 0 {
				comps = cobra.AppendActiveHelp(comps, "You must provide a wildcard search string, like so: '*-dev-*'")
				return comps, cobra.ShellCompDirectiveNoFileComp
			} else if len(args) == 1 {
				comps = cobra.AppendActiveHelp(comps, "Give a '--' to indicate start of command")
				return comps, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) >= 2 {
				comps = cobra.AppendActiveHelp(comps, "Provide the command to send to the contexts")
			}
			return comps, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}
			// split additional args from the command and populate args after "--"
			cmdArgs := util.SplitAdditionalArgs(&args)
			if len(cmdArgs) >= 1 && len(args[0]) > 0 {
				return exec.ExecuteCommand(args[0], cmdArgs, stores, config, stateDirectory, noIndex, showDebugLogs)
			}
			return fmt.Errorf("please provide a search string and the command to execute on each cluster")
		},
	}
)

func init() {
	execCmd.Flags().BoolVar(
		&showDebugLogs,
		"debug",
		false,
		"show debug logs")

	rootCommand.AddCommand(execCmd)
}
