package switcher

import (
	"fmt"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	aliasContextCmd = &cobra.Command{
		Use:   "alias",
		Short: "Create an alias for a context. Use ALIAS=CONTEXT_NAME",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || !strings.Contains(args[0], "=") || len(strings.Split(args[0], "=")) != 2 {
				return fmt.Errorf("please provide the alias in the form ALIAS=CONTEXT_NAME")
			}

			arguments := strings.Split(args[0], "=")
			ctxName, err := resolveContextName(arguments[1])
			if err != nil {
				return err
			}

			stores, config, err := initialize()
			if err != nil {
				return err
			}

			return alias.Alias(arguments[0], ctxName, stores, config, stateDirectory, noIndex)
		},
		SilenceErrors: true,
	}

	aliasLsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List all existing aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return alias.ListAliases(stateDirectory)
		},
	}

	aliasRmCmd = &cobra.Command{
		Use:   "rm",
		Short: "Remove an existing alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || len(args[0]) == 0 {
				return fmt.Errorf("please provide the alias to remove as the first argument")
			}

			return alias.RemoveAlias(args[0], stateDirectory)
		},
		SilenceErrors: true,
	}
)

func init() {
	aliasRmCmd.Flags().StringVar(
		&stateDirectory,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the state directory.")

	aliasContextCmd.AddCommand(aliasLsCmd)
	aliasContextCmd.AddCommand(aliasRmCmd)

	setFlagsForContextCommands(aliasContextCmd)

	rootCommand.AddCommand(aliasContextCmd)
}
