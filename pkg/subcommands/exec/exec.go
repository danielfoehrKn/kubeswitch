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
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	list_contexts "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/list-contexts"
	setcontext "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/set-context"
	"github.com/danielfoehrkn/kubeswitch/types"
)

func ExecuteCommand(pattern string, command []string, stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool, showDebugLogs bool) error {
	contexts, err := list_contexts.ListContexts(pattern, stores, config, stateDir, noIndex)
	if err != nil {
		return err
	}

	timestampedLogger := logrus.New()
	timestampedLogger.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%time%] %msg%",
	})

	plainLogger := logrus.New()
	plainLogger.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%",
	})

	// uses the standard plaintext logger
	standardLogger := logrus.New()

	if showDebugLogs {
		standardLogger.SetLevel(logrus.DebugLevel)
	}

	for _, context := range contexts {
		tmpKubeconfigFile, _, err := setcontext.SetContext(context, stores, config, stateDir, noIndex, false)
		if err != nil {
			return err
		}

		timestampedLogger.Printf("=== START Executing on %s ===\n", context)

		// Disable output buffering, enable streaming
		cmdOptions := cmd.Options{
			Buffered:  false,
			Streaming: true,
		}

		var envCmd *cmd.Cmd

		// Create Cmd with options
		if config != nil && config.ExecShell != nil {
			cmdArgument := ""
			for _, s := range command {
				cmdArgument = fmt.Sprintf("%s %s", cmdArgument, s)
			}
			args := append([]string{"-c"}, cmdArgument)

			envCmd = cmd.NewCmdOptions(cmdOptions, *config.ExecShell, args...)
			standardLogger.Debugf("Executing: \"%s -c %s\" \n", *config.ExecShell, cmdArgument)
		} else {
			envCmd = cmd.NewCmdOptions(cmdOptions, command[0], command[1:]...)
			standardLogger.Debugf("Executing: \"%s %s\"", command[0], command[1:])
		}

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
					plainLogger.Infof("%s \n", line)
				case line, open := <-envCmd.Stderr:
					if !open {
						envCmd.Stderr = nil
						continue
					}
					standardLogger.Errorf("%s \n", line)
				}
			}
		}()

		// Run and wait for Cmd to return, discard Status
		<-envCmd.Start()

		// Wait for goroutine to print everything
		<-doneChan

		timestampedLogger.Infof("=== END Executing on %s ===\n", context)
	}
	return nil
}
