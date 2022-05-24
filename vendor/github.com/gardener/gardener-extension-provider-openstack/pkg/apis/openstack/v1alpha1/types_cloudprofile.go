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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudProfileConfig contains provider-specific configuration that is embedded into Gardener's `CloudProfile`
// resource.
type CloudProfileConfig struct {
	metav1.TypeMeta `json:",inline"`
	// Constraints is an object containing constraints for certain values in the control plane config.
	Constraints Constraints `json:"constraints"`
	// DNSServers is a list of IPs of DNS servers used while creating subnets.
	// +optional
	DNSServers []string `json:"dnsServers,omitempty"`
	// DHCPDomain is the dhcp domain of the OpenStack system configured in nova.conf. Only meaningful for
	// Kubernetes 1.10.1+. See https://github.com/kubernetes/kubernetes/pull/61890 for details.
	// +optional
	DHCPDomain *string `json:"dhcpDomain,omitempty"`
	// KeyStoneURL is the URL for auth{n,z} in OpenStack (pointing to KeyStone).
	// +optional
	KeyStoneURL string `json:"keystoneURL,omitempty"`
	// KeyStoneURLs is a region-URL mapping for auth{n,z} in OpenStack (pointing to KeyStone).
	// +optional
	KeyStoneURLs []KeyStoneURL `json:"keystoneURLs,omitempty"`
	// MachineImages is the list of machine images that are understood by the controller. It maps
	// logical names and versions to provider-specific identifiers.
	MachineImages []MachineImages `json:"machineImages"`
	// RequestTimeout specifies the HTTP timeout against the OpenStack API.
	// +optional
	RequestTimeout *metav1.Duration `json:"requestTimeout,omitempty"`
	// RescanBlockStorageOnResize specifies whether the storage plugin scans and checks new block device size before it resizes
	// the filesystem.
	// +optional
	RescanBlockStorageOnResize *bool `json:"rescanBlockStorageOnResize,omitempty"`
	// IgnoreVolumeAZ specifies whether the volumes AZ should be ignored when scheduling to nodes,
	// to allow for differences between volume and compute zone naming. This setting only works for
	// shoots with kubernetes version 1.20.x or newer.
	// +optional
	IgnoreVolumeAZ *bool `json:"ignoreVolumeAZ,omitempty"`
	// NodeVolumeAttachLimit specifies how many volumes can be attached to a node.
	// +optional
	NodeVolumeAttachLimit *int32 `json:"nodeVolumeAttachLimit,omitempty"`
	// UseOctavia specifies whether the OpenStack Octavia network load balancing is used.
	// +optional
	UseOctavia *bool `json:"useOctavia,omitempty"`
	// UseSNAT specifies whether S-NAT is supposed to be used for the Gardener managed OpenStack router.
	// +optional
	UseSNAT *bool `json:"useSNAT,omitempty"`
	// ServerGroupPolicies specify the allowed server group policies for worker groups.
	// +optional
	ServerGroupPolicies []string `json:"serverGroupPolicies,omitempty"`
	// ResolvConfOptions specifies options to be added to /etc/resolv.conf on workers
	// +optional
	ResolvConfOptions []string `json:"resolvConfOptions,omitempty"`
}

// Constraints is an object containing constraints for the shoots.
type Constraints struct {
	// FloatingPools contains constraints regarding allowed values of the 'floatingPoolName' block in the control plane config.
	FloatingPools []FloatingPool `json:"floatingPools"`
	// LoadBalancerProviders contains constraints regarding allowed values of the 'loadBalancerProvider' block in the control plane config.
	LoadBalancerProviders []LoadBalancerProvider `json:"loadBalancerProviders"`
}

// FloatingPool contains constraints regarding allowed values of the 'floatingPoolName' block in the control plane config.
type FloatingPool struct {
	// Name is the name of the floating pool.
	Name string `json:"name"`
	// Region is the region name.
	// +optional
	Region *string `json:"region,omitempty"`
	// Domain is the domain name.
	// +optional
	Domain *string `json:"domain,omitempty"`
	// DefaultFloatingSubnet is the default floating subnet for the floating pool.
	// +optional
	DefaultFloatingSubnet *string `json:"defaultFloatingSubnet,omitempty"`
	// NonConstraining specifies whether this floating pool is not constraining, that means additionally available independent of other FP constraints.
	// +optional
	NonConstraining *bool `json:"nonConstraining,omitempty"`
	// LoadBalancerClasses contains a list of supported labeled load balancer network settings.
	// +optional
	LoadBalancerClasses []LoadBalancerClass `json:"loadBalancerClasses,omitempty"`
}

// KeyStoneURL is a region-URL mapping for auth{n,z} in OpenStack (pointing to KeyStone).
type KeyStoneURL struct {
	// Region is the name of the region.
	Region string `json:"region"`
	// URL is the keystone URL.
	URL string `json:"url"`
}

// LoadBalancerClass defines a restricted network setting for generic LoadBalancer classes.
type LoadBalancerClass struct {
	// Name is the name of the LB class
	Name string `json:"name"`
	// Purpose is reflecting if the loadbalancer class has a special purpose e.g. default, internal.
	// +optional
	Purpose *string `json:"purpose"`
	// FloatingSubnetID is the subnetwork ID of a dedicated subnet in floating network pool.
	// +optional
	FloatingSubnetID *string `json:"floatingSubnetID,omitempty"`
	// FloatingSubnetTags is a list of tags which can be used to select subnets in the floating network pool.
	// +optional
	FloatingSubnetTags *string `json:"floatingSubnetTags,omitempty"`
	// FloatingSubnetName is can either be a name or a name pattern of a subnet in the floating network pool.
	// +optional
	FloatingSubnetName *string `json:"floatingSubnetName,omitempty"`
	// FloatingNetworkID is the network ID of the floating network pool.
	// +optional
	FloatingNetworkID *string `json:"floatingNetworkID,omitempty"`
	// SubnetID is the ID of a local subnet used for LoadBalancer provisioning. Only usable if no FloatingPool
	// configuration is done.
	// +optional
	SubnetID *string `json:"subnetID,omitempty"`
}

// LoadBalancerProvider contains constraints regarding allowed values of the 'loadBalancerProvider' block in the control plane config.
type LoadBalancerProvider struct {
	// Name is the name of the load balancer provider.
	Name string `json:"name"`
	// Region is the region name.
	// +optional
	Region *string `json:"region,omitempty"`
}

// MachineImages is a mapping from logical names and versions to provider-specific identifiers.
type MachineImages struct {
	// Name is the logical name of the machine image.
	Name string `json:"name"`
	// Versions contains versions and a provider-specific identifier.
	Versions []MachineImageVersion `json:"versions"`
}

// MachineImageVersion contains a version and a provider-specific identifier.
type MachineImageVersion struct {
	// Version is the version of the image.
	Version string `json:"version"`
	// Image is the name of the image.
	Image string `json:"image,omitempty"`
	// Regions is an optional mapping to the correct Image ID for the machine image in the supported regions.
	Regions []RegionIDMapping `json:"regions,omitempty"`
}

// RegionIDMapping is a mapping to the correct ID for the machine image in the given region.
type RegionIDMapping struct {
	// Name is the name of the region.
	Name string `json:"name"`
	// ID is the ID for the machine image in the given region.
	ID string `json:"id"`
}
