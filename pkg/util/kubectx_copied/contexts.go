// THIS FILE HAS ORIGINALLY BEEN COPIED FROM THE KUBECTX PROJECT AND CONTAINS THE ORIGINAL LICENSE
// Copyright 2021 Google LLC
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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (k *Kubeconfig) contextsNode() (*yaml.Node, error) {
	contexts := valueOf(k.rootNode, "contexts")
	if contexts == nil {
		return nil, errors.New("\"contexts\" entry is nil")
	} else if contexts.Kind != yaml.SequenceNode {
		return nil, errors.New("\"contexts\" is not a sequence node")
	}
	return contexts, nil
}

func (k *Kubeconfig) contextNode(name string) (*yaml.Node, error) {
	contexts, err := k.contextsNode()
	if err != nil {
		return nil, err
	}

	for _, contextNode := range contexts.Content {
		nameNode := valueOf(contextNode, "name")
		if nameNode.Kind == yaml.ScalarNode && nameNode.Value == name {
			return contextNode, nil
		}
	}
	return nil, errors.Errorf("context with name \"%s\" not found", name)
}

// GetCurrentContext returns "current-context" value in given
// kubeconfig object Node, or returns "" if not found.
func (k *Kubeconfig) GetCurrentContext() string {
	v := valueOf(k.rootNode, "current-context")
	if v == nil {
		return ""
	}
	return v.Value
}

// GetContextNames returns all context names in the kubeconfig
func (k *Kubeconfig) GetContextNames() ([]string, error) {
	contexts, err := k.contextsNode()
	if err != nil {
		return nil, err
	}

	var contextNames []string
	for _, contextNode := range contexts.Content {
		contextName := valueOf(contextNode, "name")
		contextNames = append(contextNames, contextName.Value)
	}

	return contextNames, nil
}

// GetKubeswitchContext returns the "kubeswitch-context" value in given
// kubeconfig object Node, or returns "" if not found.
func (k *Kubeconfig) GetKubeswitchContext() string {
	v := valueOf(k.rootNode, "kubeswitch-context")
	if v == nil {
		return ""
	}
	return v.Value
}

// IsGardenerKubeconfig returns if this kubeconfig is a kubeconfig created by a kubeswitch Gardener Store
// i.e needs to contain meta information added previously by the gardener store
func (k *Kubeconfig) IsGardenerKubeconfig() bool {
	v := valueOf(k.rootNode, "gardener-landscape-identity")
	if v == nil {
		return false
	}
	return true
}

// GetGardenerLandscapeIdentity returns the "gardener-landscape-identity" value in given
// kubeconfig object Node, or returns "" if not found.
func (k *Kubeconfig) GetGardenerLandscapeIdentity() string {
	v := valueOf(k.rootNode, "gardener-landscape-identity")
	if v == nil {
		return ""
	}
	return v.Value
}

// GetGardenerProject returns the "gardener-project" value in given
// kubeconfig object Node, or returns "" if not found.
func (k *Kubeconfig) GetGardenerProject() string {
	v := valueOf(k.rootNode, "gardener-project")
	if v == nil {
		return ""
	}
	return v.Value
}

// GetGardenerClusterName returns the "gardener-cluster-name" value in given
// kubeconfig object Node, or returns "" if not found.
func (k *Kubeconfig) GetGardenerClusterName() string {
	v := valueOf(k.rootNode, "gardener-cluster-name")
	if v == nil {
		return ""
	}
	return v.Value
}

// GetGardenerClusterType returns the "gardener-cluster-type" value in given
// kubeconfig object Node, or returns "" if not found.
func (k *Kubeconfig) GetGardenerClusterType() string {
	v := valueOf(k.rootNode, "gardener-cluster-type")
	if v == nil {
		return ""
	}
	return v.Value
}

func valueOf(mapNode *yaml.Node, key string) *yaml.Node {
	if mapNode.Kind != yaml.MappingNode {
		return nil
	}
	for i, ch := range mapNode.Content {
		if i%2 == 0 && ch.Kind == yaml.ScalarNode && ch.Value == key {
			return mapNode.Content[i+1]
		}
	}
	return nil
}
