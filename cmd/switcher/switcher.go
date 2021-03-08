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

package switcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/pkg"
	switchconfig "github.com/danielfoehrkn/kubeswitch/pkg/config"
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
	vaultTokenFileName    = ".vault-token"
	defaultKubeconfigName = "config"
	defaultKubeconfigPath = "$HOME/.kube/config"
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

	rootCommand = &cobra.Command{
		Use:   "switch",
		Short: "Launch the switch binary",
		Long:  `The kubectx for operators.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return pkg.Switcher(stores, config, stateDirectory, showPreview)
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

			return alias.Alias(arguments[0], arguments[1], stores, config, stateDirectory)
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

			return history.SetPreviousContext(stores, config, stateDirectory)
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

			return history.ListHistory(stores, config, stateDirectory)
		},
	}

	setContextCmd := &cobra.Command{
		Use:   "set-context",
		Short: "Switch to context name provided as first argument",
		Long:  `Switch to context name provided as first argument. Context name has to exist in any of the found Kubeconfig files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return setcontext.SetContext(args[0], stores, config, stateDirectory)
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

			return list_contexts.ListContexts(stores, config, stateDirectory)
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

	rootCommand.AddCommand(setContextCmd)
	rootCommand.AddCommand(listContextsCmd)
	rootCommand.AddCommand(deleteCmd)
	rootCommand.AddCommand(hookCmd)
	rootCommand.AddCommand(historyCmd)
	rootCommand.AddCommand(previousContextCmd)
	rootCommand.AddCommand(aliasContextCmd)

	setContextCmd.SilenceUsage = true
	aliasContextCmd.SilenceErrors = true
	aliasRmCmd.SilenceErrors = true

	setCommonFlags(setContextCmd)
	setCommonFlags(listContextsCmd)
	setCommonFlags(historyCmd)
	setCommonFlags(previousContextCmd)
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
	command.Flags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		defaultKubeconfigPath,
		"path to be recursively searched for kubeconfig files.  Can be a file or a directory on the local filesystem or a path in Vault.")
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
	config, err := switchconfig.LoadConfigFromFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read switch config file: %v", err)
	}

	if config == nil {
		config = &types.Config{}
	}

	if kubeconfigName == defaultKubeconfigName {
		if config.KubeconfigName != "" {
			kubeconfigName = config.KubeconfigName
		}
	}


	var (
		useVaultStore      = false
		useFilesystemStore = false
		stores             []store.KubeconfigStore
	)

	for _, kubeconfigStoreFromConfig := range config.KubeconfigStores {
		var s store.KubeconfigStore

		switch kubeconfigStoreFromConfig.Kind {
		case types.StoreKindFilesystem:
			// TODO: in the future we want to allow multiple filsystem store configurations
			//  to e.g search for a diff. kubeconfig name or have a different rediscovery interval
			if useFilesystemStore {
				return nil, nil, fmt.Errorf("configuring multliple filesystem stores is not supported. Provide multiple kubeconfig paths within the filesystem store configuration")
			}
			useFilesystemStore = true

			// add kubeconfig path that is set via an environment variable
			kubeconfigEnv := os.Getenv("KUBECONFIG")
			if len(kubeconfigEnv) > 0 && !isDuplicatePath(config.KubeconfigStores, kubeconfigEnv) && !strings.HasSuffix(kubeconfigEnv, ".tmp") {
				kubeconfigStoreFromConfig.Paths = append(kubeconfigStoreFromConfig.Paths, kubeconfigEnv)
			}

			pathFromFlag := getKubeconfigPathFromFlag(config)
			if len(kubeconfigPath) > 0 {
				kubeconfigStoreFromConfig.Paths = append(kubeconfigStoreFromConfig.Paths, pathFromFlag)
			}

			s = &store.FilesystemStore{
				Logger:          logrus.New().WithField("store", types.StoreKindFilesystem),
				KubeconfigStore: kubeconfigStoreFromConfig,
				KubeconfigName:  kubeconfigName,
			}

		case types.StoreKindVault:
			if useVaultStore {
				return nil, nil, fmt.Errorf("configuring multliple vault stores is not supported")
			}
			useVaultStore = true
			vaultStore, err := newVaultStore(kubeconfigStoreFromConfig)
			if err != nil {
				return nil, nil, err
			}
			s = vaultStore
		case types.StoreKindGardener:
			gardenerStore, err := store.NewGardenerStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to create Gardener store: %w", err)
			}
			s = gardenerStore
		default:
			return nil, nil, fmt.Errorf("unknown store %q", kubeconfigStoreFromConfig.Kind)
		}

		stores = append(stores, s)
	}
	return stores, config, nil
}

// getKubeconfigPathFromFlag gets the kubeconfig path configured in the flag --kubeconfig-path
// does not add the path in case the configured path does not exist
// this is to not require a kubeconfig file in the default location
func getKubeconfigPathFromFlag(config *types.Config) string {
	if len(kubeconfigPath) == 0 {
		return ""
	}

	if kubeconfigPath == defaultKubeconfigPath {
		if _, err := os.Stat(os.ExpandEnv(defaultKubeconfigPath)); err != nil {
			return ""
		}
	}

	return os.ExpandEnv(defaultKubeconfigPath)
}

func isDuplicatePath(kubeconfigStores []types.KubeconfigStore, newPath string) bool {
	// expensive operation in case there a lot of kubeconfig stores and paths
	// but fortunately it is highly unlikely that there are ever that many stores and paths configured
	for _, store := range kubeconfigStores {
		for _, path := range store.Paths {
			if path == newPath {
				return true
			}
		}
	}
	return false
}

func newVaultStore(kubeconfigStore types.KubeconfigStore) (*store.VaultStore, error) {
	vaultStoreConfig := &types.StoreConfigVault{}
	if kubeconfigStore.Config != nil {
		buf, err := yaml.Marshal(kubeconfigStore.Config)
		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(buf, vaultStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vault config: %w", err)
		}
	}

	vaultAPI := vaultStoreConfig.VaultAPIAddress
	if len(vaultAPIAddressFromFlag) > 0 {
		vaultAPI = vaultAPIAddressFromFlag
	}

	vaultAddress := os.Getenv("VAULT_ADDR")
	if len(vaultAddress) > 0 {
		vaultAPI = vaultAddress
	}

	if len(vaultAPI) == 0 {
		return nil, fmt.Errorf("when using the vault kubeconfig store, the API address of the vault has to be provided either by command line argument \"vaultAPI\", via environment variable \"VAULT_ADDR\" or via SwitchConfig file")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var vaultToken string

	// https://www.vaultproject.io/docs/commands/token-helper
	tokenBytes, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", home, vaultTokenFileName))
	if tokenBytes != nil {
		vaultToken = string(tokenBytes)
	}

	vaultTokenEnv := os.Getenv("VAULT_TOKEN")
	if len(vaultTokenEnv) > 0 {
		vaultToken = vaultTokenEnv
	}

	if len(vaultToken) == 0 {
		return nil, fmt.Errorf("when using the vault kubeconfig store, a vault API token must be provided. Per default, the token file in \"~.vault-token\" is used. The default token can be overriden via the environment variable \"VAULT_TOKEN\"")
	}

	vaultConfig := &vaultapi.Config{
		Address: vaultAPI,
	}
	client, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetToken(vaultToken)

	return &store.VaultStore{
		Logger:          logrus.New().WithField("store", types.StoreKindVault),
		KubeconfigName:  kubeconfigName,
		KubeconfigStore: kubeconfigStore,
		Client:          client,
	}, nil
}
