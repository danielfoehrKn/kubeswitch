package switcher

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "show Switch Version info",
		Long:    "show the Switch version information",
		Example: "switch version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf(`Switch:
		version     : %s
		build date  : %s
		go version  : %s
		go compiler : %s
		platform    : %s/%s
`, version, buildDate, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)

			return nil
		},
	}
)

func init() {
	rootCommand.AddCommand(versionCmd)
}
