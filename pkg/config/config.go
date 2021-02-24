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

package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/danielfoehrkn/kubectlSwitch/types"
	"gopkg.in/yaml.v2"
)

// LoadConfigFromFile takes a filename and de-serializes the contents into a Configuration object.
func LoadConfigFromFile(filename string) (*types.Config, error) {
	// a config file is not required. Its ok if it does not exist.
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &types.Config{}
	if len(bytes) == 0 {
		return config, nil
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config with path '%s': %v", filename, err)
	}
	return config, nil
}
