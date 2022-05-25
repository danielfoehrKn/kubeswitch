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
	"fmt"

	"gopkg.in/yaml.v3"
)

// ModifyKubeswitchContext adds a top-level field with the key "kubeswitch-context" to the kubeconfig file.
// This context is the kubeswitch prefix (store dependent) / <kubeconfig-context>
// this is done when creating a new temporary copy of the kubeconfig file when "switching" to it
// During change of namespaces the current context information
func (k *Kubeconfig) ModifyKubeswitchContext(context string) error {
	currentCtxNode := valueOf(k.rootNode, "kubeswitch-context")
	if currentCtxNode != nil {
		currentCtxNode.Value = context
		return nil
	}

	// if kubeswitch-context field doesn't exist, create new field
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "kubeswitch-context",
		Tag:   "!!str"}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: context,
		Tag:   "!!str"}
	k.rootNode.Content = append(k.rootNode.Content, keyNode, valueNode)
	return nil
}

// ModifyGardenerLandscapeIdentity add a top-level field with the following identifiers to the kubeconfig file.
// - "landscape-identity"
// Only relevant for Gardener stores
func (k *Kubeconfig) ModifyGardenerLandscapeIdentity(gardenerLandscapeIdentity string) error {
	identityNode := valueOf(k.rootNode, "gardener-landscape-identity")
	if identityNode != nil {
		identityNode.Value = gardenerLandscapeIdentity
		return nil
	}

	// if kubeswitch-context field doesn't exist, create new field
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "gardener-landscape-identity",
		Tag:   "!!str"}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: gardenerLandscapeIdentity,
		Tag:   "!!str"}
	k.rootNode.Content = append(k.rootNode.Content, keyNode, valueNode)
	return nil
}

// ModifyGardenerClusterNamespace adds top-level fields with the following identifiers to the kubeconfig file.
// - "shoot-namespace"
// Only relevant for Gardener stores
func (k *Kubeconfig) ModifyGardenerClusterNamespace(namespace string) error {
	nsNode := valueOf(k.rootNode, "gardener-cluster-namespace")
	if nsNode != nil {
		nsNode.Value = namespace
		return nil
	}

	// if kubeswitch-context field doesn't exist, create new field
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "gardener-cluster-namespace",
		Tag:   "!!str"}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: namespace,
		Tag:   "!!str"}
	k.rootNode.Content = append(k.rootNode.Content, keyNode, valueNode)
	return nil
}

// ModifyGardenerClusterName adds top-level fields with the following identifiers to the kubeconfig file.
// - "shoot-name"
// Only relevant for Gardener stores
func (k *Kubeconfig) ModifyGardenerClusterName(name string) error {
	nameNode := valueOf(k.rootNode, "gardener-cluster-name")
	if nameNode != nil {
		nameNode.Value = name
		return nil
	}

	// if kubeswitch-context field doesn't exist, create new field
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "gardener-cluster-name",
		Tag:   "!!str"}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: name,
		Tag:   "!!str"}
	k.rootNode.Content = append(k.rootNode.Content, keyNode, valueNode)
	return nil
}

// ModifyGardenerClusterType adds top-level fields with the following identifiers to the kubeconfig file.
// - "gardener-cluster-type"  (e.g "Shoot" or "Seed")
// Only relevant for Gardener stores
func (k *Kubeconfig) ModifyGardenerClusterType(clusterType string) error {
	typeNode := valueOf(k.rootNode, "gardener-cluster-type")
	if typeNode != nil {
		typeNode.Value = clusterType
		return nil
	}

	// if kubeswitch-context field doesn't exist, create new field
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "gardener-cluster-type",
		Tag:   "!!str"}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: clusterType,
		Tag:   "!!str"}
	k.rootNode.Content = append(k.rootNode.Content, keyNode, valueNode)
	return nil
}

func (k *Kubeconfig) ModifyCurrentContext(name string) error {
	currentCtxNode := valueOf(k.rootNode, "current-context")
	if currentCtxNode != nil {
		currentCtxNode.Value = name
		return nil
	}

	// if current-context field doesn't exist, create new field
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "current-context",
		Tag:   "!!str"}

	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: name,
		Tag:   "!!str"}
	k.rootNode.Content = append(k.rootNode.Content, keyNode, valueNode)
	return nil
}

func (k *Kubeconfig) ModifyContextName(old, new string) error {
	contexts, err := k.contextsNode()
	if err != nil {
		return err
	}

	var changed bool
	for _, contextNode := range contexts.Content {
		nameNode := valueOf(contextNode, "name")
		if nameNode.Kind == yaml.ScalarNode && nameNode.Value == old {
			nameNode.Value = new
			changed = true
			break
		}
	}
	if !changed {
		return fmt.Errorf("context with name %q not found", old)
	}
	return nil
}

func (k *Kubeconfig) RemoveContext(name string) error {
	contexts := valueOf(k.rootNode, "contexts")
	if contexts == nil {
		return fmt.Errorf("contexts entry is nil")
	} else if contexts.Kind != yaml.SequenceNode {
		return fmt.Errorf("contexts is not a sequence node")
	}

	var keepContexts []*yaml.Node
	for _, contextNode := range contexts.Content {
		contextName := valueOf(contextNode, "name")
		if contextName.Value == name {
			continue
		}

		keepContexts = append(keepContexts, contextNode)
	}

	contexts.Content = keepContexts

	return nil
}
