// Copyright 2021 The Kubeswitch authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/becheran/wildmatch-go"
	"github.com/danielfoehrkn/kubeswitch/pkg"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	setcontext "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/set-context"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func ExecuteCommand(pattern string, command []string, stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) error {
	c, err := pkg.DoSearch(stores, config, stateDir, noIndex)
	if err != nil {
		return fmt.Errorf("cannot list contexts: %v", err)
	}

	m := wildmatch.NewWildMatch(pattern)
	var contexts []string
	for discoveredKubeconfig := range *c {
		if discoveredKubeconfig.Error != nil {
			logger.Warnf("cannot list contexts. Error returned from search: %v", discoveredKubeconfig.Error)
			continue
		}

		name := discoveredKubeconfig.Name
		if len(discoveredKubeconfig.Alias) > 0 {
			name = discoveredKubeconfig.Alias
		}
		result := m.IsMatch(name)
		if result {
			contexts = append(contexts, name)
		}
	}
	// Sort alphabetically
	sort.Strings(contexts)

	for _, context := range contexts {
		tmpKubeconfigFile, err := setcontext.SetContext(context, stores, config, stateDir, noIndex, false, false)
		if err != nil {
			return err
		}
		fmt.Printf("=== START Executing on %s ===\n", context)
		cmd := exec.Command(command[0], command[1:]...)

		// Set environment variables for the command
		cmd.Env = os.Environ()

		kubeconfigEnvVar := fmt.Sprintf("KUBECONFIG=%s", *tmpKubeconfigFile)
		cmd.Env = append(cmd.Env, kubeconfigEnvVar)

		// Redirect the command's output to the current process's output
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Environ()

		// Run the command
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Command execution failed: %v\n", err)
			return err
		}
		fmt.Printf("=== END Executing on %s ===\n", context)
	}
	return nil
}
