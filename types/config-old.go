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

import "time"

// contains old configuration file format to facilitate upgrades with
// automatic config file conversion

type ConfigOld struct {
	Kind                          string           `yaml:"kind"`
	KubeconfigName                string           `yaml:"kubeconfigName"`
	KubeconfigRediscoveryInterval *time.Duration   `yaml:"kubeconfigRediscoveryInterval"`
	VaultAPIAddress               string           `yaml:"vaultAPIAddress"`
	Hooks                         []Hook           `yaml:"hooks"`
	KubeconfigPaths               []KubeconfigPath `yaml:"kubeconfigPaths"`
}

type KubeconfigPath struct {
	Path  string    `yaml:"path"`
	Store StoreKind `yaml:"store"`
}
