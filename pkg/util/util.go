// Copyright 2021 Daniel Foehr
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

package util

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/types"
)

// getContextsForKubeconfigPath takes a kubeconfig path and gets the kubeconfig from the backing store
// then it parses the kubeconfig to extract the context names
// returns the kubeconfig as a string as a first argument, and the context names as a second argument
func GetContextsForKubeconfigPath(kubeconfigBytes []byte, kind types.StoreKind, kubeconfigPath string) (*string, []string, error) {
	// parse into struct that does not contain the credentials
	config, err := ParseSanitizedKubeconfig(kubeconfigBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse Kubeconfig with path '%s': %v", kubeconfigPath, err)
	}

	kubeconfigData, err := yaml.Marshal(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not marshal kubeconfig with path '%s': %v", kubeconfigPath, err)
	}

	data := string(kubeconfigData)
	contextsFromKubeconfig, err := getContextsFromKubeconfig(kind, kubeconfigPath, config)
	return &data, contextsFromKubeconfig, err
}

// parseSanitizedKubeconfig parses the kubeconfig bytes into a kubeconfig struct without credentials
func ParseSanitizedKubeconfig(data []byte) (*types.KubeConfig, error) {
	config := types.KubeConfig{}

	// unmarshal in a form that does not include the credentials
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal kubeconfig: %v", err)
	}
	return &config, nil
}

// getContextsFromKubeconfig gets all the context names from the kubeconfig file
func getContextsFromKubeconfig(kind types.StoreKind, path string, kubeconfig *types.KubeConfig) ([]string, error) {
	parentFoldername := filepath.Base(filepath.Dir(path))
	if kind == types.StoreKindVault {
		// for vault, the secret name itself contains the semantic information (not the key of the kv-pair of the vault secret)
		parentFoldername = filepath.Base(path)
	} else if kind == types.StoreKindGardener {
		// the Gardener store encodes the path with semantic information
		// <landscape-name>--shoot-<project-name>--<shoot-name>
		parentFoldername = strings.ReplaceAll(path, "--", "-")
	}
	return getContextNames(kubeconfig, parentFoldername), nil
}

// sets the parent folder name to each context in the kubeconfig file
func getContextNames(config *types.KubeConfig, parentFoldername string) []string {
	var contextNames []string
	for _, context := range config.Contexts {
		split := strings.Split(context.Name, "/")
		if len(split) > 1 {
			// already has the directory name in there. override it in case it changed
			contextNames = append(contextNames, fmt.Sprintf("%s/%s", parentFoldername, split[len(split)-1]))
		} else {
			contextNames = append(contextNames, fmt.Sprintf("%s/%s", parentFoldername, context.Name))
		}
	}
	return contextNames
}
