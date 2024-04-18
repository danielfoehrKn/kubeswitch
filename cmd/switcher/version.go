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
	"slices"

	"github.com/danielfoehrkn/kubeswitch/types"

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
		version       : %s
		build date    : %s
		go version    : %s
		go compiler   : %s
		platform      : %s/%s
		backing-stores: %s
`, version, buildDate, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH, getSupportedStores())

			return nil
		},
	}
)

// getSupportedStores returns a string of supported backing stores
func getSupportedStores() string {
	validStoreKinds := types.ValidStoreKinds.List()
	slices.Sort(validStoreKinds)
	validStoreKindsString := "["
	for _, kind := range validStoreKinds {
		validStoreKindsString = fmt.Sprintf("%s %s", validStoreKindsString, kind)
	}
	return fmt.Sprintf("%s ]", validStoreKindsString)
}

func init() {
	rootCommand.AddCommand(versionCmd)
}
