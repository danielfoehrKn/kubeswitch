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

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

// StoreKind identifies a supported store kind - filesystem, vault, Gardener.
type StoreKind string

// ValidStoreKinds contains all valid store kinds
var ValidStoreKinds = sets.NewString(string(StoreKindVault), string(StoreKindFilesystem), string(StoreKindGardener), string(StoreKindGKE))
// ValidConfigVersions contains all valid config versions
var ValidConfigVersions = sets.NewString("v1alpha1")

const (
	// StoreKindFilesystem is an identifier for the filesystem store
	StoreKindFilesystem StoreKind = "filesystem"
	// StoreKindVault is an identifier for the vault store
	StoreKindVault      StoreKind = "vault"
	// StoreKindGardener is an identifier for the gardener store
	StoreKindGardener   StoreKind = "gardener"
	// StoreKindGKE is an identifier for the GKE store
	StoreKindGKE   StoreKind = "gke"
)

type Config struct {
	// Kind is the type of the config. Expects "SwitchConfig"
	Kind                          string            `yaml:"kind"`
	// Version is the version of the config file.
	// Possible values: "v1alpha1"
	Version                        string          `yaml:"version"`
	// KubeconfigName is the global default for how the kubeconfig is
	// identified in the backing store.
	// Can be overridden in the individual kubeconfig store configuration
	// + optional
	KubeconfigName                *string            `yaml:"kubeconfigName"`
	// ShowPreview configures if the selection dialog shows a sanitized preview of the kubeconfig file.
	// Can be overridden via command line flag --show-preview true/false
	// default: true
	// + optional
	ShowPreview                *bool            `yaml:"showPreview"`
	// RefreshIndexAfter is the global default for how how often
	// the index for this kubeconfig store shall be refreshed.
	// Not setting this field will cause kubeswitch to not use an index
	// Can be overridden in the individual kubeconfig store configuration
	// + optional
	RefreshIndexAfter *time.Duration `yaml:"refreshIndexAfter"`
	// Hooks defines configurations for commands that shall be executed prior to the search
	Hooks                         []Hook            `yaml:"hooks"`
	// KubeconfigStores contains the configuration for kubeconfig stores
	KubeconfigStores              []KubeconfigStore `yaml:"kubeconfigStores"`
}

type KubeconfigStore struct {
	// ID is the ID of the kubeconfig store.
	// Used to write distinct index files for each store
	// Not required if only one store of a store kind is configured
	// + optional
	ID 					*string 			`yaml:"id"`
	// Kind identifies a supported store kind - filesystem, vault, Gardener.
	Kind                StoreKind        	`yaml:"kind"`
	// KubeconfigName defines how the kubeconfig is identified in the backing store
	// For the Filesystem store, this is the name of the file that contains the kubeconfig
	// For the Vault store, this is the secret key
	// For the Gardener store this field is not used
	// + optional
	KubeconfigName      *string           	`yaml:"kubeconfigName"`
	// Paths contains the paths to search for in the backing store
	Paths               []string         	`yaml:"paths"`
	// RefreshIndexAfter defines how often the index for this kubeconfig store shall be refreshed.
	// Not setting this field will cause kubeswitch to not use an index
	// + optional
	RefreshIndexAfter *time.Duration 		`yaml:"refreshIndexAfter"`
	// Required defines if errors when initializing this store should be logged
	// defaults to true
	// useful when configuring a kubeconfig store that is not always available
	// However, when searching on an index and wanting to retrieve the kubeconfig from an unavailable store,
	// it will throw an errors nonetheless
	// + optional
	Required *bool 						`yaml:"required"`
	// ShowPrefix configures if the search result should include store specific prefix (e.g for the filesystem store the parent directory name)
	// default: true
	ShowPrefix *bool `yaml:"showPrefix"`
	// Config is store-specific configuration.
	// Please check the documentation for each backing provider to see what configuration is
	// possible here
	Config              interface{} 		`yaml:"config"`
}

