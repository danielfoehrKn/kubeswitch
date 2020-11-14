package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/danielfoehrkn/kubectlSwitch/types"
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
