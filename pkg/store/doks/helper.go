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

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// defaultConfigName is the name of the default `doctl` config file
	defaultConfigName = "config.yaml"
)

// GetDoctlConfiguration parses the `doctl` config.yaml file to return a map(context_name -> access_token}
// A context_name in `doctl` represents a DO account.
func GetDoctlConfiguration() (*DoctlConfig, error) {
	configFilePath, err := getDefaultDoctlConfigFilepath()
	if err != nil {
		return nil, err
	}

	return loadDoctlConfigFromFile(*configFilePath)
}

// loadConfigFromFile takes a filename and de-serializes the contents into a Configuration object.
func loadDoctlConfigFromFile(configPath string) (*DoctlConfig, error) {
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &DoctlConfig{}
	if len(bytes) == 0 {
		return config, nil
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal doctl config at '%s': %v", configPath, err)
	}
	return config, nil
}

// getDefaultDoctlConfigFilepath returns the config file created by the official DigitalOcean command line client containing `doctl` configuration
// which most importantly includes the contexts and associated access tokens
func getDefaultDoctlConfigFilepath() (*string, error) {
	configHome, err := defaultConfigHome()
	if err != nil {
		return nil, err
	}

	fp := filepath.Join(*configHome, defaultConfigName)
	return &fp, nil
}

// defaultConfigHome is the per-OS default location to store application specific configuration data
// OSX: /Users/danielfoehr/Library/Application\Support
func defaultConfigHome() (*string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(cfgDir, "doctl")
	return &path, nil
}
