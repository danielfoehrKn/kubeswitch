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

package openstack

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/gardener-extension-provider-openstack/pkg/utils"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudProfileConfig contains provider-specific configuration that is embedded into Gardener's `CloudProfile`
// resource.
type CloudProfileConfig struct {
	metav1.TypeMeta
	// Constraints is an object containing constraints for certain values in the control plane config.
	Constraints Constraints
	// DNSServers is a list of IPs of DNS servers used while creating subnets.
	DNSServers []string
	// DHCPDomain is the dhcp domain of the OpenStack system configured in nova.conf. Only meaningful for
	// Kubernetes 1.10.1+. See https://github.com/kubernetes/kubernetes/pull/61890 for details.
	DHCPDomain *string
	// KeyStoneURL is the URL for auth{n,z} in OpenStack (pointing to KeyStone).
	KeyStoneURL string
	// KeystoneCACert is the CA Bundle for the KeyStoneURL.
	KeyStoneCACert *string
	// KeyStoneForceInsecure is a flag to control whether the OpenStack client should perform no certificate validation.
	KeyStoneForceInsecure bool
	// KeyStoneURLs is a region-URL mapping for auth{n,z} in OpenStack (pointing to KeyStone).
	KeyStoneURLs []KeyStoneURL
	// MachineImages is the list of machine images that are understood by the controller. It maps
	// logical names and versions to provider-specific identifiers.
	MachineImages []MachineImages
	// RequestTimeout specifies the HTTP timeout against the OpenStack API.
	RequestTimeout *metav1.Duration
	// RescanBlockStorageOnResize specifies whether the storage plugin scans and checks new block device size before it resizes
	// the filesystem.
	RescanBlockStorageOnResize *bool
	// IgnoreVolumeAZ specifies whether the volumes AZ should be ignored when scheduling to nodes,
	// to allow for differences between volume and compute zone naming.
	IgnoreVolumeAZ *bool
	// NodeVolumeAttachLimit specifies how many volumes can be attached to a node.
	NodeVolumeAttachLimit *int32
	// UseOctavia specifies whether the OpenStack Octavia network load balancing is used.
	UseOctavia *bool
	// UseSNAT specifies whether S-NAT is supposed to be used for the Gardener managed OpenStack router.
	UseSNAT *bool
	// ServerGroupPolicies specify the allowed server group policies for worker groups.
	ServerGroupPolicies []string
	// ResolvConfOptions specifies options to be added to /etc/resolv.conf on workers
	ResolvConfOptions []string
	// StorageClasses defines storageclasses for the shoot
	// +optional
	StorageClasses []StorageClassDefinition
}

// Constraints is an object containing constraints for the shoots.
type Constraints struct {
	// FloatingPools contains constraints regarding allowed values of the 'floatingPoolName' block in the control plane config.
	FloatingPools []FloatingPool
	// LoadBalancerProviders contains constraints regarding allowed values of the 'loadBalancerProvider' block in the control plane config.
	LoadBalancerProviders []LoadBalancerProvider
}

// FloatingPool contains constraints regarding allowed values of the 'floatingPoolName' block in the control plane config.
type FloatingPool struct {
	// Name is the name of the floating pool.
	Name string
	// Region is the region name.
	Region *string
	// Domain is the domain name.
	Domain *string
	// DefaultFloatingSubnet is the default floating subnet for the floating pool.
	DefaultFloatingSubnet *string
	// NonConstraining specifies whether this floating pool is not constraining, that means additionally available independent of other FP constraints.
	NonConstraining *bool
	// LoadBalancerClasses contains a list of supported labeled load balancer network settings.
	LoadBalancerClasses []LoadBalancerClass
}

// KeyStoneURL is a region-URL mapping for auth{n,z} in OpenStack (pointing to KeyStone).
type KeyStoneURL struct {
	// Region is the name of the region.
	Region string
	// URL is the keystone URL.
	URL string
	// CACert is the CA Bundle for the KeyStoneURL.
	CACert *string
}

