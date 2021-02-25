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

package state

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/danielfoehrkn/k8ctx/types"
	"gopkg.in/yaml.v2"
)

const (
	// aliasFileName is the filename of the state file that contains all created aliases
	aliasFileName = "alias"
)

type Alias struct {
	aliasFilepath string
	Content       types.ContextAlias
}

// GetDefaultAlias get the default alias with the path to the state file set
func GetDefaultAlias(stateDir string) (*Alias, error) {
	a := Alias{
		aliasFilepath: fmt.Sprintf("%s/k8ctx.%s", stateDir, aliasFileName),
	}

	if err := a.loadFromFile(); err != nil {
		return nil, err
	}

	return &a, nil
}

// loadFromFile loads the existing alias record from the state file
func (a *Alias) loadFromFile() error {
	// an alias file is not required. Its ok if it does not exist.
	if _, err := os.Stat(a.aliasFilepath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	bytes, err := ioutil.ReadFile(a.aliasFilepath)
	if err != nil {
		return fmt.Errorf("failed to read alias file from %q. File corrupt?: %v", a.aliasFilepath, err)
	}

	existingAliases := types.ContextAlias{}
	if len(bytes) == 0 {
		return nil
	}

	err = yaml.Unmarshal(bytes, &existingAliases)
	if err != nil {
		return fmt.Errorf("could not unmarshal index file with path '%s': %v", a.aliasFilepath, err)
	}
	a.Content = existingAliases
	return nil
}

// WriteAlias overwrites the alias state file with new Content
func (a *Alias) WriteAlias(aliasName, contextName string) error {
	if a.Content.ContextToAliasMapping == nil {
		a.Content.ContextToAliasMapping = make(map[string]string, 1)
	}
	a.Content.ContextToAliasMapping[contextName] = aliasName
	return a.WriteAllAliases()
}

// WriteAlias overwrites the alias state file with new Content
func (a *Alias) WriteAllAliases() error {
	// overwrite the existing state file (only state is last execution anyways atm.)
	file, err := os.Create(a.aliasFilepath)
	if err != nil {
		return err
	}
	defer file.Close()

	output, err := yaml.Marshal(a.Content)
	if err != nil {
		return err
	}

	_, err = file.Write(output)
	if err != nil {
		return err
	}

	return nil
}
