package switcher

import (
	gardenercontrolplane "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/gardener"
	"github.com/spf13/cobra"
	"os"
)

var (
	gardenerCmd = &cobra.Command{
		Use:   "gardener",
		Short: "gardener specific commands",
		Long:  `Commands that can only be used if a Gardener store is configured.`,
	}

	controlplaneCmd = &cobra.Command{
		Use:   "controlplane",
		Short: "Switch to the Shoot's controlplane",
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, _, err := initialize()
			if err != nil {
				return err
			}

			_, err = gardenercontrolplane.SwitchToControlplane(stores, getKubeconfigPathFromFlag())
			return err
		},
	}
)

func init() {
	setCommonFlags(controlplaneCmd)
	controlplaneCmd.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path on the local filesystem to the configuration file.")

	gardenerCmd.AddCommand(controlplaneCmd)

	rootCommand.AddCommand(gardenerCmd)
}
