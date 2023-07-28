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
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history"
	"github.com/spf13/cobra"
)

var (
	historyCmd = &cobra.Command{
		Use:     "history",
		Aliases: []string{"h", "history"},
		Short:   "Switch to any previous tuple {context,namespace} from the history",
		Long:    `Lists the context history with the ability to switch to a previous context.`,
		Args:    cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			kc, err := history.SwitchToHistory(stores, config, stateDirectory, noIndex)
			reportNewContext(kc)
			return err
		},
	}
)

func init() {
	setFlagsForContextCommands(historyCmd)
	rootCommand.AddCommand(historyCmd)
}
