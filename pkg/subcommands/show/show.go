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
package show

import (
	"fmt"

	"github.com/danielfoehrkn/kubeswitch/pkg"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
)

func Show(desiredName string, stores []storetypes.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) ([]byte, error) {
	c, err := pkg.DoSearch(stores, config, stateDir, noIndex)
	if err != nil {
		return nil, fmt.Errorf("cannot list contexts: %v", err)
	}

	for discoveredContext := range *c {
		if discoveredContext.Error != nil {
			continue
		}

		name := discoveredContext.Name
		if discoveredContext.Alias != "" {
			name = discoveredContext.Alias
		}

		if name == desiredName {
			store := *discoveredContext.Store
			kubeconfigData, err := store.GetKubeconfigForPath(discoveredContext.Path, discoveredContext.Tags)
			if err != nil {
				return nil, fmt.Errorf("failed to get kubeconfig: %v", err)
			}
			return kubeconfigData, nil
		}
	}

	return nil, fmt.Errorf("context not found")
}
