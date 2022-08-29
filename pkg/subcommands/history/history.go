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

package history

import (
	"bytes"
	"fmt"

	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history/util"
	setcontext "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/set-context"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func SwitchToHistory(stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) error {
	history, err := util.ReadHistory()
	if err != nil {
		return err
	}

	historyLength := len(history)

	idx, err := fuzzyfinder.Find(
		history,
		func(i int) string {
			// we expect a mapping context: namespace
			context, ns, err := util.ParseHistoryEntry(history[i])
			if err != nil {
				logger.Debugf("failed to parse namespace history entry")
				return ""
			}

			if ns == nil {
				return fmt.Sprintf("%d: %s", len(history)-i-1, *context)
			}

			if i+1 == historyLength {
				return fmt.Sprintf("%d: %s (ns: %s)", 0, *context, *ns)
			}

			previousContext, _, err := util.ParseHistoryEntry(history[i+1])
			if err != nil {
				logger.Debugf("failed to parse previous namespace history entry")
				return ""
			}

			// Grouping: check if the previous entry has the same context name
			// then only show the namespace
			if *context == *previousContext {
				unicodeCirceledStar := '\U0000272A'
				unicodeWhitespace := '\U00002009'

				// just to make sure that the namespace is shown in the terminal
				// window at the same position as the context
				var b bytes.Buffer
				n := len(history) - i - 1
				for n > 0 {
					n = n / 10
					b.WriteRune(unicodeWhitespace)
				}

				return fmt.Sprintf("%s%c %s", b.String(), unicodeCirceledStar, *ns)
			}

			return fmt.Sprintf("%d: %s (%s)", len(history)-i-1, *context, *ns)
		})

	if err != nil {
		return err
	}

	context, ns, err := util.ParseHistoryEntry(history[idx])
	if err != nil {
		return fmt.Errorf("failed to set namespace: %v", err)
	}

	// TODO: only switch context if the current context is not already set
	// requires to first check if a kubeconfig is already set (setcontext always creates a new file)
	// do not append to history as the old namespace will be added (only add history after changing the namespace)
	tmpKubeconfigFile, err := setcontext.SetContext(*context, stores, config, stateDir, noIndex, false)
	if err != nil {
		return err
	}

	// old history entry that does not include a namespace
	if ns == nil {
		return nil
	}

	if err := setNamespace(*ns, *tmpKubeconfigFile); err != nil {
		return err
	}

	return util.AppendToHistory(*context, *ns)
}

func setNamespace(ns string, tmpKubeconfigFile string) error {
	kubeconfig, err := kubeconfigutil.NewKubeconfigForPath(tmpKubeconfigFile)
	if err != nil {
		return err
	}

	if err := kubeconfig.SetNamespaceForCurrentContext(ns); err != nil {
		return fmt.Errorf("failed to set namespace %q: %v", ns, err)
	}

	if _, err := kubeconfig.WriteKubeconfigFile(); err != nil {
		return fmt.Errorf("failed to write namespace to kubeconfig %q: %v", ns, err)
	}

	return nil
}

// SetPreviousContext sets the previously used context from the history (position 1)
// does not add a history entry
func SetPreviousContext(stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) (*string, error) {
	history, err := util.ReadHistory()
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return nil, nil
	}

	var position int
	if len(history) == 1 {
		position = 0
	} else {
		position = 1
	}

	context, ns, err := util.ParseHistoryEntry(history[position])
	if err != nil {
		return nil, fmt.Errorf("failed to set previous context: %v", err)
	}

	tmpKubeconfigFile, err := setcontext.SetContext(*context, stores, config, stateDir, noIndex, false)
	if err != nil {
		return nil, err
	}

	// old history entry that does not include a namespace
	if ns == nil {
		return tmpKubeconfigFile, nil
	}

	return tmpKubeconfigFile, setNamespace(*ns, *tmpKubeconfigFile)
}

// SetLastContext sets the last used context from the history (position 0)
// does not add a history entry
func SetLastContext(stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) (*string, error) {
	history, err := util.ReadHistory()
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return nil, nil
	}

	context, ns, err := util.ParseHistoryEntry(history[0])
	if err != nil {
		return nil, fmt.Errorf("failed to set previous context: %v", err)
	}

	tmpKubeconfigFile, err := setcontext.SetContext(*context, stores, config, stateDir, noIndex, false)
	if err != nil {
		return nil, err
	}

	// old history entry that does not include a namespace
	if ns == nil {
		return tmpKubeconfigFile, nil
	}

	return tmpKubeconfigFile, setNamespace(*ns, *tmpKubeconfigFile)
}
