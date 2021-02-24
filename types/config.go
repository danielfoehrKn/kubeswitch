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

import "time"

type HookType string

const (
	HookTypeExecutable    HookType = "Executable"
	HookTypeInlineCommand HookType = "InlineCommand"
)

type Config struct {
	Kind                          string           `yaml:"kind"`
	KubeconfigRediscoveryInterval *time.Duration   `yaml:"kubeconfigRediscoveryInterval"`
	VaultAPIAddress               string           `yaml:"vaultAPIAddress"`
	Hooks                         []Hook           `yaml:"hooks"`
	KubeconfigPaths               []KubeconfigPath `yaml:"kubeconfigPaths"`
}

type KubeconfigPath struct {
	Path  string    `yaml:"path"`
	Store StoreKind `yaml:"store"`
}

type Hook struct {
	Name      string   `yaml:"name"`
	Type      HookType `yaml:"type"`
	Path      *string  `yaml:"path"`
	Arguments []string `yaml:"arguments"`
	Execution *struct {
		Interval *time.Duration `yaml:"interval"`
	} `yaml:"execution"`
}

type HookState struct {
	HookName          string    `yaml:"hookName"`
	LastExecutionTime time.Time `yaml:"lastExecutionTime"`
}
