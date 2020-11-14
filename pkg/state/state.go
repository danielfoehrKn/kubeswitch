package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/danielfoehrkn/kubectlSwitch/types"
)

// GetHookState loads and unmarshalls a hook state file
func GetHookState(log *logrus.Entry, hookStateFilepath string) (*types.HookState, error) {
	if _, err := os.Stat(hookStateFilepath); err != nil {
		if os.IsNotExist(err) {
			// occurs during first execution of the hook
			log.Infof("Configuration file not found under path: %q", hookStateFilepath)
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

func UpdateHookState(hookName, stateFileName string) error {
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

