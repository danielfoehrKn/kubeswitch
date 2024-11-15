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

package alias

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"

	"github.com/danielfoehrkn/kubeswitch/pkg"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/state"
	"github.com/danielfoehrkn/kubeswitch/types"
)

var logger = logrus.New()

func GetAliases(stateDir string) ([]string, error) {
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		if err := os.Mkdir(stateDir, 0755); err != nil {
			return nil, err
		}
	}

	a, err := state.GetDefaultAlias(stateDir)
	if err != nil {
		return nil, err
	}

	if a.Content.ContextToAliasMapping == nil {
		return nil, nil
	}

	var t []string
	for _, alias := range a.Content.ContextToAliasMapping {
		t = append(t, alias)
	}
	return t, nil
}

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

// Alias just maintains an alias record in the switch
// state folder instead of renaming a context in the kubeconfig
// this works independent of the backing store
func Alias(aliasName, ctxNameToBeAliased string, stores []storetypes.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) error {
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		if err := os.Mkdir(stateDir, 0755); err != nil {
			return err
		}
	}

	log := logrus.New().WithField("alias", aliasName)
	log.Debugf("Writing alias %s for context name %s", aliasName, ctxNameToBeAliased)

	aliasStore, err := state.GetDefaultAlias(stateDir)
	if err != nil {
		return err
	}

	c, err := pkg.DoSearch(stores, config, stateDir, noIndex)
	if err != nil {
		return err
	}

	for discoveredContext := range *c {
		if discoveredContext.Error != nil {
			logger.Warnf("cannot list contexts. Error returned from search: %v", discoveredContext.Error)
			continue
		}

		if discoveredContext.Store == nil {
			// this should not happen
			logger.Debugf("store returned from search is nil. This should not happen")
			continue
		}
		kubeconfigStore := *discoveredContext.Store

		var contextWithoutPrefix string
		if len(kubeconfigStore.GetContextPrefix(discoveredContext.Path)) > 0 && strings.HasPrefix(discoveredContext.Name, kubeconfigStore.GetContextPrefix(discoveredContext.Path)) {
			// we need to remove an existing prefix from the selected context
			// because otherwise the kubeconfig contains an invalid current-context
			contextWithoutPrefix = strings.TrimPrefix(discoveredContext.Name, fmt.Sprintf("%s/", kubeconfigStore.GetContextPrefix(discoveredContext.Path)))
		}

		if ctxNameToBeAliased == discoveredContext.Name || ctxNameToBeAliased == contextWithoutPrefix {
			// write the context like returned from the store (with or without prefix)
			replacedContextName, err := aliasStore.WriteAlias(aliasName, discoveredContext.Name)
			if err != nil {
				return err
			}

			var replacedContext string
			if replacedContextName != nil {
				replacedContext = fmt.Sprintf(" replacing existing alias for context with name %q", *replacedContextName)
			}

			if _, err = fmt.Printf("Set alias %q for context %q%s.\n", aliasName, discoveredContext.Name, replacedContext); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("cannot set aliasStore %q: context %q not found", aliasName, ctxNameToBeAliased)
}
