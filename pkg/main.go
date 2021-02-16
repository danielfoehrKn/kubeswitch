package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/danielfoehrkn/kubectlSwitch/pkg/index"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/util"
	"github.com/danielfoehrkn/kubectlSwitch/types"
)

var (
	// need mutex for all maps because multiple stores with multiple go routines write to the map simultaneously
	// in addition the fuzzy search reads from the maps during hot reload
	allKubeconfigContextNamesLock = sync.RWMutex{}
	allKubeconfigContextNames     []string

	contextToPathMapping     = make(map[string]string)
	contextToPathMappingLock = sync.RWMutex{}

	pathToKubeconfig     = make(map[string]string)
	pathToKubeconfigLock = sync.RWMutex{}

	pathToStore     = make(map[string]types.StoreKind)
	pathToStoreLock = sync.RWMutex{}

	logger = logrus.New()
)

func Switcher(stores []store.KubeconfigStore, switchConfig *types.Config, stateDir string, showPreview bool) error {
	c, err := DoSearch(stores, switchConfig, stateDir)
	if err != nil {
		return err
	}

	go func(channel chan DiscoveredKubeconfig) {
		for discoveredKubeconfig := range channel {
			if discoveredKubeconfig.Error != nil {
				logger.Warnf("error returned from search: %v", discoveredKubeconfig.Error)
				continue
			}

			if discoveredKubeconfig.Store == nil {
				// this should not happen
				logger.Debugf("store returned from search is nil. This should not happen")
				continue
			}
			kubeconfigStore := *discoveredKubeconfig.Store

			// write to global map that is polled by the fuzzy search
			appendToAllKubeconfigContextNames(discoveredKubeconfig.ContextNames...)
			// associate (path -> store)
			// required to map back from selected context -> path -> store -> store.getKubeconfig(path)
			writeToPathToStoreKind(discoveredKubeconfig.Path, kubeconfigStore.GetKind())

			for _, context := range discoveredKubeconfig.ContextNames {
				// add to global contextToPath map
				// required to map back from selected context -> path
				writeToContextToPathMapping(context, discoveredKubeconfig.Path)
			}
		}
	}(*c)


	// remember the store for later kubeconfig retrieval
	var kindToStore = map[types.StoreKind]store.KubeconfigStore{}
	for _, s := range stores {
		kindToStore[s.GetKind()] = s
	}

	kubeconfigPath, selectedContext, err := showFuzzySearch(kindToStore, showPreview)

	if len(kubeconfigPath) == 0 {
		return nil
	}

	storeKind := readFromPathToStoreKind(kubeconfigPath)
	store := kindToStore[storeKind]
	kubeconfigData, err := store.GetKubeconfigForPath(kubeconfigPath)
	if err != nil {
		return err
	}

	selectedContext = util.GetContextWithoutFolderPrefix(selectedContext)
	if err := util.AppendContextToHistory(selectedContext); err != nil {
		logger.Warnf("failed to append context to history file: %v", err)
	}

	tempKubeconfigPath, err := util.SetCurrentContextOnTemporaryKubeconfigFile(kubeconfigData, selectedContext)
	if err != nil {
		return fmt.Errorf("failed to write current context to temporary kubeconfig: %v", err)
	}

	// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG environment Variable
	fmt.Print(tempKubeconfigPath)

	return nil
}

// writeIndex tries to write the Index file for the kubeconfig store
// if it fails to do so, it logs a warning, but does not panic
func writeIndex(store store.KubeconfigStore, searchIndex *index.SearchIndex, ctxToPathMapping map[string]string) {
	index := types.Index{
		Kind:                 store.GetKind(),
		ContextToPathMapping: ctxToPathMapping,
	}

	if err := searchIndex.Write(index); err != nil {
		store.GetLogger().Warnf("failed to write index file to speed up future fuzzy searches: %v", err)
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

func showFuzzySearch(kindToStore map[types.StoreKind]store.KubeconfigStore, showPreview bool) (string, string, error) {
	log := logrus.New()
	// display selection dialog for all the kubeconfig context names
	idx, err := fuzzyfinder.Find(
		&allKubeconfigContextNames,
		func(i int) string {
			return readFromAllKubeconfigContextNames(i)
		},
		fuzzyfinder.WithHotReload(),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if !showPreview || i == -1 {
				return ""
			}

			// read the content of the kubeconfig here and display
			currentContextName := readFromAllKubeconfigContextNames(i)
			path := readFromContextToPathMapping(currentContextName)
			storeType := readFromPathToStoreKind(path)
			store := kindToStore[storeType]
			kubeconfig, err := getSanitizedKubeconfigForKubeconfigPath(store, path)
			if err != nil {
				log.Warnf("failed to show preview: %v", err)
				return ""
			}
			return kubeconfig
		}))

	if err != nil {
		return "", "", err
	}

	// map selection back to kubeconfig
	selectedContext := readFromAllKubeconfigContextNames(idx)
	kubeconfigPath := readFromContextToPathMapping(selectedContext)

	return kubeconfigPath, selectedContext, nil
}

