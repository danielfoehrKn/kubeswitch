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

package util

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/types"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
)

// GetContextsNamesFromKubeconfig takes kubeconfig bytes and parses the kubeconfig to extract the context names.
// returns the kubeconfig as a string as a first argument, and the context names as a second argument
func GetContextsNamesFromKubeconfig(kubeconfigBytes []byte, contextPrefix string) (*string, []string, error) {
	// parse into struct that does not contain the credentials
	config, err := ParseSanitizedKubeconfig(kubeconfigBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse Kubeconfig: %v", err)
	}

	kubeconfigData, err := yaml.Marshal(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not marshal kubeconfig: %v", err)
	}

	data := string(kubeconfigData)
	contextsFromKubeconfig := getContextNames(config, contextPrefix)
	return &data, contextsFromKubeconfig, err
}

// ParseSanitizedKubeconfig parses the kubeconfig bytes into a kubeconfig struct without credentials
func ParseSanitizedKubeconfig(data []byte) (*types.KubeConfig, error) {
	config := types.KubeConfig{}

	// unmarshal in a form that does not include the credentials
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal kubeconfig: %v", err)
	}
	return &config, nil
}

// getContextNames gets all the context names from the kubeconfig file
// and sets the parent folder name to each context in the kubeconfig file
func getContextNames(config *types.KubeConfig, prefix string) []string {
	var contextNames []string

	// add a trailing slash if prefix is set (for path-like formatting)
	if len(prefix) != 0 {
		prefix = fmt.Sprintf("%s/", prefix)
	}

	for _, context := range config.Contexts {
		contextNames = append(contextNames, fmt.Sprintf("%s%s", prefix, context.Name))
	}
	return contextNames
}

// ExpandEnv takes a string and replaces all environment variables with their values
// ~ is expanded to the user's home directory
func ExpandEnv(path string) string {
	path = strings.ReplaceAll(path, "~", "$HOME")
	return os.ExpandEnv(path)
}

// GetCurrentContext returns "current-context" value of current kubeconfig
func GetCurrentContext() (string, error) {
	kc, err := kubeconfigutil.LoadCurrentKubeconfig()
	if err != nil {
		return "", err
	}
	currCtx := kc.GetCurrentContext()
	if currCtx == "" {
		return "", fmt.Errorf("current-context is not set")
	}
	return currCtx, nil
}
