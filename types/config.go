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

package types

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

// StoreKind identifies a supported store kind - filesystem, vault, Gardener.
type StoreKind string

// ValidStoreKinds contains all valid store kinds
var ValidStoreKinds = sets.NewString(string(StoreKindVault), string(StoreKindFilesystem), string(StoreKindGardener))

const (
	// StoreKindFilesystem is an identifier for the filesystem store
	StoreKindFilesystem StoreKind = "filesystem"
	// StoreKindFilesystem is an identifier for the vault store
	StoreKindVault      StoreKind = "vault"
	// StoreKindFilesystem is an identifier for the gardener store
	StoreKindGardener   StoreKind = "gardener"
)

type Config struct {
	// Kind is the type of the config. Expects "SwitchConfig"
	Kind                          string            `yaml:"kind"`
	// KubeconfigName is the global default for how the kubeconfig is
	// identified in the backing store.
	// Can be overridden in the individual kubeconfig store configuration
	// + optional
	KubeconfigName                *string            `yaml:"kubeconfigName"`
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
	RefreshIndexAfter *time.Duration `yaml:"refreshIndexAfter"`
	// Config is store-specific configuration.
	// Please check the documentation for each backing provider to see what confiuguration is
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
	// also used as the store ID instead of the kubeconfig store ID
	// + optional
	LandscapeName  				*string 	`yaml:"landscapeName"`
}
