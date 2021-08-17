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

package gardener

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/danielfoehrkn/kubeswitch/types"
)

var allowedPaths = sets.NewString("/", "garden")

// ValidateGardenerStoreConfiguration validates the store configuration for Gardener
// returns the optional landscape name as well as the error list
// is being tested as part of the validation test suite
func ValidateGardenerStoreConfiguration(path *field.Path, store types.KubeconfigStore) (*string, field.ErrorList) {
	var errors = field.ErrorList{}

	// always find the kubeconfigs of all Shoots on the landscape
	// in the future it could be restricted via paths to only certain namespaces
	for i, p := range store.Paths {
		if !allowedPaths.Has(p) && !strings.HasPrefix(p, "garden-") {
			errors = append(errors, field.Forbidden(path.Child("paths").Index(i), "Search paths can only equal \"garden\" or have the prefix \"garden-\""))
		}

		if p == "/" && len(store.Paths) > 1 {
			errors = append(errors, field.Forbidden(path.Child("paths").Index(i), "The search path \"/\" defines to search over all namespaces. You can either define \"/\" or other namespaces  (\"garden\" or have the prefix \"garden-\""))
		}
	}

	configPath := path.Child("config")
	if store.Config == nil {
		errors = append(errors, field.Required(configPath, "Missing configuration in the SwitchConfig file for the Gardener store"))
		return nil, errors
	}

	config, err := GetStoreConfig(store)
	if err != nil {
		errors = append(errors, field.Invalid(configPath, store.Config, err.Error()))
		return nil, errors
	}

	if len(config.GardenerAPIKubeconfigPath) == 0 {
		errors = append(errors, field.Invalid(configPath.Child("gardenerAPIKubeconfigPath"), config.GardenerAPIKubeconfigPath, "The kubeconfig to the Gardener API server must be set"))
	}

	if config.LandscapeName != nil && len(*config.LandscapeName) == 0 {
		errors = append(errors, field.Invalid(configPath.Child("landscapeName"), *config.LandscapeName, "The optional Gardener landscape name must not be empty"))
	}

	return config.LandscapeName, errors
}
