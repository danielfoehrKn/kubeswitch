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

package ske

import (
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/danielfoehrkn/kubeswitch/types"
)

// ValidateSKEStoreConfiguration validates the store configuration for SKE.
// It is tested as part of the validation test suite.
func ValidateSKEStoreConfiguration(path *field.Path, store types.KubeconfigStore) field.ErrorList {
	var errors field.ErrorList

	if len(store.Paths) > 0 {
		errors = append(errors, field.Forbidden(path.Child("paths"), "Configuring paths for the SKE store is not allowed"))
	}

	configPath := path.Child("config")
	if store.Config == nil {
		errors = append(errors, field.Required(configPath, "The SKE store requires a config block with at least a projectID"))
		return errors
	}

	config, err := getStoreConfig(store)
	if err != nil {
		errors = append(errors, field.Invalid(configPath, store.Config, err.Error()))
		return errors
	}

	if config.ProjectID == "" {
		errors = append(errors, field.Required(configPath.Child("projectID"), "A STACKIT project ID must be provided"))
	}

	authMethods := 0
	if config.UseStackitCLIAuth {
		authMethods++
	}
	if config.ServiceAccountToken != "" {
		authMethods++
	}
	if config.ServiceAccountKeyPath != "" {
		authMethods++
	}
	if authMethods > 1 {
		errors = append(errors, field.Invalid(configPath, store.Config,
			"Only one authentication method may be specified: useStackitCLIAuth, serviceAccountToken, or serviceAccountKeyPath"))
	}

	return errors
}

// getStoreConfig unmarshals the raw store config into StoreConfigSKE.
func getStoreConfig(store types.KubeconfigStore) (*types.StoreConfigSKE, error) {
	buf, err := yaml.Marshal(store.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal store config: %w", err)
	}
	config := &types.StoreConfigSKE{}
	if err := yaml.Unmarshal(buf, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal store config: %w", err)
	}
	return config, nil
}