func getContextsForKubeconfigPath(kubeconfigStore store.KubeconfigStore, kubeconfigPath string) ([]string, error) {
	bytes, err := kubeconfigStore.GetKubeconfigForPath(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// parse into struct that does not contain the credentials
	config, err := parseSanitizedKubeconfig(bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse Kubeconfig with path '%s': %v", kubeconfigPath, err)
	}

	kubeconfigData, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("could not marshal kubeconfig with path '%s': %v", kubeconfigPath, err)
	}

	// save kubeconfig content to in-memory map to avoid duplicate read operation in getSanitizedKubeconfigForKubeconfigPath
	writeToPathToKubeconfig(kubeconfigPath, string(kubeconfigData))
	return getContextsFromKubeconfig(kubeconfigStore.GetKind(), kubeconfigPath, config)
}

func getSanitizedKubeconfigForKubeconfigPath(kubeconfigStore store.KubeconfigStore, path string) (string, error) {
	// during first run without index, the files are already read in the getContextsForKubeconfigPath and saved in-memory
	kubeconfig := readFromPathToKubeconfig(path)
	if len(kubeconfig) > 0 {
		return kubeconfig, nil
	}

	data, err := kubeconfigStore.GetKubeconfigForPath(path)
	if err != nil {
		return "", fmt.Errorf("could not read file with path '%s': %v", path, err)
	}

	config, err := parseSanitizedKubeconfig(data)
	if err != nil {
		return "", fmt.Errorf("could not parse Kubeconfig with path '%s': %v", path, err)
	}

	kubeconfigData, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("could not marshal kubeconfig with path '%s': %v", path, err)
	}

	// save kubeconfig content to in-memory map to avoid duplicate read operation in getSanitizedKubeconfigForKubeconfigPath
	writeToPathToKubeconfig(path, string(kubeconfigData))

	return string(kubeconfigData), nil
}

func getContextsFromKubeconfig(kind types.StoreKind, path string, kubeconfig *types.KubeConfig) ([]string, error) {
	parentFoldername := filepath.Base(filepath.Dir(path))
	if kind == types.StoreKindVault {
		// for vault, the secret name itself contains the semantic information (not the key of the kv-pair of the vault secret)
		parentFoldername = filepath.Base(path)
	}
	return getContextNames(kubeconfig, parentFoldername), nil
}

func parseSanitizedKubeconfig(data []byte) (*types.KubeConfig, error) {
	config := types.KubeConfig{}

	// unmarshal in a form that does not include the credentials
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal kubeconfig: %v", err)
	}
	return &config, nil
}

// sets the parent folder name to each context in the kubeconfig file
func getContextNames(config *types.KubeConfig, parentFoldername string) []string {
	var contextNames []string
	for _, context := range config.Contexts {
		split := strings.Split(context.Name, "/")
		if len(split) > 1 {
			// already has the directory name in there. override it in case it changed
			contextNames = append(contextNames, fmt.Sprintf("%s/%s", parentFoldername, split[len(split)-1]))
		} else {
			contextNames = append(contextNames, fmt.Sprintf("%s/%s", parentFoldername, context.Name))
		}
	}
	return contextNames
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

func readFromPathToStoreKind(key string) types.StoreKind {
	pathToStoreLock.RLock()
	defer pathToStoreLock.RUnlock()
	return pathToStore[key]
}

func writeToPathToStoreKind(key string, value types.StoreKind) {
	pathToStoreLock.Lock()
	defer pathToStoreLock.Unlock()
	pathToStore[key] = value
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
