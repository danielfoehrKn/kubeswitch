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

package migration

import (
	"k8s.io/utils/pointer"

	"github.com/danielfoehrkn/kubeswitch/types"
)

// ConvertConfiguration converts the old configuration to the new format
func ConvertConfiguration(old types.ConfigOld) types.Config {
	config := types.Config{
		Kind:              "SwitchConfig",
		Version:           "v1alpha1",
		RefreshIndexAfter: old.KubeconfigRediscoveryInterval,
		Hooks:             old.Hooks,
	}

	if len(old.KubeconfigName) > 0 {
		config.KubeconfigName = &old.KubeconfigName
	}

	filesystemStore := types.KubeconfigStore{
		ID:    pointer.StringPtr("default"),
		Kind:  types.StoreKindFilesystem,
		Paths: []string{},
	}

	vaultStore := types.KubeconfigStore{
		ID:    pointer.StringPtr("default"),
		Kind:  types.StoreKindVault,
		Paths: []string{},
	}

	if len(old.VaultAPIAddress) > 0 {
		vaultStore.Config = types.StoreConfigVault{
			VaultAPIAddress: old.VaultAPIAddress,
		}
	}

	for _, path := range old.KubeconfigPaths {
		switch path.Store {
		case types.StoreKindFilesystem:
			filesystemStore.Paths = append(filesystemStore.Paths, path.Path)
		case types.StoreKindVault:
			vaultStore.Paths = append(vaultStore.Paths, path.Path)
		}
	}

	if len(filesystemStore.Paths) > 0 {
		config.KubeconfigStores = append(config.KubeconfigStores, filesystemStore)
	}

	if len(vaultStore.Paths) > 0 {
		config.KubeconfigStores = append(config.KubeconfigStores, vaultStore)
	}

	return config
}
