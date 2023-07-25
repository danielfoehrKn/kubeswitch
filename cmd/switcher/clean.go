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
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/clean"
	"github.com/spf13/cobra"
)

var (
	cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Cleans all temporary and cached kubeconfig files",
		Long:  `Cleans the temporary kubeconfig files created in the directory $HOME/.kube/switch_tmp and flushes every cache`,
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, _, err := initialize()
			if err != nil {
				return err
			}
			return clean.Clean(stores)
		},
	}
)

func init() {
	rootCommand.AddCommand(cleanCmd)
}
