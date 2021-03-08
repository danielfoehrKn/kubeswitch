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
)

type Config struct {
	Kind                          string            `yaml:"kind"`
	KubeconfigName                string            `yaml:"kubeconfigName"`
	KubeconfigRediscoveryInterval *time.Duration    `yaml:"rediscoveryInterval"`
	Hooks                         []Hook            `yaml:"hooks"`
	KubeconfigStores              []KubeconfigStore `yaml:"kubeconfigStores"`
}

type KubeconfigStore struct {
	Kind                StoreKind        	`yaml:"kind"`
	KubeconfigName      string           	`yaml:"kubeconfigName"`
	Paths               []string         	`yaml:"paths"`
	RediscoveryInterval *time.Duration   	`yaml:"rediscoveryInterval"`
	Config              interface{} 		`yaml:"config"`
}

type StoreConfigVault struct {
	VaultAPIAddress  string `yaml:"vaultAPIAddress"`
}

type StoreConfigGardener struct {
	GardenerAPIKubeconfigPath  	string 	`yaml:"gardenerAPIKubeconfigPath"`
	LandscapeName  				*string 	`yaml:"landscapeName"`
}
