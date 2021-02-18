package switcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/config"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/clean"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/history"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/hooks"
	list_contexts "github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/list-contexts"
	setcontext "github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/set-context"
	"github.com/danielfoehrkn/kubectlSwitch/types"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const vaultTokenFileName = ".vault-token"

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
		Short: "Launch the kubeconfig switcher",
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
	historyCmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"h"},
		Short:   "Switch to a previous context from the history",
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

			list_contexts.ListContexts(stores, config, stateDirectory)
			return nil
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

	setContextCmd.SilenceUsage = true

	setCommonFlags(setContextCmd)
	setCommonFlags(listContextsCmd)
	setCommonFlags(historyCmd)
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
		os.ExpandEnv("$HOME/.kube/config"),
		"path to be recursively searched for kubeconfig files.  Can be a file or a directory on the local filesystem or a path in Vault.")
	command.Flags().StringVar(
		&storageBackend,
		"store",
		"filesystem",
		"the backing store to be searched for kubeconfig files. Can be either \"filesystem\" or \"vault\"")
	command.Flags().StringVar(
		&kubeconfigName,
		"kubeconfig-name",
		"config",
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
	switchConfig, err := config.LoadConfigFromFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read switch config file: %v", err)
	}

	if switchConfig == nil {
		switchConfig = &types.Config{}
	}

	if len(kubeconfigPath) > 0 {
		switchConfig.KubeconfigPaths = append(switchConfig.KubeconfigPaths, types.KubeconfigPath{
			Path:  kubeconfigPath,
			Store: types.StoreKind(storageBackend),
		})
	}

	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if len(kubeconfigEnv) > 0 && !isDuplicatePath(switchConfig.KubeconfigPaths, kubeconfigEnv) && !strings.HasSuffix(kubeconfigEnv, ".tmp") {
		switchConfig.KubeconfigPaths = append(switchConfig.KubeconfigPaths, types.KubeconfigPath{
			Path:  kubeconfigEnv,
			Store: types.StoreKind(storageBackend),
		})
	}

	var (
		useVaultStore      = false
		useFilesystemStore = false
		stores             []store.KubeconfigStore
	)

	for _, configuredKubeconfigPath := range switchConfig.KubeconfigPaths {
		var s store.KubeconfigStore

		switch configuredKubeconfigPath.Store {
		case types.StoreKindFilesystem:
			if useFilesystemStore {
				continue
			}
			useFilesystemStore = true
			s = &store.FilesystemStore{
				Logger:          logrus.New().WithField("store", types.StoreKindFilesystem),
				KubeconfigPaths: switchConfig.KubeconfigPaths,
				KubeconfigName:  kubeconfigName,
			}
		case types.StoreKindVault:
			if useVaultStore {
				continue
			}
			useVaultStore = true
			vaultStore, err := getVaultStore(switchConfig.VaultAPIAddress, switchConfig.KubeconfigPaths)
			if err != nil {
				return nil, nil, err
			}
			s = vaultStore
		default:
			return nil, nil, fmt.Errorf("unknown store %q", configuredKubeconfigPath.Store)
		}

		stores = append(stores, s)
	}
	return stores, switchConfig, nil
}

func isDuplicatePath(paths []types.KubeconfigPath, newPath string) bool {
	for _, p := range paths {
		if p.Path == newPath {
			return true
		}
	}
	return false
}

func getVaultStore(vaultAPIAddressFromSwitchConfig string, paths []types.KubeconfigPath) (*store.VaultStore, error) {
	vaultAPI := vaultAPIAddressFromSwitchConfig

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

	config := &vaultapi.Config{
		Address: vaultAPI,
	}
	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(vaultToken)

	return &store.VaultStore{
		Logger:          logrus.New().WithField("store", types.StoreKindVault),
		KubeconfigName:  kubeconfigName,
		KubeconfigPaths: paths,
		Client:          client,
	}, nil
}
