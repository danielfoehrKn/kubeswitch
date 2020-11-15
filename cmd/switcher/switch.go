package switcher

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/clean"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/hooks"
	"github.com/danielfoehrkn/kubectlSwitch/types"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const vaultTokenFileName = ".vault-token"

var (
	// root command
	kubeconfigDir  string
	kubeconfigName string
	showPreview    bool

	// vault store
	storageBackend              string
	vaultAPIAddress             string
	vaultSecretEnginePathPrefix string

	// hook command
	configPath     string
	stateDir       string
	hookName       string
	runImmediately bool

	rootCommand = &cobra.Command{
		Use:   "switch",
		Short: "Launch the kubeconfig switcher",
		Long:  `Simple tool for switching between kubeconfig files.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			var (
				kubeconfigStore store.KubeconfigStore
				log             *logrus.Entry
			)

			switch storageBackend {
			case string(types.StoreKindFilesystem):
				log = logrus.New().WithField("store", types.StoreKindFilesystem)

				kubeconfigStore = &store.FilesystemStore{
					KubeconfigDirectory: kubeconfigDir,
					KubeconfigName:      kubeconfigName,
				}
			case string(types.StoreKindVault):
				log = logrus.New().WithField("store", types.StoreKindVault)

				vaultAddress := os.Getenv("VAULT_ADDR")
				if len(vaultAddress) > 0 {
					vaultAPIAddress = vaultAddress
				}

				if len(vaultAPIAddress) == 0 {
					return fmt.Errorf("when using the vault kubeconfig store, the API address of the vault has to be provided either by command line argument \"vaultAPI\" or via environment variable \"VAULT_ADDR\"")
				}

				home, err := os.UserHomeDir()
				if err != nil {
					return err
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
					return fmt.Errorf("when using the vault kubeconfig store, a vault API token must be provided.  Per default, the token file in  \"~.vault-token\" is used. The default oken can be overriden via the  environment variable \"VAULT_ADDR\"")
				}

				config := &vaultapi.Config{
					Address: vaultAPIAddress,
				}
				client, err := vaultapi.NewClient(config)
				if err != nil {
					return err
				}
				client.SetToken(vaultToken)

				kubeconfigStore = &store.VaultStore{
					KubeconfigName:              kubeconfigName,
					Client:                      client,
					VaultSecretEnginePathPrefix: vaultSecretEnginePathPrefix,
				}
			default:
				return fmt.Errorf("unknown store %q", kubeconfigStore)
			}

			return pkg.Switcher(log, kubeconfigStore, configPath, stateDir, showPreview)
		},
	}
)

func init() {
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
		Short: "Runs configured hooks",
		Long:  `Runs hooks configured in the configuration path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New().WithField("hook", hookName)
			return hooks.Hooks(log, configPath, stateDir, hookName, runImmediately)
		},
	}

	hookCmd.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path to the configuration file.")

	hookCmd.Flags().StringVar(
		&stateDir,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the state directory.")

	hookCmd.Flags().StringVar(
		&hookName,
		"name",
		"",
		"the name of the hook that should be run.")

	hookCmd.Flags().BoolVar(
		&runImmediately,
		"run-immediately",
		true,
		"run hooks right away. Do not respect the hooks execution configuration.")

	rootCommand.AddCommand(deleteCmd)
	rootCommand.AddCommand(hookCmd)
}

func NewCommandStartSwitcher() *cobra.Command {
	return rootCommand
}

func init() {
	rootCommand.Flags().StringVar(
		&kubeconfigDir,
		"kubeconfig-directory",
		os.ExpandEnv("$HOME/.kube"),
		"directory on the local filesystem to be recursively searched for kubeconfig files. When using vault as store, the flag \"vaultSecretEnginePathPrefix\" instead defines the path where kubeconfig secrets are expected.")
	rootCommand.Flags().StringVar(
		&stateDir,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the local directory used for storing internal state.")
	rootCommand.Flags().StringVar(
		&kubeconfigName,
		"kubeconfig-name",
		"config",
		"only shows kubeconfig files with this name. Accepts wilcard arguments '*' and '?'. Defaults to 'config'.")
	rootCommand.Flags().BoolVar(
		&showPreview,
		"show-preview",
		true,
		"show preview of the selected kubeconfig. Possibly makes sense to disable when using vault as the kubeconfig store to prevent excessive requests against the API.")
	rootCommand.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path to the configuration file.")
	rootCommand.Flags().StringVar(
		&vaultAPIAddress,
		"vault-api-address",
		"",
		"the API address of the Vault store.")
	rootCommand.Flags().StringVar(
		&storageBackend,
		"store",
		"filesystem",
		"the storage for the kubeconfig files. Can be either \"filesystem\" or \"vault\"")
	rootCommand.Flags().StringVar(
		&vaultSecretEnginePathPrefix,
		"vaultSecretEnginePathPrefix",
		"",
		"the prefix to use for the vault secret engine when exporting the kubeconfigs. Only used for store \"vault\".")
}
