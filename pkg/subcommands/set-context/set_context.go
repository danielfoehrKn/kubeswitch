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

package setcontext

import (
	"fmt"
	"strings"

	"github.com/danielfoehrkn/kubeswitch/pkg"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	historyutil "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history/util"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func SetContext(desiredContext string, stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool, appendToHistory bool, writeToStdOut bool) (*string, error) {
	c, err := pkg.DoSearch(stores, config, stateDir, noIndex)
	if err != nil {
		return nil, err
	}

	var mError *multierror.Error
	for discoveredContext := range *c {
		if discoveredContext.Error != nil {
			// remember in case the wanted context name cannot be found
			mError = multierror.Append(mError, discoveredContext.Error)
			continue
		}

		if discoveredContext.Store == nil {
			// this should not happen
			logger.Debugf("store returned from search is nil. This should not happen")
			continue
		}

		kubeconfigStore := *discoveredContext.Store

		contextWithoutPrefix := discoveredContext.Name
		if len(kubeconfigStore.GetContextPrefix(discoveredContext.Path)) > 0 && strings.HasPrefix(discoveredContext.Name, kubeconfigStore.GetContextPrefix(discoveredContext.Path)) {
			contextWithoutPrefix = strings.TrimPrefix(discoveredContext.Name, fmt.Sprintf("%s/", kubeconfigStore.GetContextPrefix(discoveredContext.Path)))
		}

		matchesContextWithoutPrefix := desiredContext == contextWithoutPrefix
		if desiredContext == discoveredContext.Name || matchesContextWithoutPrefix || desiredContext == discoveredContext.Alias {
			kubeconfigData, err := kubeconfigStore.GetKubeconfigForPath(discoveredContext.Path)
			if err != nil {
				return nil, err
			}

			kubeconfig, err := kubeconfigutil.NewKubeconfig(kubeconfigData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kubeconfig: %v", err)
			}

			originalContextBeforeAlias := ""
			if len(discoveredContext.Alias) > 0 {
				originalContextBeforeAlias = contextWithoutPrefix
			}

			if err := kubeconfig.SetContext(contextWithoutPrefix, originalContextBeforeAlias, kubeconfigStore.GetContextPrefix(discoveredContext.Path)); err != nil {
				return nil, err
			}

			if err := kubeconfig.SetKubeswitchContext(desiredContext); err != nil {
				return nil, err
			}

			tempKubeconfigPath, err := kubeconfig.WriteKubeconfigFile()
			if err != nil {
				return nil, fmt.Errorf("failed to write temporary kubeconfig file: %v", err)
			}

			if appendToHistory {
				// get namespace for current context
				ns, err := kubeconfig.NamespaceOfContext(kubeconfig.GetCurrentContext())
				if err != nil {
					return nil, fmt.Errorf("failed to get namespace of current context: %v", err)
				}

				if err := historyutil.AppendToHistory(desiredContext, ns); err != nil {
					logger.Warnf("failed to append context to history file: %v", err)
				}
			}

			// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG environment Variable
			if writeToStdOut {
				fmt.Print(tempKubeconfigPath)
			}
			return &tempKubeconfigPath, nil
		}
	}

	if mError != nil {
		return nil, fmt.Errorf("context with name %q not found. Possibly due to errors: %v", desiredContext, mError.Error())
	}

	return nil, fmt.Errorf("context with name %q not found", desiredContext)
}
