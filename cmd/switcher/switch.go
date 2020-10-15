package switcher

import (
	"os"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
	"github.com/spf13/cobra"
)

var (
	kubeconfigDir  string
	kubeconfigName string
	showPreview    bool
	rootCommand    = &cobra.Command{
		Use:   "switch",
		Short: "Launch the kubeconfig switcher",
		Long: `Simple tool for switching between kubeconfig files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.Switcher(kubeconfigDir, kubeconfigName, showPreview)
		},
	}
)

func init() {
	deleteCmd := &cobra.Command{
		Use:   "clean",
		Short: "Cleans all temporary kubeconfig files",
		Long: `Cleans the temporary kubeconfig files created in the directory $HOME/.kube/switch_tmp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.Clean()
		},
	}
	rootCommand.AddCommand(deleteCmd)
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
}
