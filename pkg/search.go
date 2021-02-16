package pkg

import (
	"sync"

	"github.com/danielfoehrkn/kubectlSwitch/pkg/index"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/types"
)

type DiscoveredKubeconfig struct {
	// Path is the kubeconfig path in the backing store (filesystem / Vault)
	Path string
	// ContextNames are the context names in the kubeconfig
	ContextNames []string
	// Store is a reference to the backing store that contains the kubeconfig
	Store *store.KubeconfigStore
	// Error is an error that occured during the search
	Error error
}

// DoSearch executes a concurrent search over the given kubeconfig stores
// returns results from all stores on the return channel
func DoSearch(stores []store.KubeconfigStore, switchConfig *types.Config, stateDir string) (*chan DiscoveredKubeconfig, error) {
	resultChannel := make(chan DiscoveredKubeconfig)
	wgResultChannel := sync.WaitGroup{}
	wgResultChannel.Add(len(stores))

	for _, kubeconfigStore := range stores {
		logger := kubeconfigStore.GetLogger()

		if err := kubeconfigStore.VeryKubeconfigPaths(); err != nil {
			return nil, err
		}

		searchIndex, err := index.New(logger, kubeconfigStore.GetKind(), stateDir)
		if err != nil {
			return nil, err
		}

		shouldReadFromIndex, err := shouldReadFromIndex(searchIndex, kubeconfigStore, switchConfig)
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
					resultChannel <- DiscoveredKubeconfig{
						Path:         path,
						ContextNames: []string{contextName},
						Store:        &store,
						Error:        nil,
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
					resultChannel <- DiscoveredKubeconfig{
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

				// write to result channel
				resultChannel <- DiscoveredKubeconfig{
					Path:         channelResult.KubeconfigPath,
					ContextNames: contexts,
					Store:        &store,
					Error:        nil,
				}

				for _, context := range contexts {
					// add to local contextToPath map to write the index for this store only
					localContextToPathMapping[context] = channelResult.KubeconfigPath
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

func shouldReadFromIndex(searchIndex *index.SearchIndex, kubeconfigStore store.KubeconfigStore, switchConfig *types.Config) (bool, error) {
	if searchIndex.HasContent() && searchIndex.HasKind(kubeconfigStore.GetKind()) {
		// found an index for the correct Store kind
		// check if should use existing index or not
		shouldReadFromIndex, err := searchIndex.ShouldBeUsed(switchConfig)
		if err != nil {
			return false, err
		}
		return shouldReadFromIndex, nil
	}
	return false, nil
}
