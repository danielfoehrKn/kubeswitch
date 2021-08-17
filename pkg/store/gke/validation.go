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
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/danielfoehrkn/kubeswitch/types"
)

// ValidateGKEStoreConfiguration validates the store configuration for GKE
// returns the optional landscape name as well as the error list
// is being tested as part of the validation test suite
func ValidateGKEStoreConfiguration(path *field.Path, store types.KubeconfigStore) (*string, field.ErrorList) {
	var errors = field.ErrorList{}

	if len(store.Paths) > 0 {
		errors = append(errors, field.Forbidden(path.Child("paths"), "Configuring paths for the GKE store is not allowed"))
	}

	configPath := path.Child("config")
	// if there is no special store configuration, use default authentication
	// assumes it can find application default credentials created via gcloud auth login
	if store.Config == nil {
		return nil, errors
	}

	config, err := GetStoreConfig(store)
	if err != nil {
		errors = append(errors, field.Invalid(configPath, store.Config, err.Error()))
		return nil, errors
	}

	if config.GKEAuthentication != nil &&
		config.GCPAccount != nil &&
		config.GKEAuthentication.AuthenticationType != nil &&
		*config.GKEAuthentication.AuthenticationType != types.GcloudAuthentication {
		errors = append(errors, field.Invalid(configPath.Child("gcpAccount"), config.GCPAccount, "Can only specify a GCP account when using authentication via gcloud"))
	}

	if config.GKEAuthentication != nil &&
		config.GKEAuthentication.AuthenticationType != nil &&
		*config.GKEAuthentication.AuthenticationType == types.APIKeyAuthentication &&
		(config.GKEAuthentication.APIKeyFilePath == nil ||
			len(*config.GKEAuthentication.APIKeyFilePath) == 0) {
		errors = append(errors, field.Invalid(configPath.Child("gkeAuthentication").Child("apiKeyFilePath"), config.GCPAccount, "The filepath to the file containing thr GCP API key must be specified"))
	}

	if config.GKEAuthentication != nil &&
		config.GKEAuthentication.AuthenticationType != nil &&
		*config.GKEAuthentication.AuthenticationType == types.ServiceAccountAuthentication &&
		(config.GKEAuthentication.ServiceAccountFilePath == nil ||
			len(*config.GKEAuthentication.ServiceAccountFilePath) == 0) {
		errors = append(errors, field.Invalid(configPath.Child("gkeAuthentication").Child("serviceAccountFilePath"), config.GCPAccount, "The filepath to the file containing thr GCP service account must be specified"))
	}

	if config.LandscapeName != nil && len(*config.LandscapeName) == 0 {
		errors = append(errors, field.Invalid(configPath.Child("landscapeName"), *config.LandscapeName, "The optional GKE landscape name must not be empty"))
	}

	return config.LandscapeName, errors
}
