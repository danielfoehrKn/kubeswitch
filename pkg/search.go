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

package pkg

import (
	"sync"

	"github.com/danielfoehrkn/kubeswitch/pkg/index"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	aliasstate "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/state"
	aliasutil "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/util"
	"github.com/danielfoehrkn/kubeswitch/types"
)

type DiscoveredContext struct {
	// Path is the kubeconfig path in the backing store (filesystem / Vault)
	Path string
	// Name ist the context name in the kubeconfig
	Name string
	// Alias is a custom alias defined for this context name
	Alias string
	// Store is a reference to the backing store that contains the kubeconfig
	Store *store.KubeconfigStore
	// Error is an error that occured during the search
	Error error
}

// DoSearch executes a concurrent search over the given kubeconfig stores
// returns results from all stores on the return channel
func DoSearch(stores []store.KubeconfigStore, config *types.Config, stateDir string) (*chan DiscoveredContext, error) {
	// first get defined alias in order to check if found kubecontext names should be display and returned
	// with a different name
	alias, err := aliasstate.GetDefaultAlias(stateDir)
	if err != nil {
		return nil, err
	}

	contextToAliasMapping := make(map[string]string)
	if alias != nil && alias.Content.ContextToAliasMapping != nil {
		contextToAliasMapping = alias.Content.ContextToAliasMapping
	}

	resultChannel := make(chan DiscoveredContext)
	wgResultChannel := sync.WaitGroup{}
	wgResultChannel.Add(len(stores))

	for _, kubeconfigStore := range stores {
		logger := kubeconfigStore.GetLogger()

		if err := kubeconfigStore.VerifyKubeconfigPaths(); err != nil {
			return nil, err
		}

		searchIndex, err := index.New(logger, kubeconfigStore.GetKind(), stateDir)
		if err != nil {
			return nil, err
		}

		shouldReadFromIndex, err := shouldReadFromIndex(searchIndex, kubeconfigStore, config)
		if err != nil {
			return nil, err
		}

		if shouldReadFromIndex {
			go func(store store.KubeconfigStore, index index.SearchIndex) {
				// reading from this store is finished, decrease wait counter
				defer wgResultChannel.Done()

				// directly set from pre-computed index
				content := index.GetContent()
				for contextName, path := range content {
					resultChannel <- DiscoveredContext{
						Path:  path,
						Name:  contextName,
						Alias: aliasutil.GetContextForAlias(contextName, contextToAliasMapping),
						Store: &store,
						Error: nil,
					}
				}
			}(kubeconfigStore, *searchIndex)

			continue
		}

		// otherwise, we need to query the backing store for the kubeconfig files
		c := make(chan store.SearchResult)
		go func(store store.KubeconfigStore, channel chan store.SearchResult) {
			// only close when directory search is over, otherwise send on closed resultChannel
			defer close(channel)
			store.GetLogger().Debugf("Starting search for store: %s", store.GetKind())
			store.StartSearch(channel)
		}(kubeconfigStore, c)

		go func(store store.KubeconfigStore, storeSearchChannel chan store.SearchResult, index index.SearchIndex) {
			// remember the context to kubeconfig path mapping for this this store
			// to write it to the index. Do not use the global "ContextToPathMapping"
			// as this contains contexts names from all stores combined
			localContextToPathMapping := make(map[string]string)
			for channelResult := range storeSearchChannel {
				if channelResult.Error != nil {
					resultChannel <- DiscoveredContext{
						Error: channelResult.Error,
					}
					continue
				}

				// get the context names from the parsed kubeconfig
				contexts, err := getContextsForKubeconfigPath(store, channelResult.KubeconfigPath)
				if err != nil {
					// do not throw Error, try to parse the other files
					// this will happen a lot when using vault as storage because the secrets key value needs to match the desired kubeconfig name
					// this however cannot be checked without retrieving the actual secret (path discovery is only list operation)
					continue
				}

				for _, contextName := range contexts {
					// write to result channel
					resultChannel <- DiscoveredContext{
						Path:  channelResult.KubeconfigPath,
						Name:  contextName,
						Alias: aliasutil.GetContextForAlias(contextName, contextToAliasMapping),
						Store: &store,
						Error: nil,
					}
					// add to local contextToPath map to write the index for this store only
					localContextToPathMapping[contextName] = channelResult.KubeconfigPath
				}
			}

			// reading from this store is finished, decrease wait counter
			wgResultChannel.Done()

			// write store index file now that the path discovery is complete
			writeIndex(store, &index, localContextToPathMapping)
		}(kubeconfigStore, c, *searchIndex)
	}

	go func() {
		defer close(resultChannel)
		wgResultChannel.Wait()
	}()

	return &resultChannel, nil
}

func shouldReadFromIndex(searchIndex *index.SearchIndex, kubeconfigStore store.KubeconfigStore, config *types.Config) (bool, error) {
	if searchIndex.HasContent() && searchIndex.HasKind(kubeconfigStore.GetKind()) {
		// found an index for the correct Store kind
		// check if should use existing index or not
		shouldReadFromIndex, err := searchIndex.ShouldBeUsed(config)
		if err != nil {
			return false, err
		}
		return shouldReadFromIndex, nil
	}
	return false, nil
}
