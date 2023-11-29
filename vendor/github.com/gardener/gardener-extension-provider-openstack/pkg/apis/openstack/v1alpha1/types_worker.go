// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package v1alpha1

import (
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerStatus contains information about created worker resources.
type WorkerStatus struct {
	metav1.TypeMeta `json:",inline"`

	// MachineImages is a list of machine images that have been used in this worker. Usually, the extension controller
	// gets the mapping from name/version to the provider-specific machine image data in its componentconfig. However, if
	// a version that is still in use gets removed from this componentconfig it cannot reconcile anymore existing `Worker`
	// resources that are still using this version. Hence, it stores the used versions in the provider status to ensure
	// reconciliation is possible.
	// +optional
	MachineImages []MachineImage `json:"machineImages,omitempty"`

	// ServerGroupDependencies is a list of external server group dependencies.
	// +optional
	ServerGroupDependencies []ServerGroupDependency `json:"serverGroupDependencies,omitempty"`
}

// MachineImage is a mapping from logical names and versions to provider-specific machine image data.
type MachineImage struct {
	// Name is the logical name of the machine image.
	Name string `json:"name"`
	// Version is the logical version of the machine image.
	Version string `json:"version"`
	// Image is the name of the image.
	Image string `json:"image,omitempty"`
	// ID is the id of the image. (one of Image or ID must be set)
	ID string `json:"id,omitempty"`
}

// ServerGroupDependency is a reference to an external machine dependency of OpenStack server groups.
type ServerGroupDependency struct {
	// PoolName identifies the worker pool that this dependency belongs
	PoolName string `json:"poolName"`
	// ID is the provider's generated ID for a server group
	ID string `json:"id"`
	// Name is the name of the server group
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerConfig contains configuration data for a worker pool.
type WorkerConfig struct {
	metav1.TypeMeta `json:",inline"`

	// NodeTemplate contains resource information of the machine which is used by Cluster Autoscaler to generate
	// nodeTemplate during scaling a nodeGroup from zero.
	NodeTemplate *extensionsv1alpha1.NodeTemplate `json:"nodeTemplate,omitempty"`
	// ServerGroup contains configuration data for the worker pool's server group. If this object is present,
	// OpenStack provider extension will try to create a new server group for instances of this worker pool.
	ServerGroup *ServerGroup `json:"serverGroup,omitempty"`

	// MachineLabels define key value pairs to add to machines.
	MachineLabels []MachineLabel `json:"machineLabels,omitempty"`
}

// MachineLabel define key value pair to label machines.
type MachineLabel struct {
	// Name is the machine label key
	Name string `json:"name"`
	// Value is the machine label value
	Value string `json:"value"`
	// TriggerRollingOnUpdate controls if the machines should be rolled if the value changes
	TriggerRollingOnUpdate bool `json:"triggerRollingOnUpdate,omitempty"`
}

// ServerGroup contains configuration data for setting up a server group.
type ServerGroup struct {
	// Policy describes the kind of affinity policy for instances of the server group.
	// https://docs.openstack.org/python-openstackclient/ussuri/cli/command-objects/server-group.html
	Policy string `json:"policy"`
}
