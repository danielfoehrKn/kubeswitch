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
