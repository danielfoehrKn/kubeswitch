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

package kubeconfigutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// TemporaryKubeconfigDir is a constant for the directory where the switcher stores the temporary kubeconfig files
	TemporaryKubeconfigDir = "$HOME/.kube/.switch_tmp"
)

type Kubeconfig struct {
	temporaryKubeconfigPath string
	rootNode                *yaml.Node
}

// GetContextWithoutFolderPrefix returns the real kubeconfig context name
// selectable kubeconfig context names have the folder prefixed like <parent-folder>/<context-name>
func GetContextWithoutFolderPrefix(path string) string {
	split := strings.SplitAfterN(path, "/", 2)
	return split[len(split)-1]
}

func ParseTemporaryKubeconfig(kubeconfigData []byte) (*Kubeconfig, error) {
	n := &yaml.Node{}
	if err := yaml.Unmarshal(kubeconfigData, n); err != nil {
		return nil, err
	}

	k := &Kubeconfig{
		rootNode:                n.Content[0],
		temporaryKubeconfigPath: os.ExpandEnv(TemporaryKubeconfigDir),
	}

	if k.rootNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("kubeconfig file does not have expected format")
	}
	return k, nil
}

// SetContext sets the given context as a current context on the kubeconfig
// if the given context is an alias, then it also modifies the kubeconfig so that both the current-context
// as well as the contexts.context name are set to the alias (otherwise the current-context
// would point to a non-existing context name)
func (k *Kubeconfig) SetContext(context, originalContextBeforeAlias string) error {
	if len(originalContextBeforeAlias) > 0 {
		// context has an alias
		// original context is always saved with the folder name in the alias state
		// remove the folder name to find and replace it in the kubeconfig with the alias
		originalContext := GetContextWithoutFolderPrefix(originalContextBeforeAlias)
		if err := k.ModifyContextName(originalContext, context); err != nil {
			return fmt.Errorf("failed to set alias context on selected kubeconfig: %v", err)
		}
	} else {
		// if selected context is not an alias, it does contain the folder name as
		// prefix which has to be removed
		context = GetContextWithoutFolderPrefix(context)
	}

	if err := k.ModifyCurrentContext(context); err != nil {
		return fmt.Errorf("failed to set current context on selected kubeconfig: %v", err)
	}
	return nil
}

// WriteTemporaryKubeconfigFile writes the temporary kubeconfig file to the local filesystem
// and returns the kubeconfig path
func (k *Kubeconfig) WriteTemporaryKubeconfigFile() (string, error) {
	err := os.Mkdir(k.temporaryKubeconfigPath, 0700)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	// write temporary kubeconfig file
	file, err := ioutil.TempFile(k.temporaryKubeconfigPath, "config.*.tmp")
	if err != nil {
		return "", err
	}

	enc := yaml.NewEncoder(file)
	enc.SetIndent(0)

	if err := enc.Encode(k.rootNode); err != nil {
		return "", err
	}

	return file.Name(), nil
}
