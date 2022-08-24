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

package kubeconfigutil

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// TemporaryKubeconfigDir is a constant for the directory where the switcher stores the temporary kubeconfig files
	TemporaryKubeconfigDir = "$HOME/.kube/.switch_tmp"
)

type Kubeconfig struct {
	path       string
	useTmpFile bool
	rootNode   *yaml.Node
}

// LoadCurrentKubeconfig loads the current kubeconfig
func LoadCurrentKubeconfig() (*Kubeconfig, error) {
	path, err := kubeconfigPath()
	if err != nil {
		return nil, err
	}
	return NewKubeconfigForPath(path)
}

// NewKubeconfigForPath creates a kubeconfig representation based on an existing kubeconfig
// given by the path argument
// This will overwrite the kubeconfig given by path when calling WriteKubeconfigFile()
func NewKubeconfigForPath(path string) (*Kubeconfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig file: %v", err)
	}
	return New(bytes, path, false)
}

// Create kubeconfig file in the temporary directory
func NewKubeconfig(kubeconfigData []byte) (*Kubeconfig, error) {
	return New(kubeconfigData, os.ExpandEnv(TemporaryKubeconfigDir), true)
}

// New creates a new Kubeconfig representation based on the given kubeconfig data
// the format is validated
func New(kubeconfigData []byte, path string, useTmpFile bool) (*Kubeconfig, error) {
	n := &yaml.Node{}
	if err := yaml.Unmarshal(kubeconfigData, n); err != nil {
		return nil, err
	}

	k := &Kubeconfig{
		rootNode:   n.Content[0],
		path:       path,
		useTmpFile: useTmpFile,
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
func (k *Kubeconfig) SetContext(currentContext, originalContextBeforeAlias string, prefix string) error {
	if len(originalContextBeforeAlias) > 0 {
		// currentContext variable already has an alias
		// get the original currentContext name to find and replace it with the alias

		// can originalContextBeforeAlias  still contain the prefix?
		//  - yes if this store is configured with prefix (FIX: then need to remove prefix!)
		//  - no if store disabled prefix (works today)
		if len(prefix) > 0 && strings.HasPrefix(originalContextBeforeAlias, prefix) {
			originalContextBeforeAlias = strings.TrimPrefix(originalContextBeforeAlias, fmt.Sprintf("%s/", prefix))
		}

		if err := k.ModifyContextName(originalContextBeforeAlias, currentContext); err != nil {
			return fmt.Errorf("failed to set currentContext on selected kubeconfig: %v", err)
		}
	}

	if err := k.ModifyCurrentContext(currentContext); err != nil {
		return fmt.Errorf("failed to set current context on selected kubeconfig: %v", err)
	}
	return nil
}

func (k *Kubeconfig) SetKubeswitchContext(context string) error {
	if err := k.ModifyKubeswitchContext(context); err != nil {
		return fmt.Errorf("failed to set switch context on selected kubeconfig: %v", err)
	}
	return nil
}

// SetGardenerStoreMetaInformation is a function to add meta information to kubeconfig which is required for subsequent runs of kubeswitch
// Only relevant to the Gardener store
func (k *Kubeconfig) SetGardenerStoreMetaInformation(landscapeIdentity, clusterType, project, name string) error {
	if err := k.ModifyGardenerLandscapeIdentity(landscapeIdentity); err != nil {
		return fmt.Errorf("failed to set Gardener meta information (Landscape Identity): %v", err)
	}

	if err := k.ModifyGardenerClusterType(clusterType); err != nil {
		return fmt.Errorf("failed to set Gardener meta information (Cluster Type): %v", err)
	}

	if err := k.ModifyGardenerProject(project); err != nil {
		return fmt.Errorf("failed to set Gardener meta information (project name): %v", err)
	}

	if err := k.ModifyGardenerClusterName(name); err != nil {
		return fmt.Errorf("failed to set Gardener meta information (Shoot/Seed name): %v", err)
	}
	return nil
}

func (k *Kubeconfig) SetNamespaceForCurrentContext(namespace string) error {
	currentContext := k.GetCurrentContext()
	if len(currentContext) == 0 {
		return fmt.Errorf("current-context is not set")
	}

	if err := k.SetNamespace(currentContext, namespace); err != nil {
		return fmt.Errorf("failed to set namespace %q: %v", namespace, err)
	}

	return nil
}

// WriteKubeconfigFile writes kubeconfig bytes to the local filesystem
// and returns the kubeconfig path
func (k *Kubeconfig) WriteKubeconfigFile() (string, error) {
	var (
		file *os.File
		err  error
	)
	// if we do not use a tmp file, then k.path is the path to a directory to create the tmp file in
	if k.useTmpFile {
		err = os.Mkdir(k.path, 0700)
		if err != nil && !os.IsExist(err) {
			return "", err
		}

		// write temporary kubeconfig file
		file, err = os.CreateTemp(k.path, "config.*.tmp")
		if err != nil {
			return "", err
		}
	} else {
		file, err = os.OpenFile(k.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to open existing kubeconfig file: %v", err)
		}
	}

	enc := yaml.NewEncoder(file)
	enc.SetIndent(0)

	if err := enc.Encode(k.rootNode); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (k *Kubeconfig) GetBytes() ([]byte, error) {
	return yaml.Marshal(k.rootNode)
}

func kubeconfigPath() (string, error) {
	// KUBECONFIG env var
	if v := os.Getenv("KUBECONFIG"); v != "" {
		list := filepath.SplitList(v)
		if len(list) > 1 {
			// TODO KUBECONFIG=file1:file2 currently not supported
			return "", errors.New("multiple files in KUBECONFIG are currently not supported")
		}
		return v, nil
	}

	// default path
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("HOME environment variable not set")
	}
	return filepath.Join(home, ".kube", "config"), nil
}
