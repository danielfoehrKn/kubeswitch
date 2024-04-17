// Copyright 2024 The Kubeswitch authors
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

package doks

// DoctlConfig represents the `doctl` config file which is created when users interact with the `doctl` command line tool (such as when running `doctl auth init`)
type DoctlConfig struct {
	// DefaultContextName is the name of the top-level context in the `doctl` config file.
	// Typically: "default"
	DefaultContextName string `yaml:"context"`
	// DefaultAuthContextAccessToken is the access token associated with the DefaultContextName
	DefaultAuthContextAccessToken string `yaml:"access-token"`
	// AuthContexts represents a mapping {context_name -> token}
	AuthContexts     map[string]string `yaml:"auth-contexts"`
	Kubeconfig       Kubeconfig        `yaml:"kubeconfig"`
	HttpRetryMax     int               `yaml:"http-retry-max"`
	HttpRetryWaitMax int               `yaml:"http-retry-wait-max"`
	HttpRetryWaitMin int               `yaml:"http-retry-wait-min"`
	ApiUrl           string            `yaml:"api-url"`
}

// Kubeconfig is the kubeconfig sub-section of the `doctl` config file
type Kubeconfig struct {
	// SaveConfig contains configuration when saving the DOKS kubeconfig via `doctl`
	SaveConfig Save `yaml:"save"`
}

// Save contains configuration when saving the DOKS kubeconfig via `doctl`
type Save struct {
	// ExpirySeconds is the expiratio nseconds for tokens used as authentication of created kubeconfigs
	// The default value is 0, in which case DO's default will be taken (defaults to 7days)
	ExpirySeconds string `yaml:"expiry-seconds"`
}
