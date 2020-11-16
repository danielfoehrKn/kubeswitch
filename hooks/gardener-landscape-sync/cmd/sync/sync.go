package sync

import (
	"fmt"
	"io/ioutil"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/danielfoehrkn/kubectlSwitch/hooks/gardener-landscape-sync/pkg"
	"github.com/danielfoehrkn/kubectlSwitch/hooks/gardener-landscape-sync/pkg/hookstore"
)

const vaultTokenFileName = ".vault-token"


var (
	logger = logrus.New()
	log    = logger.WithField("hook", "gardener-landscape-sync")

	gardenKubeconfigPath string
	exportPath           string
	landscapeName        string
	clean                bool
	shootKubeconfigName  string
	kubeconfigStore      string
	vaultAPIAddress      string
	stateDir             string

	rootCommand = &cobra.Command{
		Use:   "sync",
		Short: "Sync the kubeconfig of Shoot clusters to vault or the local filesystem.",
		Long:  `Hook for the \"switch\" tool for Gardener landscapes to sync the kubeconfigs of Shoot clusters.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			var store hookstore.KubeconfigStore
			switch kubeconfigStore {
			case hookstore.KubeconfigStoreFilesystem:
				store = &hookstore.FileStore{}
			case hookstore.KubeconfigStoreVault:
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

				store = &hookstore.VaultStore{
					Client: client,
				}
			default:
				return fmt.Errorf("unknown store %q", kubeconfigStore)
			}

			return pkg.RunHook(log, store, clean, shootKubeconfigName, gardenKubeconfigPath, exportPath, landscapeName, stateDir)
		},
	}
)

func NewCommandStartSync() *cobra.Command {
	return rootCommand
}

func init() {
	logger.SetOutput(os.Stdout)
	rootCommand.Flags().StringVar(
		&gardenKubeconfigPath,
		"garden-kubeconfig-path",
		"",
		"local directory path to the kubeconfig of the Garden cluster. The cluster has to contain the Shoot resources.")
	rootCommand.Flags().StringVar(
		&exportPath,
		"export-path",
		"",
		"root of the path where the Shoot kubeconfig files are exported to. Can be a local filesystem path or path for vault. The path for exported kubeconfigs is: export-directory/<landscape-name>/shoots/seed-<seed-name>/<landscape-name>-shoot-<project-name>-<shoot-name>.")
	rootCommand.Flags().StringVar(
		&landscapeName,
		"landscape-name",
		"",
		"name of the Gardener landscape e.g \"dev\".")
	rootCommand.Flags().StringVar(
		&shootKubeconfigName,
		"export-kubeconfig-name",
		"config",
		"name for all the exported shoot cluster kubeconfig files.")
	rootCommand.Flags().BoolVar(
		&clean,
		"clean-directory",
		false,
		"clean the export path and all sub paths before exporting the new kubeconfig files. Used to prevent holding on to kubeconfigs of already deleted clusters.")
	rootCommand.Flags().StringVar(
		&kubeconfigStore,
		"store",
		"filesystem",
		"the storage for the kubeconfig files. Can be either \"filesystem\" or \"vault\"")
	rootCommand.Flags().StringVar(
		&stateDir,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the switchers state directory used to read the Search Index.")
	rootCommand.Flags().StringVar(
		&vaultAPIAddress,
		"vault-api-address",
		"",
		"the API address of the Vault store.")

}