type StoreConfigVault struct {
	// VaultAPIAddress is the URL of the Vault API
	VaultAPIAddress  string `yaml:"vaultAPIAddress"`
}

type StoreConfigGardener struct {
	// GardenerAPIKubeconfigPath is the path on the local filesystem pointing to the kubeconfig
	// for the Gardener API server
	GardenerAPIKubeconfigPath  	string 	`yaml:"gardenerAPIKubeconfigPath"`
	// LandscapeName is a custom name for the Gardener landscape
	// uses this name instead of the default ID from the Gardener API ConfigMap "cluster-identity"
	// also used as the store ID if the kubeconfig store ID is not specified
	// + optional
	LandscapeName  				*string 	`yaml:"landscapeName"`
}

type StoreConfigGKE struct {
	// GKEAuthentication contains authentication configuration for GCP
	GKEAuthentication *GKEAuthentication `yaml:"gkeAuthentication"`
	// GCPAccount is the name of the gcp account kubeswitch shall discover GKE clusters from
	// Only used when relying on gcloud authentication.
	// Used to verify that gcloud currently has the correct account activated
	// However, will not actively activate another account (has to be done manually by the user).
	// If not specified, will use the currently activated account when using gcloud
	// + optional
	GCPAccount *string `yaml:"gcpAccount"`
	// ProjectID contains an optional list of projects that will be considered in the search for existing GKE clusters.
	// If no projects are given, will discover clusters from every found project.
	ProjectIDs  []string `yaml:"projectIDs"`
	// LandscapeName is a custom name for the GKE landscape
	// this name will be used during the search prefixing the cluster name to help distinguish
	// between GKE clusters from different GCP accounts
	// + optional
	LandscapeName  				*string 	`yaml:"landscapeName"`
}

// GCPAuthenticationType
// Required permission to list GKE clusters: container.clusters.list
// Requires to have the container.clusters.get permission. The least-privileged IAM role that provides this permission is container.clusterViewer.
type GCPAuthenticationType string

const (
	// GcloudAuthentication is an identifier for the gcloud authentication type that requires and uses a local installation
	// of the gcloud command line tool
	// Google Application Default Credentials are used for authentication.
	// When using gcloud  'gcloud auth application-default login' so that
	// the library can find a valid access token provided via gcloud's oauth flow at the default location
	// cat $HOME/.config/gcloud/application_default_credentials.json
	GcloudAuthentication GCPAuthenticationType = "gcloud"
	// APIKeyAuthentication is an identifier for the authentication type with API keys
	// also see: https://cloud.google.com/docs/authentication/api-keys
	APIKeyAuthentication GCPAuthenticationType = "api-key"
	// ServiceAccountAuthentication is an identifier for the authentication type with GCP service accounts
	// also see: https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication#environments-without-gcloud
	// To be able to use kubeswitch with multiple GCP accounts at once, please use Service Accounts and configure one
	// GKE store per account
	ServiceAccountAuthentication GCPAuthenticationType = "service-account"
	// LegacyAuthentication is an identifier for the gcloud authentication type with legacy credentials
	// also see: https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication#legacy-auth
	LegacyAuthentication GCPAuthenticationType = "legacy"
)

type GKEAuthentication struct {
	// possible values:
	// defaults to "gcloud"
	// + optional
	AuthenticationType *GCPAuthenticationType `yaml:"authenticationType"`
	// APIKeyFilePath is the path on the local filesystem to the file that contains
	// an API key used to authenticate against the Google Kubernetes Engine API
	// + optional
	APIKeyFilePath  *string `yaml:"apiKeyFilePath"`
	// ServiceAccountFilePath is the path on the local filesystem to the file that contains
	// the GCP service account used to authenticate against the Google Kubernetes Engine API
	// + optional
	ServiceAccountFilePath  *string `yaml:"serviceAccountFilePath"`
}
