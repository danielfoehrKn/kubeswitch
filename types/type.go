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

package types

// KubeUser is a user in a kubeconfig file
type KubeUser struct {
	// Name is the name of the user
	// only parse the name, not the credentials
	Name string `yaml:"name"`
	// User contains user auth information
	User User `yaml:"user"`
}

// User contains the AuthInformation for a user
type User struct {
	// AuthProvider contains configuration for an external auth provider
	AuthProvider AuthProvider `yaml:"auth-provider"`
}
// AuthProvider cotnains the config and name of the kubeconfig auth provider plugin
type AuthProvider struct {
	// Cluster is the cluster identifier of the context
	Config map[string]string `yaml:"config"`
	// User is the user identifier of the context
	Name    string `yaml:"name"`
}

// KubeCluster is a cluster configuration of a kubeconfig file
type KubeCluster struct {
	// Name is the name of the cluster
	Name    string `yaml:"name"`
	// Cluster contains cluster configuration information
	Cluster Cluster `yaml:"cluster"`
}

type Cluster struct {
	// CertificateAuthorityData contains CA info
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
	// Server is the API server address
	Server                   string `yaml:"server"`
	// Insecure defines if the API server can be accessed with no CA checks
	Insecure                 bool   `yaml:"insecure-skip-tls-verify,omitempty"`
}

// KubeConfig is a representation of a kubeconfig file
// does not include sensitive fields that could include credentials
// used to show a preview of the Kubeconfig file
type KubeConfig struct {
	// TypeMeta common k8s type meta definition
	TypeMeta       TypeMeta `yaml:",inline"`
	// CurrentContext is the current context of the kubeconfig file
	CurrentContext string   `yaml:"current-context"`
	// Contexts are all defined contexts of the kubeconfig file
	Contexts []KubeContext `yaml:"contexts"`
	// Clusters are the cluster configurations
	Clusters []KubeCluster `yaml:"clusters"`
	// Users are the user configurations
	Users []KubeUser `yaml:"users"`
}

type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	Kind string `yaml:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
	// +optional
	APIVersion string `yaml:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}

type KubeContext struct {
	// Name is the name of the context
	Name    string `yaml:"name"`
	// Context contains context configuration
	Context Context `yaml:"context"`
}

type Context struct {
	// Cluster is the cluster identifier of the context
	Cluster string `yaml:"cluster"`
	// User is the user identifier of the context
	User    string
}
