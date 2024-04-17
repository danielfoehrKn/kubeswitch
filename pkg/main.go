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

package pkg

import (
	"fmt"
	"strings"
	"sync"
	"time"

	historyutil "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history/util"
	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/danielfoehrkn/kubeswitch/pkg/index"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	aliasutil "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/util"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
	"github.com/danielfoehrkn/kubeswitch/types"
)

var (
	// need mutex for all maps because multiple stores with multiple go routines write to the map simultaneously
	// in addition the fuzzy search reads from the maps during hot reload
	allKubeconfigContextNamesLock = sync.RWMutex{}
	allKubeconfigContextNames     []string

	contextToPathMapping     = make(map[string]string)
	contextToPathMappingLock = sync.RWMutex{}

	pathToTagsMapping     = make(map[string]map[string]string)
	pathToTagsMappingLock = sync.RWMutex{}

	pathToKubeconfig     = make(map[string]string)
	pathToKubeconfigLock = sync.RWMutex{}

	pathToStoreID   = make(map[string]string)
	pathToStoreLock = sync.RWMutex{}

	aliasToContext     = make(map[string]string)
	aliasToContextLock = sync.RWMutex{}

	hotReloadLock sync.RWMutex

	// aggregated errors that were suppressed during the search
	// are logged on exit
	searchError error

	logger = logrus.New()
)

func Switcher(stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex, showPreview bool) (*string, *string, error) {
	c, err := DoSearch(stores, config, stateDir, noIndex)
	if err != nil {
		return nil, nil, err
	}

	go func(channel chan DiscoveredContext) {
		for discoveredContext := range channel {
			if discoveredContext.Error != nil {
				// aggregate the errors during the search to show after the selection screen
				logger.Debugf("%v", discoveredContext.Error)
				searchError = multierror.Append(searchError, discoveredContext.Error)
				continue
			}

			if discoveredContext.Store == nil {
				// this should not happen
				logger.Debugf("store returned from search is nil. This should not happen")
				continue
			}
			kubeconfigStore := *discoveredContext.Store

			contextName := discoveredContext.Name
			if len(discoveredContext.Alias) > 0 {
				contextName = discoveredContext.Alias
				writeToAliasToContext(discoveredContext.Alias, discoveredContext.Name)
			}

			// write to global map that is polled by the fuzzy search
			appendToAllKubeconfigContextNames(contextName)
			// add to global contextToPath map
			// required to map back from selected context -> path
			writeToContextToPathMapping(contextName, discoveredContext.Path)
			// required to map back from kubeconfig path -> tags
			writeToPathToTagsMapping(discoveredContext.Path, discoveredContext.Tags)
			// associate (path -> store)
			// required to map back from selected context -> path -> store -> store.getKubeconfig(path)
			writeToPathToStoreID(discoveredContext.Path, kubeconfigStore.GetID())
		}
	}(*c)

	// remember the store for later kubeconfig retrieval
	var kindToStore = map[string]store.KubeconfigStore{}
	for _, s := range stores {
		kindToStore[s.GetID()] = s
	}

	defer logSearchErrors()

	kubeconfigPath, selectedContext, err := showFuzzySearch(kindToStore, showPreview)
	if err != nil {
		return nil, nil, err
	}

	if len(kubeconfigPath) == 0 {
		return nil, nil, nil
	}

	// map back kubeconfig path to the store kind
	storeID := readFromPathToStoreID(kubeconfigPath)

	// get the store for the store ID
	store := kindToStore[storeID]

	// get the tags associated with the selected kubeconfig path
	tags := readFromPathToTagsMapping(kubeconfigPath)

	// use the store to get the kubeconfig for the selected kubeconfig path
	kubeconfigData, err := store.GetKubeconfigForPath(kubeconfigPath, tags)
	if err != nil {
		return nil, nil, err
	}

	kubeconfig, err := kubeconfigutil.NewKubeconfig(kubeconfigData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse selected kubeconfig. Please check if this file is a valid kubeconfig: %v", err)
	}

	// save the original selected context for the history
	contextForHistory := selectedContext

	if len(store.GetContextPrefix(kubeconfigPath)) > 0 && strings.HasPrefix(selectedContext, store.GetContextPrefix(kubeconfigPath)) {
		// we need to remove an existing prefix from the selected context
		// because otherwise the kubeconfig contains an invalid current-context
		selectedContext = strings.TrimPrefix(selectedContext, fmt.Sprintf("%s/", store.GetContextPrefix(kubeconfigPath)))
	}

	if err := kubeconfig.SetContext(selectedContext, aliasutil.GetContextForAlias(selectedContext, aliasToContext), store.GetContextPrefix(kubeconfigPath)); err != nil {
		return nil, nil, err
	}

	if err := kubeconfig.SetKubeswitchContext(contextForHistory); err != nil {
		return nil, nil, err
	}

	// write a temporary kubeconfig file and return the path
	tempKubeconfigPath, err := kubeconfig.WriteKubeconfigFile()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write temporary kubeconfig file: %v", err)
	}

	// get namespace for current context
	ns, err := kubeconfig.NamespaceOfContext(kubeconfig.GetCurrentContext())
	if err != nil {
		logger.Warnf("failed to append context to history file: failed to get namespace of current context: %v", err)
	} else if err := historyutil.AppendToHistory(contextForHistory, ns); err != nil {
		logger.Warnf("failed to append context to history file: %v", err)
	}

	return &tempKubeconfigPath, &selectedContext, nil
}

