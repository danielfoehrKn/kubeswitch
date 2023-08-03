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

	delete_context "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/delete-context"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/hooks"
	list_contexts "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/list-contexts"
	set_context "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/set-context"
	unset_context "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/unset-context"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	previousContextCmd = &cobra.Command{
		Use:     "set-previous-context",
		Aliases: []string{"spc"},
		Short:   "Switch to the previous context from the history",
		Args:    cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			kubeconfigPath, contextName, err := history.SetPreviousContext(stores, config, stateDirectory, noIndex)
			reportNewContext(kubeconfigPath, contextName)
			return err
		},
	}

	lastContextCmd = &cobra.Command{
		Use:     "set-last-context",
		Aliases: []string{"slc"},
		Short:   "Switch to the last used context from the history",
		Args:    cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			kubeconfigPath, contextName, err := history.SetLastContext(stores, config, stateDirectory, noIndex)
			reportNewContext(kubeconfigPath, contextName)
			return err
		},
	}

	listContextsCmd = &cobra.Command{
		Use:     "list-contexts [wildcard-search]",
		Aliases: []string{"ls"},
		Short:   "List all available contexts",
		Long:    `List all available contexts - give a second parameter to do a wildcard search. Eg: switch list-contexts "*-dev*"`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			var comps []string
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			comps = cobra.AppendActiveHelp(comps, "You can provide a wildcard search string, like so: '*-dev-*' to limit the search")
			return comps, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}
			// Get all contexts by default
			pattern := "*"
			if len(args) == 1 && len(args[0]) > 0 {
				pattern = args[0]
			}
			contexts, err := list_contexts.ListContexts(pattern, stores, config, stateDirectory, noIndex)
			if err != nil {
				return err
			}
			for _, context := range contexts {
				fmt.Println(context)
			}
			return nil
		},
	}

	setContextCmd = &cobra.Command{
		Use:     "set-context",
		Short:   "Switch to context name provided as first argument",
		Long:    `Switch to context name provided as first argument. KubeContext name has to exist in any of the found Kubeconfig files.`,
		Aliases: []string{"set", "sc", "set-context"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			lc, _ := listContexts(toComplete)
			return lc, cobra.ShellCompDirectiveNoFileComp
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New().WithField("hook", "")
			return hooks.Hooks(log, configPath, stateDirectory, "", false)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			kubeconfigPath, contextName, err := set_context.SetContext(args[0], stores, config, stateDirectory, noIndex, true)
			reportNewContext(kubeconfigPath, contextName)
			return err
		},
		SilenceUsage: true,
	}

	deleteContextCmd = &cobra.Command{
		Use:   "delete-context",
		Short: "Delete context name provided as first argument",
		Long:  `Delete context name provided as first argument. KubeContext name has to exist in the current Kubeconfig file.`,
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			lc, _ := listContexts(toComplete)
			return lc, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctxName, err := resolveContextName(args[0])
			fmt.Println("delete-context", ctxName, args, err)
			if err != nil {
				return err
			}
			return delete_context.DeleteContext(ctxName)
		},
	}

	unsetContextCmd = &cobra.Command{
		Use:   "unset-context",
		Short: "Unset current-context",
		Long:  `Unset current-context in the current Kubeconfig file.`,
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return unset_context.UnsetCurrentContext()
		},
	}

	currentContextCmd = &cobra.Command{
		Use:   "current-context",
		Short: "Show current-context",
		Long:  `Show current-context in the current Kubeconfig file.`,
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := util.GetCurrentContext()
			if err != nil {
				return err
			}
			fmt.Println(ctx)
			return nil
		},
	}
)

func init() {
	rootCommand.AddCommand(currentContextCmd)
	rootCommand.AddCommand(deleteContextCmd)
	rootCommand.AddCommand(setContextCmd)
	rootCommand.AddCommand(listContextsCmd)
	rootCommand.AddCommand(unsetContextCmd)
	rootCommand.AddCommand(previousContextCmd)
	rootCommand.AddCommand(lastContextCmd)

	setFlagsForContextCommands(setContextCmd)
	setFlagsForContextCommands(listContextsCmd)
	// need to add flags as the namespace history allows switching to any {context: namespace} combination
	setFlagsForContextCommands(previousContextCmd)
	setFlagsForContextCommands(lastContextCmd)
}

func listContexts(prefix string) ([]string, error) {
	stores, config, err := initialize()
	if err != nil {
		return nil, err
	}

	lc, err := list_contexts.ListContexts("*", stores, config, stateDirectory, noIndex)
	if err != nil {
		return nil, err
	}
	return lc, nil
}

func resolveContextName(contextName string) (string, error) {
	if contextName == "." {
		c, err := util.GetCurrentContext()
		if err != nil {
			return "", err
		}
		contextName = c
	}
	return contextName, nil
}

func setFlagsForContextCommands(command *cobra.Command) {
	setCommonFlags(command)
	command.Flags().StringVar(
		&storageBackend,
		"store",
		"filesystem",
		"the backing store to be searched for kubeconfig files. Can be either \"filesystem\" or \"vault\"")
	command.Flags().StringVar(
		&kubeconfigName,
		"kubeconfig-name",
		defaultKubeconfigName,
		"only shows kubeconfig files with this name. Accepts wilcard arguments '*' and '?'. Defaults to 'config'.")
	command.Flags().StringVar(
		&vaultAPIAddressFromFlag,
		"vault-api-address",
		"",
		"the API address of the Vault store. Overrides the default \"vaultAPIAddress\" field in the SwitchConfig. This flag is overridden by the environment variable \"VAULT_ADDR\".")
	command.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path on the local filesystem to the configuration file.")
	// not used for setContext command. Makes call in switch.sh script easier (no need to exclude flag from call)
	command.Flags().BoolVar(
		&showPreview,
		"show-preview",
		true,
		"show preview of the selected kubeconfig. Possibly makes sense to disable when using vault as the kubeconfig store to prevent excessive requests against the API.")
}

func reportNewContext(kubeconfigPath *string, contextName *string) {
	if kubeconfigPath == nil || contextName == nil {
		return
	}

	// print kubeconfig path and context name to std.out
	// captured by calling script setting KUBECONFIG environment variable
	// prefixed with "__ " to distinguish kubeconfig path output from other responses (e.g., errors, list of context, ...)
	fmt.Printf("__ %s,%s", *kubeconfigPath, *contextName)
}
