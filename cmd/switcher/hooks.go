package switcher

import (
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/hooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	hookCmd = &cobra.Command{
		Use:   "hooks",
		Short: "Run configured hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New().WithField("hook", hookName)
			return hooks.Hooks(log, configPath, stateDirectory, hookName, runImmediately)
		},
	}

	hookLsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List configured hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New().WithField("hook-ls", hookName)
			return hooks.ListHooks(log, configPath, stateDirectory)
		},
	}
)

func init() {
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

	rootCommand.AddCommand(hookCmd)
}
