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

package alias

import (
	"fmt"
	"os"

	"github.com/danielfoehrkn/k8ctx/pkg"
	"github.com/danielfoehrkn/k8ctx/pkg/store"
	"github.com/danielfoehrkn/k8ctx/pkg/subcommands/alias/state"
	kubeconfigutil "github.com/danielfoehrkn/k8ctx/pkg/util/kubectx_copied"
	"github.com/danielfoehrkn/k8ctx/types"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func ListAliases(stateDir string) error {
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		if err := os.Mkdir(stateDir, 0755); err != nil {
			return err
		}
	}

	a, err := state.GetDefaultAlias(stateDir)
	if err != nil {
		return err
	}

	if a.Content.ContextToAliasMapping == nil {
		fmt.Println("No aliases registered")
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Alias", "Context"})

	for ctx, alias := range a.Content.ContextToAliasMapping {
		t.AppendRows([]table.Row{
			{alias, ctx},
		})
	}
	t.AppendSeparator()
	t.AppendFooter(table.Row{"Total", len(a.Content.ContextToAliasMapping)})
	t.Render()

	return nil
}

func RemoveAlias(aliasToRemove, stateDir string) error {
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		if err := os.Mkdir(stateDir, 0755); err != nil {
			return err
		}
	}

	a, err := state.GetDefaultAlias(stateDir)
	if err != nil {
		return err
	}

	if a.Content.ContextToAliasMapping == nil {
		fmt.Println("No aliases registered")
		return nil
	}

	newAliases := map[string]string{}
	aliasFound := false
	for ctx, alias := range a.Content.ContextToAliasMapping {
		if alias == aliasToRemove {
			aliasFound = true
			continue
		}
		newAliases[ctx] = alias
	}

	if !aliasFound {
		return fmt.Errorf("alias with name %q does not exist", aliasToRemove)
	}

	a.Content.ContextToAliasMapping = newAliases
	if err := a.WriteAllAliases(); err != nil {
		return fmt.Errorf("failed to write aliases: %v", err)
	}
	fmt.Printf("Removed alias %q. There are now %d alias(es) defined. \n", aliasToRemove, len(newAliases))

	return nil
}

// Alias just maintains an alias record in the k8ctx
// state folder instead of renaming a context in the kubeconfig
// this works independent of the backing store
func Alias(aliasName, ctxNameToBeAliased string, stores []store.KubeconfigStore, config *types.Config, stateDir string) error {
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		if err := os.Mkdir(stateDir, 0755); err != nil {
			return err
		}
	}

	a, err := state.GetDefaultAlias(stateDir)
	if err != nil {
		return err
	}

	c, err := pkg.DoSearch(stores, config, stateDir)
	if err != nil {
		return err
	}

	for discoveredContext := range *c {
		if discoveredContext.Error != nil {
			logger.Warnf("cannot list contexts. Error returned from search: %v", discoveredContext.Error)
			continue
		}

		contextWithoutFolderPrefix := kubeconfigutil.GetContextWithoutFolderPrefix(discoveredContext.Name)
		if discoveredContext.Name == ctxNameToBeAliased || contextWithoutFolderPrefix == ctxNameToBeAliased {
			// write the context with the folder name
			if err := a.WriteAlias(aliasName, discoveredContext.Name); err != nil {
				return err
			}

			if _, err = fmt.Printf("Set alias %q for context %q", aliasName, discoveredContext.Name); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("cannot set alias %q: context %q not found", aliasName, ctxNameToBeAliased)
}
