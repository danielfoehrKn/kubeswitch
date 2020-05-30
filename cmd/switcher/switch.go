package switcher

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
)

var (
	kubeconfigDir     string
	kubeconfigName     string
	showPreview     bool
	cmd = &cobra.Command{
		Use:   "switch",
		Short: "Launch the kubeconfig switcher",
		Long: `Simple tool for switching between kubeconfig files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.Switcher(kubeconfigDir, kubeconfigName, showPreview)
		},
	}
)

func NewCommandStartSwitcher() *cobra.Command {
	return cmd
}

func init() {
	cmd.Flags().StringVar(
		&kubeconfigDir,
		"kubeconfig-directory",
		os.ExpandEnv("$HOME/.kube/switch"),
		"directory containing the kubeconfig files.")

	cmd.Flags().StringVar(
		&kubeconfigName,
		"kubeconfig-name",
		"config",
		"only shows kubeconfig files with exactly this name.")
	cmd.Flags().BoolVar(
		&showPreview,
		"show-preview",
		true,
		"show preview of the selected kubeconfig.")
}
