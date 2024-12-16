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
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	gardenerstore "github.com/danielfoehrkn/kubeswitch/pkg/store/gardener"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	historyutil "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history/util"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
	"github.com/danielfoehrkn/kubeswitch/types"
)

const (
	defaultKubeconfigPath       = "$HOME/.kube/config"
	linuxEnvKubeconfigSeperator = ":"
)

var (
	logger                = logrus.New()
	kubeconfigPathFromEnv = os.Getenv("KUBECONFIG")
)

func SwitchToControlplane(stores []storetypes.KubeconfigStore, kubeconfigPathFromFlag string) (*string, error) {
	kubeconfig, err := getKubeconfig(kubeconfigPathFromFlag)
	if err != nil {
		return nil, err
	}

	if !kubeconfig.IsGardenerKubeconfig() {
		return nil, fmt.Errorf("the currently used kubeconfig is not a Gardener kubeconfig. Cannot switch to controlplane. Please make sure you have previously used kubeswitch and the Gardener kuebconfig store to obtain the Shoot kubeconfig")
	}

	clusterType := kubeconfig.GetGardenerClusterType()
	if len(clusterType) == 0 {
		return nil, fmt.Errorf("the cluster type (SHoot/Seed) must be set as metadata in the kubeconfig")
	}

	if clusterType != string(gardenerstore.GardenerResourceShoot) {
		return nil, fmt.Errorf("cannot switch to the controlplane for %s clusters", clusterType)
	}

	clusterName := kubeconfig.GetGardenerClusterName()
	if len(clusterName) == 0 {
		return nil, fmt.Errorf("the cluster name must be set as metadata in the kubeconfig")
	}

	project := kubeconfig.GetGardenerProject()
	if len(project) == 0 {
		return nil, fmt.Errorf("the Gardener project name must be set as metadata in the kubeconfig")
	}

	landscapeIdentity := kubeconfig.GetGardenerLandscapeIdentity()
	if len(landscapeIdentity) == 0 {
		return nil, fmt.Errorf("the Gardener landscape identity must be set as metadata in the kubeconfig")
	}

	foundCorrectStore := false
	var targetStore *store.GardenerStore
	for _, kubeconfigStore := range stores {
		if kubeconfigStore.GetKind() == types.StoreKindGardener {
			gardenerStore, ok := kubeconfigStore.(*store.GardenerStore)
			if !ok {
				return nil, fmt.Errorf("internal error")
			}

			if !gardenerStore.IsInitialized() {
				if err := gardenerStore.InitializeGardenerStore(); err != nil {
					if gardenerStore.GetStoreConfig().Required != nil && !*gardenerStore.GetStoreConfig().Required {
						continue
					}
					return nil, fmt.Errorf("failed to initialize Gardener store with ID %q: %v", kubeconfigStore.GetID(), err)
				}
			}

			if landscapeIdentity == gardenerStore.LandscapeIdentity {
				foundCorrectStore = true
				targetStore = gardenerStore
				break
			}
		}
	}

	if !foundCorrectStore {
		return nil, fmt.Errorf("unable to find Seed for Shoot. Landscape identity %q not found", landscapeIdentity)
	}

	kubeconfigBytes, seedName, err := targetStore.GetControlplaneKubeconfigForShoot(clusterName, project)
	if err != nil {
		return nil, fmt.Errorf("gardener store returned an error obtaining the kubeconfig for the Shoot's Seed cluster: %v", err)
	}

	// append to history
	kubeconfig, err = kubeconfigutil.NewKubeconfig(kubeconfigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse obtained Seed kubeconfig: %v", err)
	}

	tempKubeconfigPath, err := kubeconfig.WriteKubeconfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to write temporary kubeconfig file: %v", err)
	}

	// get namespace for current context
	ns, err := kubeconfig.NamespaceOfContext(kubeconfig.GetCurrentContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace of current context: %v", err)
	}

	seedPath := gardenerstore.GetSeedIdentifier(targetStore.LandscapeName, *seedName)
	context := targetStore.GetContextPrefix(seedPath)
	context = fmt.Sprintf("%s/%s", context, kubeconfig.GetCurrentContext())

	if err := historyutil.AppendToHistory(context, ns); err != nil {
		logger.Warnf("failed to append context to history file: %v", err)
	}

	// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG environment Variable
	fmt.Print(tempKubeconfigPath)

	return &tempKubeconfigPath, nil
}

func getKubeconfig(kubeconfigPathFromFlag string) (*kubeconfigutil.Kubeconfig, error) {
	kubeconfigPath := kubeconfigPathFromFlag

	// kubeconfig path from flag is preferred over env (just not if it is only the default)
	if (len(kubeconfigPath) == 0 || kubeconfigPath == os.ExpandEnv(defaultKubeconfigPath)) && len(kubeconfigPathFromEnv) > 0 {
		if len(strings.Split(kubeconfigPathFromEnv, linuxEnvKubeconfigSeperator)) > 1 {
			return nil, fmt.Errorf("providing multiple kubeconfig files via environment variable KUBECONFIG is not supported")
		}

		kubeconfigPath = os.ExpandEnv(kubeconfigPathFromEnv)
	}

	if _, err := os.Stat(kubeconfigPath); err != nil {
		return nil, fmt.Errorf("unable to list namespaces. The kubeconfig file %q does not exist", kubeconfigPath)
	}

	return kubeconfigutil.NewKubeconfigForPath(kubeconfigPath)
}
