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

package gke

import (
	"fmt"

	"github.com/danielfoehrkn/kubeswitch/types"
	"gopkg.in/yaml.v3"
)

// GetStoreConfig unmarshalls to the Gardener store config from the configuration
func GetStoreConfig(store types.KubeconfigStore) (*types.StoreConfigGKE, error) {
	if store.Config == nil {
		return nil, fmt.Errorf("providing a configuration for the Gardener store is required. Please configure your SwitchConfig file properly")
	}

	storeConfig := &types.StoreConfigGKE{}
	buf, err := yaml.Marshal(store.Config)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, storeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config for the GKE kubeconfig store: %w", err)
	}
	return storeConfig, nil
}
