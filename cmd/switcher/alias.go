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
	"strings"

	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias"
	"github.com/spf13/cobra"
)

var (
	aliasContextCmd = &cobra.Command{
		Use:   "alias",
		Short: "Create an alias for a context. Use ALIAS=CONTEXT_NAME",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || !strings.Contains(args[0], "=") || len(strings.Split(args[0], "=")) != 2 {
				return fmt.Errorf("please provide the alias in the form ALIAS=CONTEXT_NAME")
			}

			arguments := strings.Split(args[0], "=")
			ctxName, err := resolveContextName(arguments[1])
			if err != nil {
				return err
			}

			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return alias.Alias(arguments[0], ctxName, stores, config, stateDirectory, noIndex)
		},
		SilenceErrors: true,
	}

	aliasLsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List all existing aliases",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return alias.ListAliases(stateDirectory)
		},
	}

	aliasRmCmd = &cobra.Command{
		Use:   "rm",
		Short: "Remove an existing alias",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			aliases, _ := alias.GetAliases(stateDirectory)
			return aliases, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return alias.RemoveAlias(args[0], stateDirectory)
		},
		SilenceErrors: true,
	}
)

func init() {
	aliasRmCmd.Flags().StringVar(
		&stateDirectory,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the state directory.")

	aliasContextCmd.AddCommand(aliasLsCmd)
	aliasContextCmd.AddCommand(aliasRmCmd)

	setFlagsForContextCommands(aliasContextCmd)

	rootCommand.AddCommand(aliasContextCmd)
}
