package switcher

import (
	"os"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
	"github.com/spf13/cobra"
)

var (
	// root command
	kubeconfigDir  string
	kubeconfigName string
	showPreview    bool

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
			return pkg.Switcher(configPath, kubeconfigDir, kubeconfigName, showPreview)
		},
	}
)

func init() {
	deleteCmd := &cobra.Command{
		Use:   "clean",
		Short: "Cleans all temporary kubeconfig files",
		Long:  `Cleans the temporary kubeconfig files created in the directory $HOME/.kube/switch_tmp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.Clean()
		},
	}

	hookCmd := &cobra.Command{
		Use:   "hooks",
		Short: "Runs configured hooks",
		Long:  `Runs hooks configured in the configuration path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.Hooks(configPath, stateDir, hookName, runImmediately)
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
		"directory containing the kubeconfig files.")
	rootCommand.Flags().StringVar(
		&kubeconfigName,
		"kubeconfig-name",
		"config",
		"only shows kubeconfig files with this name. Accepts wilcard arguments '*' and '?'. Defaults to 'config'.")
	rootCommand.Flags().BoolVar(
		&showPreview,
		"show-preview",
		true,
		"show preview of the selected kubeconfig.")
	rootCommand.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path to the configuration file.")
}
