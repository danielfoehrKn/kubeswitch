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
	"os"

	gardenercontrolplane "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/gardener"
	"github.com/spf13/cobra"
)

var (
	gardenerCmd = &cobra.Command{
		Use:   "gardener",
		Short: "gardener specific commands",
		Long:  `Commands that can only be used if a Gardener store is configured.`,
	}

	controlplaneCmd = &cobra.Command{
		Use:   "controlplane",
		Short: "Switch to the Shoot's controlplane",
		RunE: func(cmd *cobra.Command, args []string) error {
			stores, _, err := initialize()
			if err != nil {
				return err
			}

			_, err = gardenercontrolplane.SwitchToControlplane(stores, getKubeconfigPathFromFlag())
			return err
		},
	}
)

func init() {
	setCommonFlags(controlplaneCmd)
	controlplaneCmd.Flags().StringVar(
		&configPath,
		"config-path",
		os.ExpandEnv("$HOME/.kube/switch-config.yaml"),
		"path on the local filesystem to the configuration file.")

	gardenerCmd.AddCommand(controlplaneCmd)

	rootCommand.AddCommand(gardenerCmd)
}
