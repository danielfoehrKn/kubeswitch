// Copyright 2021 The Kubeswitch authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package switcher

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "show switch version info",
		Long:    "show the switch version information",
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
