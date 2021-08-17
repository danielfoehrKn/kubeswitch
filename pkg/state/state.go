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

package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// GetHookState loads and unmarshalls a hook state file
func GetHookState(log *logrus.Entry, hookStateFilepath string) (*types.HookState, error) {
	if _, err := os.Stat(hookStateFilepath); err != nil {
		if os.IsNotExist(err) {
			// occurs during first execution of the hook
			log.Debugf("State file not found under path: %q", hookStateFilepath)
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
