package switcher

import (
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history"
	"github.com/spf13/cobra"
)

var (
	historyCmd = &cobra.Command{
		Use:     "history",
		Aliases: []string{"h", "history"},
		Short:   "Switch to any previous tuple {context,namespace} from the history",
		Long:    `Lists the context history with the ability to switch to a previous context.`,
		Args:    cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}

			kc, err := history.SwitchToHistory(stores, config, stateDirectory, noIndex)
			reportNewContext(kc)
			return err
		},
	}
)

func init() {
	setFlagsForContextCommands(historyCmd)
	rootCommand.AddCommand(historyCmd)
}
