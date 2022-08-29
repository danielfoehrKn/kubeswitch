package switcher

import (
	"github.com/spf13/cobra"
	"os"
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
		Run: func(cmd *cobra.Command, args []string) {
			root := cmd.Root()
			if setName != "" {
				root.Use = setName
			}
			switch args[0] {
			case "bash":
				root.GenBashCompletion(os.Stdout)
			case "zsh":
				root.GenZshCompletion(os.Stdout)
			case "fish":
				root.GenFishCompletion(os.Stdout, true)
			}
		},
	}
)

func init() {
	completionCmd.Flags().StringVarP(&setName, "cmd", "c", "", "generate completion for the specified command")

	rootCommand.AddCommand(completionCmd)
}
