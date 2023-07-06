package switcher

import (
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/ns"
	"github.com/spf13/cobra"
)

var (
	namespaceCommand = &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"ns"},
		Short:   "Change the current namespace",
		Long:    `Search namespaces in the current cluster and change to it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && len(args[0]) > 0 {
				return ns.SwitchToNamespace(args[0], getKubeconfigPathFromFlag())
			}

			return ns.SwitchNamespace(getKubeconfigPathFromFlag(), stateDirectory, noIndex)
		},
		SilenceErrors: true,
	}
)

func init() {
	setCommonFlags(namespaceCommand)
	rootCommand.AddCommand(namespaceCommand)
}
