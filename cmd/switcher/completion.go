package switcher

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	setName string

	completionCmd = &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Short:                 "generate completion script",
		Long:                  "load the completion script for switch into the current shell",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			if setName != "" {
				root.Use = setName
			}
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			}
			return fmt.Errorf("unsupported shell type: %s", args[0])
		},
	}
)

func init() {
	completionCmd.Flags().StringVarP(&setName, "cmd", "c", "", "generate completion for the specified command")

	rootCommand.AddCommand(completionCmd)
}