// writeIndex tries to write the Index file for the kubeconfig store
// if it fails to do so, it logs a warning, but does not panic
func writeIndex(store store.KubeconfigStore, searchIndex *index.SearchIndex, ctxToPathMapping map[string]string, ctxToTagsMapping map[string]map[string]string) {
	index := types.Index{
		Kind:                 store.GetKind(),
		ContextToPathMapping: ctxToPathMapping,
		ContextToTags:        ctxToTagsMapping,
	}

	if err := searchIndex.Write(index); err != nil {
		store.GetLogger().Warnf("failed to write kubeconfig store index file: %v", err)
		return
	}

	indexStateToWrite := types.IndexState{
		Kind:           store.GetKind(),
		LastUpdateTime: time.Now().UTC(),
	}

	if err := searchIndex.WriteState(indexStateToWrite); err != nil {
		store.GetLogger().Warnf("failed to write index state file: %v", err)
	}
}

func showFuzzySearch(storeIDToStore map[string]store.KubeconfigStore, showPreview bool) (string, string, error) {
	// display selection dialog for all kubeconfig context names
	idx, err := fuzzyfinder.Find(
		&allKubeconfigContextNames,
		func(i int) string {
			return readFromAllKubeconfigContextNames(i)
		},
		getFuzzyFinderOptions(storeIDToStore, showPreview)...,
	)

	if err != nil {
		return "", "", err
	}

	// map selection back to kubeconfig
	selectedContext := readFromAllKubeconfigContextNames(idx)
	kubeconfigPath := readFromContextToPathMapping(selectedContext)

	return kubeconfigPath, selectedContext, nil
}

// getFuzzyFinderOptions returns a list of fuzzy finder options
func getFuzzyFinderOptions(storeIDToStore map[string]store.KubeconfigStore, showPreview bool) []fuzzyfinder.Option {
	options := []fuzzyfinder.Option{fuzzyfinder.WithHotReloadLock(hotReloadLock.RLocker())}

	if showPreview {
		log := logrus.New()
		withPreviewWindow := fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if !showPreview || i == -1 {
				return ""
			}

			// read the content of the kubeconfig here and display
			hotReloadLock.RLock()
			currentContextName := readFromAllKubeconfigContextNames(i)
			hotReloadLock.RUnlock()

			path := readFromContextToPathMapping(currentContextName)
			tags := readFromPathToTagsMapping(path)
			storeID := readFromPathToStoreID(path)
			kubeconfigStore := storeIDToStore[storeID]

			var storeSpecificPreview *string
			previewer, ok := kubeconfigStore.(store.Previewer)
			if ok {
				pr, err := previewer.GetSearchPreview(path, tags)
				if err != nil {
					log.Debugf("failed to get preview for store %s: %v", kubeconfigStore.GetID(), err)
					return ""
				}
				storeSpecificPreview = &pr
			}

			preview, err := getSanitizedKubeconfigForKubeconfigPath(kubeconfigStore, path, tags)
			if err != nil {
				log.Debugf("failed to get kubeconfig preview: %v", err)
				return ""
			}

			if storeSpecificPreview != nil {
				separators := make([]string, 20)
				for i := 0; i < 20; i++ {
					separators[i] = "-"
				}
				preview = fmt.Sprintf("%s \n %s \n \n %s", preview, strings.Join(separators, "-"), *storeSpecificPreview)
			}

			return preview
		})

		options = append(options, withPreviewWindow)
	}

	return options
}

