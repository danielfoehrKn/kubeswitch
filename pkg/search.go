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
	"os"
	"sync"

	"github.com/danielfoehrkn/kubeswitch/pkg/index"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	aliasstate "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/state"
	aliasutil "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/util"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/sirupsen/logrus"
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
func DoSearch(stores []store.KubeconfigStore, config *types.Config, stateDir string, noIndex bool) (*chan DiscoveredContext, error) {
	// Silence STDOUT during search to not interfere with the search selection screen
	// restore after search is over
	originalSTDOUT := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() {
		os.Stdout = originalSTDOUT
	}()

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
			// Required defines if errors when initializing this store should be logged
			if kubeconfigStore.GetStoreConfig().Required != nil && !*kubeconfigStore.GetStoreConfig().Required {
				continue
			}

			return nil, err
		}

		searchIndex, err := index.New(logger, kubeconfigStore.GetKind(), stateDir, kubeconfigStore.GetID())
		if err != nil {
			return nil, err
		}

		// do not use index if explicitly disabled via command line flag --no-index
		var readFromIndex bool
		if noIndex {
			readFromIndex = false
		} else {
			readFromIndex, err = shouldReadFromIndex(searchIndex, kubeconfigStore, config)
			if err != nil {
				return nil, err
			}
		}

		if readFromIndex {
			logrus.Debugf("Reading from index for store %s with kind %s", kubeconfigStore.GetID(), kubeconfigStore.GetKind())

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
			// remember the context to kubeconfig path mapping for this store
			// to write it to the index. Do not use the global "ContextToPathMapping"
			// as this contains contexts names from all stores combined
			localContextToPathMapping := make(map[string]string)
			for channelResult := range storeSearchChannel {
				if channelResult.Error != nil {
					// Required defines if errors when initializing this store should be logged
					if store.GetStoreConfig().Required != nil && !*store.GetStoreConfig().Required {
						continue
					}

					resultChannel <- DiscoveredContext{
						Error: fmt.Errorf("store %q returned an error during the search: %v", store.GetID(), channelResult.Error),
					}
					continue
				}

				bytes, err := store.GetKubeconfigForPath(channelResult.KubeconfigPath)
				if err != nil {
					// do not throw Error, try to parse the other files
					// this will happen a lot when using vault as storage because the secrets key value needs to match the desired kubeconfig name
					// this however cannot be checked without retrieving the actual secret (path discovery is only list operation)
					continue
				}

				// get the context names from the parsed kubeconfig
				kubeconfigString, contexts, err := util.GetContextsNamesFromKubeconfig(bytes, store.GetContextPrefix(channelResult.KubeconfigPath))
				if err != nil {
					store.GetLogger().Debugf("failed to get kubeconfig context names for kubeconfig with path %q: %v", channelResult.KubeconfigPath, err)
					resultChannel <- DiscoveredContext{
						Error: fmt.Errorf("failed to get kubeconfig context names for kubeconfig with path %q: %v", channelResult.KubeconfigPath, err),
					}
					// do not throw Error, try to parse the other files
					continue
				}

				// save kubeconfig content to in-memory map to avoid duplicate read operation in getSanitizedKubeconfigForKubeconfigPath
				writeToPathToKubeconfig(channelResult.KubeconfigPath, *kubeconfigString)

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
			if len(localContextToPathMapping) > 0 {
				writeIndex(store, &index, localContextToPathMapping)
			}
		}(kubeconfigStore, c, *searchIndex)
	}

	go func() {
		defer close(resultChannel)
		wgResultChannel.Wait()
	}()

	return &resultChannel, nil
}

func shouldReadFromIndex(searchIndex *index.SearchIndex, kubeconfigStore store.KubeconfigStore, config *types.Config) (bool, error) {
	// never write an index for the store from env variables and --kubeconfig-path command line falg
	if kubeconfigStore.GetID() == fmt.Sprintf("%s.%s", types.StoreKindFilesystem, "env-and-flag") {
		return false, nil
	}

	if searchIndex.HasContent() && searchIndex.HasKind(kubeconfigStore.GetKind()) {
		// found an index for the correct Store kind
		// check if should use existing index or not
		shouldReadFromIndex, err := searchIndex.ShouldBeUsed(config, kubeconfigStore.GetStoreConfig().RefreshIndexAfter)
		if err != nil {
			return false, err
		}
		return shouldReadFromIndex, nil
	}
	return false, nil
}
