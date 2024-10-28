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
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/ns"
	"github.com/spf13/cobra"
)

var (
	checkExistence   bool = true
	namespaceCommand      = &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"ns"},
		Short:   "Change the current namespace",
		Long:    `Search namespaces in the current cluster and change to it.`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := ns.ListNamespaces(getKubeconfigPathFromFlag(), stateDirectory, noIndex)
			return list, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && len(args[0]) > 0 {
				return ns.SwitchToNamespace(args[0], getKubeconfigPathFromFlag(), checkExistence)
			}

			return ns.SwitchNamespace(getKubeconfigPathFromFlag(), stateDirectory, noIndex)
		},
		SilenceErrors: true,
	}
	unsetNamespaceCommand = &cobra.Command{
		Use:   "unset-namespace",
		Short: "Unset the current namespace",
		Long:  `Unset the current namespace in the contexts (effectively changes it to default)`,
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ns.SwitchToNamespace("default", getKubeconfigPathFromFlag(), false)
		},
		SilenceErrors: true,
	}
)

func init() {
	setCommonFlags(namespaceCommand)
	namespaceCommand.Flags().BoolVar(&checkExistence, "check-existence", true, "Check if the namespace exists before switching to it (default true)")
	rootCommand.AddCommand(namespaceCommand)
	rootCommand.AddCommand(unsetNamespaceCommand)
}
