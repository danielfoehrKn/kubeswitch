// Copyright 2021 The Kubeswitch authors
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

package switcher

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/utils/pointer"

	"github.com/danielfoehrkn/kubeswitch/pkg"
	switchconfig "github.com/danielfoehrkn/kubeswitch/pkg/config"
	"github.com/danielfoehrkn/kubeswitch/pkg/config/validation"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/clean"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/hooks"
	list_contexts "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/list-contexts"
	setcontext "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/set-context"
	"github.com/danielfoehrkn/kubeswitch/types"
)

const (
	vaultTokenFileName          = ".vault-token"
	defaultKubeconfigName       = "config"
	defaultKubeconfigPath       = "$HOME/.kube/config"
	linuxEnvKubeconfigSeperator = ":"
)

var (
	// root command
	kubeconfigPath string
	kubeconfigName string
	showPreview    bool

	// vault store
	storageBackend          string
	vaultAPIAddressFromFlag string

	// hook command
	configPath     string
	stateDirectory string
	hookName       string
	runImmediately bool

	// version command
	version   string
	buildDate string

	showDebugLogs bool
	noIndex       bool

	rootCommand = &cobra.Command{
		Use:     "switch",
		Short:   "Launch the switch binary",
		Long:    `The kubectx for operators.`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			// config file setting overwrites the command line default (--showPreview true)
			if showPreview && config.ShowPreview != nil && !*config.ShowPreview {
				showPreview = false
			}

			return pkg.Switcher(stores, config, stateDirectory, noIndex, showPreview)
		},
	}
)

