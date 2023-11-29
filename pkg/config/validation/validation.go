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

package validation

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	gardenerstore "github.com/danielfoehrkn/kubeswitch/pkg/store/gardener"
	gkestore "github.com/danielfoehrkn/kubeswitch/pkg/store/gke"
	"github.com/danielfoehrkn/kubeswitch/types"
)

// ValidateConfig validates the SwitchConfig
func ValidateConfig(config *types.Config) field.ErrorList {
	var (
		errors     = field.ErrorList{}
		storeKinds = sets.Set[string]{}
		storesPath = field.NewPath("kubeconfigStores")
		usesIndex  = false
	)

	if config.RefreshIndexAfter != nil {
		usesIndex = true
	}

	if !types.ValidConfigVersions.Has(config.Version) {
		errors = append(errors, field.Invalid(field.NewPath("version"), config.Version, fmt.Sprintf("Config version %q is unknown. Valid versions are %q", config.Version, types.ValidConfigVersions)))
	}

	for i, kubeconfigStore := range config.KubeconfigStores {
		id := kubeconfigStore.ID
		if kubeconfigStore.ID == nil {
			emtpy := ""
			id = &emtpy
		}

		storeUsesIndex := usesIndex
		if kubeconfigStore.RefreshIndexAfter != nil {
			storeUsesIndex = true
		}

		indexFieldPath := storesPath.Index(i)

		if !types.ValidStoreKinds.Has(string(kubeconfigStore.Kind)) {
			errors = append(errors, field.Invalid(indexFieldPath.Child("kind"), kubeconfigStore.Kind, fmt.Sprintf("kind %q of kubeconfig store is unknown. Valid kinds are %q", kubeconfigStore.Kind, types.ValidStoreKinds)))
		}

		if len(kubeconfigStore.Paths) == 0 &&
			(kubeconfigStore.Kind == types.StoreKindFilesystem ||
				kubeconfigStore.Kind == types.StoreKindVault) {
			errors = append(errors, field.Invalid(indexFieldPath.Child("paths"), "", "Must provide at least one path for the kubeconfig store."))
		}

		if kubeconfigStore.Kind == types.StoreKindGardener {
			landscapeName, errorList := gardenerstore.ValidateGardenerStoreConfiguration(indexFieldPath, kubeconfigStore)
			errors = append(errors, errorList...)

			// the Gardener landscape name is the default ID of the store
			if landscapeName != nil && len(*landscapeName) > 0 && kubeconfigStore.ID == nil {
				id = landscapeName
			}
		}

		if kubeconfigStore.Kind == types.StoreKindGKE {
			errorList := gkestore.ValidateGKEStoreConfiguration(indexFieldPath, kubeconfigStore)
			errors = append(errors, errorList...)
		}

		// if the kubeconfig store uses an index, we need to specify a unique ID for the kubeconfigStore to write a unique index file name
		if storeUsesIndex && storeKinds.Has(fmt.Sprintf("%s:%s", kubeconfigStore.Kind, *id)) {
			errors = append(errors, field.Invalid(indexFieldPath.Child("id"), id, fmt.Sprintf("there are multiple kubeconfig stores with the same Kind %q configured. "+
				"In the switch configuration file, please set a unique ID for the kubeconfig store",
				kubeconfigStore.Kind)))
		}

		storeKinds.Insert(fmt.Sprintf("%s:%s", kubeconfigStore.Kind, *id))
	}

	if len(config.Hooks) > 0 {
		errors = append(errors, validateHooks(field.NewPath("hooks"), config.Hooks)...)
	}

	return errors
}

// validateHooks validates hook configuration
func validateHooks(path *field.Path, hooks []types.Hook) field.ErrorList {
	var errors = field.ErrorList{}

	for i, hook := range hooks {
		if !types.ValidHookTypes.Has(string(hook.Type)) {
			errors = append(errors, field.Invalid(path.Index(i).Child("type"), hook.Type, fmt.Sprintf("Unknown hook type. Valid hook types are %q", types.ValidHookTypes)))
		}

		if hook.Type == types.HookTypeExecutable && hook.Path == nil {
			errors = append(errors, field.Required(path.Index(i).Child("path"), "Path to the hook executable has to be provided"))
		}

		if hook.Type == types.HookTypeInlineCommand && len(hook.Arguments) == 0 {
			errors = append(errors, field.Required(path.Index(i).Child("arguments"), "arguments have to be provided for a hook with an inline command"))
		}
	}
	return errors
}
