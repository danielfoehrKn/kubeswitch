package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/danielfoehrkn/kubectlSwitch/pkg/index"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/types"
)

const (
	// kubeconfigCurrentContext is a constant for the current context in a kubeconfig file
	kubeconfigCurrentContext = "current-context:"
	// TemporaryKubeconfigDir is a constant for the directory where the switcher stores the temporary kubeconfig files
	TemporaryKubeconfigDir = "$HOME/.kube/.switch_tmp"
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
)

func Switcher(stores []store.KubeconfigStore, switchConfig *types.Config, configPath string, stateDir string, showPreview bool) error {
	for _, kubeconfigStore := range stores {
		logger := kubeconfigStore.GetLogger()

		if err := kubeconfigStore.VeryKubeconfigPaths(); err != nil {
			return err
		}

		searchIndex, err := index.New(logger, kubeconfigStore.GetKind(), stateDir)
		if err != nil {
			return err
		}

		shouldReadFromIndex, err := shouldReadFromIndex(searchIndex, kubeconfigStore, switchConfig)
		if err != nil {
			return err
		}

		if shouldReadFromIndex {
			go func(store store.KubeconfigStore, index index.SearchIndex) {
				// directly set from pre-computed index
				content := index.GetContent()
				for contextName, path := range content {
					// add to allKubeconfigContextNames for fuzzy search
					appendToAllKubeconfigContextNames(contextName)
					writeToPathToStoreKind(path, store.GetKind())
					writeToContextToPathMapping(contextName, path)
				}
			}(kubeconfigStore, *searchIndex)

			continue
		}

		// otherwise, we need to query the backing store for the kubeconfig files
		c := make(chan store.SearchResult)
		go func(store store.KubeconfigStore, channel chan store.SearchResult) {
			// only close when directory search is over, otherwise send on closed channel
			defer close(channel)
			store.GetLogger().Debugf("Starting search for store: %s", store.GetKind())
			store.StartSearch(channel)
		}(kubeconfigStore, c)

		go func(store store.KubeconfigStore, channel chan store.SearchResult, index index.SearchIndex) error {
			// remember the context to kubeconfig path mapping for this this store
			// to write it to the index. Do not use the global "ContextToPathMapping"
			// as this contains contexts names from all stores combined
			localContextToPathMapping := make(map[string]string)
			for channelResult := range channel {
				if channelResult.Error != nil {
					logger.Warnf("error returned from search: %v", channelResult.Error)
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

				writeToPathToStoreKind(channelResult.KubeconfigPath, store.GetKind())

				// write to global map that is concurrently read by the fuzzy search
				appendToAllKubeconfigContextNames(contexts...)

				for _, context := range contexts {
					// add to global contextToPath map
					writeToContextToPathMapping(context, channelResult.KubeconfigPath)
					// add to local contextToPath map to write the index for this store only
					localContextToPathMapping[context] = channelResult.KubeconfigPath
				}
			}

			// write index file as soon as the path discovery is complete
			writeIndex(store, &index, localContextToPathMapping)
			return nil
		}(kubeconfigStore, c, *searchIndex)
	}

	var kindToStore = map[types.StoreKind]store.KubeconfigStore{}
	for _, s := range stores {
		kindToStore[s.GetKind()] = s
	}

	kubeconfigPath, selectedContext, err := showFuzzySearch(kindToStore, showPreview)

	if len(kubeconfigPath) == 0 {
		return nil
	}

	// remove the folder name from the context name
	split := strings.Split(selectedContext, "/")
	selectedContext = split[len(split)-1]

	storeKind := readFromPathToStoreKind(kubeconfigPath)
	store := kindToStore[storeKind]
	kubeconfigData, err := store.GetKubeconfigForPath(kubeconfigPath)
	if err != nil {
		return err
	}

	tempKubeconfigPath, err := setCurrentContextOnTemporaryKubeconfigFile(kubeconfigData, selectedContext)
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

func setCurrentContextOnTemporaryKubeconfigFile(kubeconfigData []byte, ctxnName string) (string, error) {
	currentContext := fmt.Sprintf("%s %s", kubeconfigCurrentContext, ctxnName)

	lines := strings.Split(string(kubeconfigData), "\n")

	foundCurrentContext := false
	for i, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.Contains(line, kubeconfigCurrentContext) {
			foundCurrentContext = true
			lines[i] = currentContext
		}
	}

	output := strings.Join(lines, "\n")
	tempDir := os.ExpandEnv(TemporaryKubeconfigDir)
	err := os.Mkdir(tempDir, 0700)
	if err != nil && !os.IsExist(err) {
		log.Fatalln(err)
	}

	// write temporary kubeconfig file
	file, err := ioutil.TempFile(tempDir, "config.*.tmp")
	if err != nil {
		log.Fatalln(err)
	}

	_, err = file.Write([]byte(output))
	if err != nil {
		log.Fatalln(err)
	}

	// if written file does not have the current context set,
	// add the context at the last line of the file
	if !foundCurrentContext {
		return file.Name(), appendCurrentContextToTemporaryKubeconfigFile(file.Name(), currentContext)
	}

	return file.Name(), nil
}

func appendCurrentContextToTemporaryKubeconfigFile(kubeconfigPath, currentContext string) error {
	file, err := os.OpenFile(kubeconfigPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(currentContext); err != nil {
		return err
	}
	return nil
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