func init() {
	aliasContextCmd := &cobra.Command{
		Use:   "alias",
		Short: "Create an alias for a context. Use ALIAS=CONTEXT_NAME",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || !strings.Contains(args[0], "=") || len(strings.Split(args[0], "=")) != 2 {
				return fmt.Errorf("please provide the alias in the form ALIAS=CONTEXT_NAME")
			}
			arguments := strings.Split(args[0], "=")

			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return alias.Alias(arguments[0], arguments[1], stores, config, stateDirectory, noIndex)
		},
	}

	aliasLsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all existing aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return alias.ListAliases(stateDirectory)
		},
	}

	aliasRmCmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove an existing alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || len(args[0]) == 0 {
				return fmt.Errorf("please provide the alias to remove as the first argument")
			}

			return alias.RemoveAlias(args[0], stateDirectory)
		},
	}

	aliasRmCmd.Flags().StringVar(
		&stateDirectory,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the state directory.")

	aliasContextCmd.AddCommand(aliasLsCmd)
	aliasContextCmd.AddCommand(aliasRmCmd)

	previousContextCmd := &cobra.Command{
		Use:   "set-previous-context",
		Short: "Switch to the previous context from the history",
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return history.SetPreviousContext(stores, config, stateDirectory, noIndex)
		},
	}

	lastContextCmd := &cobra.Command{
		Use:   "set-last-context",
		Short: "Switch to the last used context from the history",
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return history.SetLastContext(stores, config, stateDirectory, noIndex)
		},
	}

	historyCmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"h"},
		Short:   "Switch to any previous context from the history",
		Long:    `Lists the context history with the ability to switch to a previous context.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return history.SwitchToHistory(stores, config, stateDirectory, noIndex)
		},
	}

	setContextCmd := &cobra.Command{
		Use:   "set-context",
		Short: "Switch to context name provided as first argument",
		Long:  `Switch to context name provided as first argument. KubeContext name has to exist in any of the found Kubeconfig files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return setcontext.SetContext(args[0], stores, config, stateDirectory, noIndex, true)
		},
	}

	listContextsCmd := &cobra.Command{
		Use:     "list-contexts",
		Short:   "List all available contexts without fuzzy search",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return list_contexts.ListContexts(stores, config, stateDirectory, noIndex)
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "clean",
		Short: "Cleans all temporary kubeconfig files",
		Long:  `Cleans the temporary kubeconfig files created in the directory $HOME/.kube/switch_tmp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return clean.Clean()
		},
	}

	hookCmd := &cobra.Command{
		Use:   "hooks",
		Short: "Run configured hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New().WithField("hook", hookName)
			return hooks.Hooks(log, configPath, stateDirectory, hookName, runImmediately)
		},
	}

	hookLsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List configured hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New().WithField("hook-ls", hookName)
			return hooks.ListHooks(log, configPath, stateDirectory)
		},
	}
	hookLsCmd.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path on the local filesystem to the configuration file.")

	hookLsCmd.Flags().StringVar(
		&stateDirectory,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the state directory.")

	hookCmd.AddCommand(hookLsCmd)

	hookCmd.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path on the local filesystem to the configuration file.")

	hookCmd.Flags().StringVar(
		&stateDirectory,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the state directory.")

	hookCmd.Flags().StringVar(
		&hookName,
		"hook-name",
		"",
		"the name of the hook that should be run.")

	hookCmd.Flags().BoolVar(
		&runImmediately,
		"run-immediately",
		true,
		"run hooks right away. Do not respect the hooks execution configuration.")

	versionCmd := &cobra.Command{
		Use:     "version",
		Short:   "show Switch Version info",
		Long:    "show the Switch version information",
		Example: "switch version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf(`Switch:
		version     : %s
		build date  : %s
		go version  : %s
		go compiler : %s
		platform    : %s/%s
`, version, buildDate, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)

			return nil
		},
	}
	rootCommand.AddCommand(setContextCmd)
	rootCommand.AddCommand(listContextsCmd)
	rootCommand.AddCommand(deleteCmd)
	rootCommand.AddCommand(hookCmd)
	rootCommand.AddCommand(historyCmd)
	rootCommand.AddCommand(previousContextCmd)
	rootCommand.AddCommand(lastContextCmd)
	rootCommand.AddCommand(aliasContextCmd)
	rootCommand.AddCommand(versionCmd)

	setContextCmd.SilenceUsage = true
	aliasContextCmd.SilenceErrors = true
	aliasRmCmd.SilenceErrors = true

	setCommonFlags(setContextCmd)
	setCommonFlags(listContextsCmd)
	setCommonFlags(historyCmd)
	setCommonFlags(previousContextCmd)
	setCommonFlags(lastContextCmd)
	setCommonFlags(aliasContextCmd)
}

func NewCommandStartSwitcher() *cobra.Command {
	return rootCommand
}

func init() {
	setCommonFlags(rootCommand)
	rootCommand.SilenceUsage = true
}

func setCommonFlags(command *cobra.Command) {
	command.Flags().BoolVar(
		&showDebugLogs,
		"debug",
		false,
		"show debug logs")
	command.Flags().BoolVar(
		&noIndex,
		"no-index",
		false,
		"stores do not read from index files. The index is refreshed.")
	command.Flags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		defaultKubeconfigPath,
		"path to be recursively searched for kubeconfigs. Can be a file or a directory on the local filesystem or a path in Vault.")
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
		&stateDirectory,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the local directory used for storing internal state.")
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

func initialize() ([]store.KubeconfigStore, *types.Config, error) {
	if showDebugLogs {
		logrus.SetLevel(logrus.DebugLevel)
	}

	config, err := switchconfig.LoadConfigFromFile(expandPath(configPath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read switch config file: %v", err)
	}

	if config != nil {
		if errList := validation.ValidateConfig(config); errList != nil && len(errList) > 0 {
			return nil, nil, fmt.Errorf("the switch configuration file contains errors: %s", errList.ToAggregate().Error())
		}
	} else {
		config = &types.Config{}
	}

	if kubeconfigName == defaultKubeconfigName {
		if config.KubeconfigName != nil && *config.KubeconfigName != "" {
			kubeconfigName = *config.KubeconfigName
		}
	}

	storeFromFlags := getStoreFromFlagAndEnv(config)
	if storeFromFlags != nil {
		config.KubeconfigStores = append(config.KubeconfigStores, *storeFromFlags)
	}

	if len(config.KubeconfigStores) == 0 {
		return nil, nil, fmt.Errorf("you need to point kubeswitch to a kubeconfig file. This can be done by setting the environment variable KUBECONFIG, setting the flag --kubeconfig-path, having a default kubeconfig file at ~/.kube/config or providing a switch configuration file")
	}

	var stores []store.KubeconfigStore
	for _, kubeconfigStoreFromConfig := range config.KubeconfigStores {
		var s store.KubeconfigStore

		switch kubeconfigStoreFromConfig.Kind {
		case types.StoreKindFilesystem:
			filesystemStore, err := store.NewFilesystemStore(kubeconfigName, kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = filesystemStore

		case types.StoreKindVault:
			vaultStore, err := store.NewVaultStore(vaultAPIAddressFromFlag,
				vaultTokenFileName,
				kubeconfigName,
				kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = vaultStore

		case types.StoreKindGardener:
			gardenerStore, err := store.NewGardenerStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to create Gardener store: %w", err)
			}
			s = gardenerStore

		case types.StoreKindGKE:
			gkeStore, err := store.NewGKEStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to create GKE store: %w", err)
			}
			s = gkeStore
		default:
			return nil, nil, fmt.Errorf("unknown store %q", kubeconfigStoreFromConfig.Kind)
		}
		logrus.Debugf("Added store with kind %s and ID %s", s.GetKind(), s.GetID())
		stores = append(stores, s)
	}
	return stores, config, nil
}

// getStoreFromFlagAndEnv translates the kubeconfig flag --kubeconfig-path & environment variable KUBECONFIG into a
// dedicated store in addition to the stores configured in the switch-config.yaml.
// This way, it is "just another store" -> does not need special handling
func getStoreFromFlagAndEnv(config *types.Config) *types.KubeconfigStore {
	var paths []string

	pathFromFlag := getKubeconfigPathFromFlag()
	if len(pathFromFlag) > 0 {
		logrus.Debugf("Using kubeconfig path from flag %s", pathFromFlag)
		paths = append(paths, pathFromFlag)
	}

	kubeconfigPathFromEnv := os.Getenv("KUBECONFIG")

	pathsFromEnv := strings.Split(kubeconfigPathFromEnv, linuxEnvKubeconfigSeperator)

	for _, path := range pathsFromEnv {
		if !isDuplicatePath(config.KubeconfigStores, path) && !strings.HasSuffix(path, ".tmp") && path != "" {
			// the KUBECONFIG env sets a unique, non kubeswitch set, env variable to a kubeconfig.
			paths = append(paths, expandPath(path))
			logrus.Debugf("Adding kubeconfig path from KUBECONFIG env %s", kubeconfigPathFromEnv)
		}
	}

	if len(paths) == 0 {
		return nil
	}

	return &types.KubeconfigStore{
		ID:             pointer.StringPtr("env-and-flag"),
		Kind:           types.StoreKind(storageBackend),
		KubeconfigName: pointer.StringPtr(kubeconfigName),
		Paths:          paths,
		ShowPrefix:     pointer.BoolPtr(false),
	}
}

// getKubeconfigPathFromFlag gets the kubeconfig path configured in the flag --kubeconfig-path
// does not add the path in case the configured path does not exist
// this is to not require a kubeconfig file in the default location
func getKubeconfigPathFromFlag() string {
	if len(kubeconfigPath) == 0 {
		return ""
	}

	kubeconfigPath = strings.ReplaceAll(kubeconfigPath, "~", "$HOME")
	if kubeconfigPath == defaultKubeconfigPath {
		// do not return, if the kubeconfig under the default kubeconfig path does not exist
		if _, err := os.Stat(os.ExpandEnv(defaultKubeconfigPath)); err != nil {
			return ""
		}
		// the kubeconfig under the default path exists -> return it.
		return os.ExpandEnv(defaultKubeconfigPath)
	}

	// the flag sets a non-default kubeconfig path
	return os.ExpandEnv(kubeconfigPath)
}

// isDuplicatePath searches through all kubeconfig stores in the switch-config.yaml and checks if the
// given path is already configured in any of these stores
// returns true if it is already configureed
func isDuplicatePath(kubeconfigStores []types.KubeconfigStore, newPath string) bool {
	// O(n square) operation
	// but fortunately it is highly unlikely that there are many stores and paths configured
	for _, store := range kubeconfigStores {
		for _, path := range store.Paths {
			if path == newPath {
				return true
			}
		}
	}
	return false
}

func expandPath(path string) string {
	path = strings.ReplaceAll(path, "~", "$HOME")
	return os.ExpandEnv(path)
}
