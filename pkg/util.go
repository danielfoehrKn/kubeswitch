package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

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

// LoadIndexFromFile takes a filename and de-serializes the contents into an Index object.
func LoadIndexFromFile(filename string) (*types.Index, error) {
	// an index file is not required. Its ok if it does not exist.
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file from %q: %v", filename, err)
	}

	index := &types.Index{}
	if len(bytes) == 0 {
		return index, nil
	}

	err = yaml.Unmarshal(bytes, &index)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal index file with path '%s': %v", filename, err)
	}
	return index, nil
}

// getHookState loads and unmarshalls a hook state file
func getHookState(hookStateFilepath string) (*types.HookState, error) {
	if _, err := os.Stat(hookStateFilepath); err != nil {
		if os.IsNotExist(err) {
			// occurs during first execution of the hook
			logger.Infof("Configuration file not found under path: %q", hookStateFilepath)
			return nil, nil
		}
		return nil, err
	}

	bytes, err := ioutil.ReadFile(hookStateFilepath)
	if err != nil {
		return nil, err
	}

	state := &types.HookState{}
	if len(bytes) == 0 {
		return state, nil
	}

	err = yaml.Unmarshal(bytes, &state)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal hook state file with path '%s': %v", hookStateFilepath, err)
	}

	return state, nil
}

func updateHookState(hookName, stateFileName string) error {
	// creates or truncate/clean the existing state file (only state is last execution anyways atm.)
	file, err := os.Create(stateFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	state := &types.HookState{
		HookName:          hookName,
		LastExecutionTime: time.Now().UTC(),
	}

	output, err := yaml.Marshal(state)
	if err != nil {
		return err
	}

	_, err = file.Write(output)
	if err != nil {
		return err
	}

	return nil
}

// getIndexState loads and unmarshalls an index state file
func getIndexState(stateFilepath string) (*types.IndexState, error) {
	if _, err := os.Stat(stateFilepath); err != nil {
		if os.IsNotExist(err) {
			// occurs during first execution of the hook
			logger.Warnf("Index state file not found under path: %q", stateFilepath)
			return nil, nil
		}
		return nil, err
	}

	bytes, err := ioutil.ReadFile(stateFilepath)
	if err != nil {
		return nil, err
	}

	state := &types.IndexState{}
	if len(bytes) == 0 {
		return state, nil
	}

	err = yaml.Unmarshal(bytes, &state)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal index state file with path '%s': %v", stateFilepath, err)
	}

	return state, nil
}

func writeIndexStoreState(indexToWrite types.IndexState, stateFileName string) error {
	// creates or truncate/clean the existing state file (only state is last execution anyways atm.)
	file, err := os.Create(stateFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	output, err := yaml.Marshal(indexToWrite)
	if err != nil {
		return err
	}

	_, err = file.Write(output)
	if err != nil {
		return err
	}

	return nil
}

func writeIndex(indexToUpdate types.Index, path string) error {
	// creates or truncate/clean the existing file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	output, err := yaml.Marshal(indexToUpdate)
	if err != nil {
		return err
	}

	_, err = file.Write(output)
	if err != nil {
		return err
	}

	return nil
}
