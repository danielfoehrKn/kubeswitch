package switcher

import (
	"fmt"

	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/exec"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/spf13/cobra"
)

var (
	execCmd = &cobra.Command{
		Use:     "exec wildcard-search -- command",
		Aliases: []string{"e"},
		Short:   "Execute any command towards the matching contexts from the wildcard search",
		Long:    `Execute any command to all the matching cluster contexts given by the search parameter. Eg: switch exec "*-dev-?" -- kubectl get namespaces"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return err
			}
			// split additional args from the command and populate args after "--"
			cmdArgs := util.SplitAdditionalArgs(&args)
			if len(cmdArgs) >= 1 && len(args[0]) > 0 {
				return exec.ExecuteCommand(args[0], cmdArgs, stores, config, stateDirectory, noIndex)
			}
			return fmt.Errorf("please provide a search string and the command to execute on each cluster")
		},
	}
)

func init() {
	rootCommand.AddCommand(execCmd)
}
