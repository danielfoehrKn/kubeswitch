// Copyright 2025 The Kubeswitch authors
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

	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/show"
	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show kubeconfig for a context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, config, err := initialize()
			if err != nil {
				return fmt.Errorf("cannot initialize: %w", err)
			}

			kubeconfig, err := show.Show(args[0], stores, config, stateDirectory, noIndex)
			if err != nil {
				return fmt.Errorf("cannot show kubeconfig: %w", err)
			}

			cmd.Println(string(kubeconfig))
			return nil
		},
		SilenceErrors: true,
	}
)

func init() {
	rootCommand.AddCommand(showCmd)
}
