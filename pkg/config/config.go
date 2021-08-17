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

package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/danielfoehrkn/kubeswitch/pkg/config/migration"
	"github.com/danielfoehrkn/kubeswitch/types"
)

// LoadConfigFromFile takes a filename and de-serializes the contents into a Configuration object.
func LoadConfigFromFile(filepath string) (*types.Config, error) {
	// a config file is not required. Its ok if it does not exist.
	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	config := &types.Config{}
	if len(bytes) == 0 {
		return config, nil
	}

	err = yaml.Unmarshal(bytes, &config)
	// if version field is not set, it is an old config
	// check that kubeconfig stores are not set to avoid migrating a new config
	if err != nil || (len(config.Version) == 0 && len(config.KubeconfigStores) == 0) {
		// try with old config
		oldConfig := &types.ConfigOld{}
		err = yaml.Unmarshal(bytes, &oldConfig)
		if err == nil && oldConfig != nil {
			return MigrateConfig(*oldConfig, filepath)
		}
		return nil, fmt.Errorf("could not unmarshal config with path '%s': %v", filepath, err)
	}
	return config, nil
}

func MigrateConfig(old types.ConfigOld, filename string) (*types.Config, error) {
	// first, copy the old configuration
	file, err := os.Create(fmt.Sprintf("%s.old", filename))
	if err != nil {
		return nil, fmt.Errorf("failed to migrate SwitchConfig file: %w", err)
	}
	defer file.Close()

	output, err := yaml.Marshal(old)
	if err != nil {
		return nil, err
	}

	_, err = file.Write(output)
	if err != nil {
		return nil, err
	}

	// then overwrite the configuration with the new format
	new := migration.ConvertConfiguration(old)
	fileNew, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate SwitchConfig file: %w", err)
	}
	defer fileNew.Close()

	output, err = yaml.Marshal(new)
	if err != nil {
		return nil, err
	}

	_, err = fileNew.Write(output)
	if err != nil {
		return nil, err
	}

	return &new, nil
}