// LoadBalancerClass defines a restricted network setting for generic LoadBalancer classes.
type LoadBalancerClass struct {
	// Name is the name of the LB class
	Name string
	// Purpose is reflecting if the loadbalancer class has a special purpose e.g. default, internal.
	Purpose *string
	// FloatingSubnetID is the subnetwork ID of a dedicated subnet in floating network pool.
	FloatingSubnetID *string
	// FloatingSubnetTags is a list of tags which can be used to select subnets in the floating network pool.
	FloatingSubnetTags *string
	// FloatingSubnetName is can either be a name or a name pattern of a subnet in the floating network pool.
	FloatingSubnetName *string
	// FloatingNetworkID is the network ID of the floating network pool.
	FloatingNetworkID *string
	// SubnetID is the ID of a local subnet used for LoadBalancer provisioning. Only usable if no FloatingPool
	// configuration is done.
	SubnetID *string
}

// IsSemanticallyEqual checks if the load balancer class is semantically equal to
// another given load balancer class. Name and Purpose fields are allowed to be different.
func (l LoadBalancerClass) IsSemanticallyEqual(e LoadBalancerClass) bool {
	if !utils.StringEqual(l.FloatingNetworkID, e.FloatingNetworkID) {
		return false
	}
	if !utils.StringEqual(l.FloatingSubnetID, e.FloatingSubnetID) {
		return false
	}
	if !utils.StringEqual(l.FloatingSubnetName, e.FloatingSubnetName) {
		return false
	}
	if !utils.StringEqual(l.FloatingSubnetTags, e.FloatingSubnetTags) {
		return false
	}
	if !utils.StringEqual(l.SubnetID, e.SubnetID) {
		return false
	}
	return true
}

func (l LoadBalancerClass) String() string {
	result := fmt.Sprintf("Name: %q", l.Name)
	if l.Purpose != nil {
		result += fmt.Sprintf(", Purpose: %q", *l.Purpose)
	}
	if l.FloatingSubnetID != nil {
		result += fmt.Sprintf(", FloatingSubnetID: %q", *l.FloatingSubnetID)
	}
	if l.FloatingSubnetTags != nil {
		result += fmt.Sprintf(", FloatingSubnetTags: %q", *l.FloatingSubnetTags)
	}
	if l.FloatingSubnetName != nil {
		result += fmt.Sprintf(", FloatingSubnetName: %q", *l.FloatingSubnetName)
	}
	if l.FloatingNetworkID != nil {
		result += fmt.Sprintf(", FloatingNetworkID: %q", *l.FloatingNetworkID)
	}
	if l.SubnetID != nil {
		result += fmt.Sprintf(", SubnetID: %q", *l.SubnetID)
	}
	return result
}

// LoadBalancerProvider contains constraints regarding allowed values of the 'loadBalancerProvider' block in the control plane config.
type LoadBalancerProvider struct {
	// Name is the name of the load balancer provider.
	Name string
	// Region is the region name.
	Region *string
}

// MachineImages is a mapping from logical names and versions to provider-specific identifiers.
type MachineImages struct {
	// Name is the logical name of the machine image.
	Name string
	// Versions contains versions and a provider-specific identifier.
	Versions []MachineImageVersion
}

// MachineImageVersion contains a version and a provider-specific identifier.
type MachineImageVersion struct {
	// Version is the version of the image.
	Version string
	// Image is the name of the image.
	Image string
	// Regions is an optional mapping to the correct Image ID for the machine image in the supported regions.
	Regions []RegionIDMapping
}

// RegionIDMapping is a mapping to the correct ID for the machine image in the given region.
type RegionIDMapping struct {
	// Name is the name of the region.
	Name string
	// ID is the ID for the machine image in the given region.
	ID string
}

// StorageClassDefinition is a definition of a storageClass
type StorageClassDefinition struct {
	// Name is the name of the storageclass
	Name string
	// Default set the storageclass to the default one
	// +optional
	Default *bool
	// Provisioner set the Provisioner inside the storageclass
	// +optional
	Provisioner *string
	// Parameters adds parameters to the storageclass (storageclass.parameters)
	// +optional
	Parameters map[string]string
	// Annotations sets annotations for the storageclass
	// +optional
	Annotations map[string]string
	// Labels sets labels for the storageclass
	// +optional
	Labels map[string]string
	// ReclaimPolicy sets reclaimPolicy for the storageclass
	// +optional
	ReclaimPolicy *string
	// VolumeBindingMode sets bindingMode for the storageclass
	// +optional
	VolumeBindingMode *string
}