func getSanitizedKubeconfigForKubeconfigPath(kubeconfigStore store.KubeconfigStore, path string, tags map[string]string) (string, error) {
	// during first run without index, the files are already read in the getContextsForKubeconfigPath and saved in-memory
	kubeconfig := readFromPathToKubeconfig(path)
	if len(kubeconfig) > 0 {
		return kubeconfig, nil
	}

	data, err := kubeconfigStore.GetKubeconfigForPath(path, tags)
	if err != nil {
		return "", fmt.Errorf("could not read kubeconfig with path '%s': %v", path, err)
	}

	config, err := util.ParseSanitizedKubeconfig(data)
	if err != nil {
		return "", fmt.Errorf("could not parse kubeconfig with path '%s': %v", path, err)
	}

	kubeconfigData, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("could not marshal kubeconfig with path '%s': %v", path, err)
	}

	// save kubeconfig content to in-memory map to avoid duplicate read operation in getSanitizedKubeconfigForKubeconfigPath
	writeToPathToKubeconfig(path, string(kubeconfigData))

	return string(kubeconfigData), nil
}

func readFromAllKubeconfigContextNames(index int) string {
	allKubeconfigContextNamesLock.RLock()
	defer allKubeconfigContextNamesLock.RUnlock()
	return allKubeconfigContextNames[index]
}

func appendToAllKubeconfigContextNames(values ...string) {
	allKubeconfigContextNamesLock.Lock()
	defer allKubeconfigContextNamesLock.Unlock()
	allKubeconfigContextNames = append(allKubeconfigContextNames, values...)
}

func readFromContextToPathMapping(key string) string {
	contextToPathMappingLock.RLock()
	defer contextToPathMappingLock.RUnlock()
	return contextToPathMapping[key]
}

func writeToContextToPathMapping(key, value string) {
	contextToPathMappingLock.Lock()
	defer contextToPathMappingLock.Unlock()
	contextToPathMapping[key] = value
}

func readFromPathToTagsMapping(key string) map[string]string {
	pathToTagsMappingLock.RLock()
	defer pathToTagsMappingLock.RUnlock()
	return pathToTagsMapping[key]
}

func writeToPathToTagsMapping(key string, value map[string]string) {
	pathToTagsMappingLock.Lock()
	defer pathToTagsMappingLock.Unlock()
	pathToTagsMapping[key] = value
}

func readFromPathToStoreID(key string) string {
	pathToStoreLock.RLock()
	defer pathToStoreLock.RUnlock()
	return pathToStoreID[key]
}

func writeToPathToStoreID(key string, value string) {
	pathToStoreLock.Lock()
	defer pathToStoreLock.Unlock()
	pathToStoreID[key] = value
}

func readFromPathToKubeconfig(key string) string {
	pathToKubeconfigLock.RLock()
	defer pathToKubeconfigLock.RUnlock()
	return pathToKubeconfig[key]
}

func writeToPathToKubeconfig(key, value string) {
	pathToKubeconfigLock.Lock()
	defer pathToKubeconfigLock.Unlock()
	pathToKubeconfig[key] = value
}

func writeToAliasToContext(key, value string) {
	aliasToContextLock.Lock()
	defer aliasToContextLock.Unlock()
	aliasToContext[key] = value
}

// logSearchErrors logs errors that were suppressed during the search
func logSearchErrors() {
	if searchError != nil {
		logger.Warnf("Supressed warnings during the search: %v", searchError.Error())
	}
}
