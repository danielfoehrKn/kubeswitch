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

	"github.com/go-cmd/cmd"

	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	list_contexts "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/list-contexts"
	setcontext "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/set-context"
	"github.com/danielfoehrkn/kubeswitch/types"
)

func ExecuteCommand(pattern string, command []string, stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) error {
	contexts, err := list_contexts.ListContexts(pattern, stores, config, stateDir, noIndex)
	if err != nil {
		return err
	}

	for _, context := range contexts {
		tmpKubeconfigFile, _, err := setcontext.SetContext(context, stores, config, stateDir, noIndex, false)
		if err != nil {
			return err
		}
		fmt.Printf("=== START Executing on %s ===\n", context)

		// Disable output buffering, enable streaming
		cmdOptions := cmd.Options{
			Buffered:  false,
			Streaming: true,
		}

		// Create Cmd with options
		envCmd := cmd.NewCmdOptions(cmdOptions, command[0], command[1:]...)

		// Set environment variables for the command
		envCmd.Env = os.Environ()

		kubeconfigEnvVar := fmt.Sprintf("KUBECONFIG=%s", *tmpKubeconfigFile)
		envCmd.Env = append(envCmd.Env, kubeconfigEnvVar)

		// Print STDOUT and STDERR lines streaming from Cmd
		doneChan := make(chan struct{})
		go func() {
			defer close(doneChan)
			// Done when both channels have been closed
			// https://dave.cheney.net/2013/04/30/curious-channels
			for envCmd.Stdout != nil || envCmd.Stderr != nil {
				select {
				case line, open := <-envCmd.Stdout:
					if !open {
						envCmd.Stdout = nil
						continue
					}
					fmt.Println(line)
				case line, open := <-envCmd.Stderr:
					if !open {
						envCmd.Stderr = nil
						continue
					}
					fmt.Fprintln(os.Stderr, line)
				}
			}
		}()

		// Run and wait for Cmd to return, discard Status
		<-envCmd.Start()

		// Wait for goroutine to print everything
		<-doneChan

		fmt.Printf("=== END Executing on %s ===\n", context)
	}
	return nil
}
