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

	"github.com/danielfoehrkn/kubectlSwitch/types"
	"github.com/ktr0731/go-fuzzyfinder"
	"gopkg.in/yaml.v2"
)

const (
	separator                = "/"
	kubeconfigCurrentContext = "current-context:"
	// temporaryKubeconfigDir is a constant for the directory where the switcher stores the temporary kubeconfig files
	temporaryKubeconfigDir = "$HOME/.kube/.switch_tmp"
	// indexFileName is the filename of the file containing a pre-computed context -> kubeconfig path mapping
	// located at the root of the given kubeconfigDirectory
	indexFileName = "index"
	// indexStateFileName is the filename of the index state file containing the last time a store index has been updated
	// located at the root of the given kubeconfigDirectory
	indexStateFileName = "index.state"
	// defaultKubeconfigRediscoveryInterval is the interval after which the index file is refreshed
	// rediscovering the kubeconfig paths and context names from the store may lead to delays/ more API requests during execution
	defaultKubeconfigRediscoveryInterval = 20 * time.Minute
)

type store interface {
	getKind() types.StoreKind
	checkPath(path string) error
	discoverPaths(searchPath string, kubeconfigName string, channel chan channelResult)
	getContextsForPath(path string) ([]string, error)
	getSanitizedKubeconfigForContext(path string) (string, error)
	getKubeconfigForPath(path string) ([]byte, error)
}

type FileStore struct{}

func Switcher(configPath string, kubeconfigDirectory string, kubeconfigFileName string, showPreview bool) error {
	kubeconfigStore := &FileStore{}

	if err := kubeconfigStore.checkPath(kubeconfigDirectory); err != nil {
		return err
	}

	indexFilepath := fmt.Sprintf("%s/switch.%s.%s", kubeconfigDirectory, kubeconfigStore.getKind(), indexFileName)
	index, err := LoadIndexFromFile(indexFilepath)
	if err != nil {
		return err
	}

	shouldReadFromIndex := false
	indexStateFilepath := fmt.Sprintf("%s/switch.%s.%s", kubeconfigDirectory, kubeconfigStore.getKind(), indexStateFileName)

	if index != nil && index.Kind == kubeconfigStore.getKind() {
		// found an index for the correct store kind
		// check if should use existing index or not
		shouldReadFromIndex, err = ShouldReadFromIndex(kubeconfigStore, configPath, indexStateFilepath)
		if err != nil {
			return err
		}
	}

	var kubeconfigPath, selectedContext string
	if shouldReadFromIndex {
		kubeconfigPath, selectedContext, err = fuzzySearchFromIndex(kubeconfigStore, *index, showPreview)
		if err != nil {
			return err
		}
	} else {
		kubeconfigPath, selectedContext, err = fuzzySearchWithRediscovery(kubeconfigStore, showPreview, kubeconfigDirectory, kubeconfigFileName, indexFilepath, indexStateFilepath)
		if err != nil {
			return err
		}
	}

	// remove the folder name from the context name
	split := strings.Split(selectedContext, separator)
	selectedContext = split[len(split)-1]

	kubeconfigData, err := kubeconfigStore.getKubeconfigForPath(kubeconfigPath)
	if err != nil {
		return err
	}

	tempKubeconfigPath, err := setCurrentContext(kubeconfigData, selectedContext)
	if err != nil {
		return fmt.Errorf("failed to write current context to temporary kubeconfig: %v", err)
	}

	// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG ENV Variable
	fmt.Print(tempKubeconfigPath)

	return nil
}

// ShouldReadFromIndex checks if the index file with pre-computed mappings should be used
func ShouldReadFromIndex(kubeconfigStore store, configPath string, indexStatePath string) (bool, error) {
	indexState, err := getIndexState(indexStatePath)
	if err != nil {
		return false, fmt.Errorf("failed to get index state: %v", err)
	}

	// do not read from existing index if there is no index state file
	// we do not know when the index last has been refreshed
	if indexState == nil || indexState.Kind != kubeconfigStore.getKind() {
		return false, nil
	}

	switchConfig, err := LoadConfigFromFile(configPath)
	if err != nil {
		return false, fmt.Errorf("failed to check if should read from index file: %v", err)
	}

	// the switch config has no KubeconfigRediscoveryInterval set - take default rediscovery interval
	kubeconfigRediscoveryInterval := switchConfig.KubeconfigRediscoveryInterval
	if switchConfig == nil || kubeconfigRediscoveryInterval == nil {
		return time.Now().UTC().Before(indexState.LastUpdateTime.UTC().Add(defaultKubeconfigRediscoveryInterval)), nil
	}

	return time.Now().UTC().Before(indexState.LastUpdateTime.UTC().Add(*kubeconfigRediscoveryInterval)), nil
}

type channelResult struct {
	kubeconfigPath string
	error          error
}

var (
	allKubeconfigContextNames []string

	contextToPathMapping     = make(map[string]string)
	contextToPathMappingLock = sync.RWMutex{}

	pathToKubeconfig     = make(map[string]string)
	pathToKubeconfigLock = sync.RWMutex{}
)

func fuzzySearchFromIndex(store store, index types.Index, showPreview bool) (string, string, error) {
	contextToPathMapping = index.ContextToPathMapping

	// build allKubeconfigContextNames from pre-computed index
	allKubeconfigContextNames = make([]string, len(contextToPathMapping))
	i := 0
	for k := range contextToPathMapping {
		allKubeconfigContextNames[i] = k
		i++
	}

	return showFuzzySearch(store, showPreview)
}

func fuzzySearchWithRediscovery(store store, showPreview bool, rootDir, kubeconfigFileName, indexFilePath string, indexStatePath string) (string, string, error) {
	var channel = make(chan channelResult)

	go func() {
		// only close when directory search is over, otherwise send on closed channel
		defer close(channel)
		store.discoverPaths(rootDir, kubeconfigFileName, channel)
	}()

	go func() error {
		for channelResult := range channel {
			if channelResult.error != nil {
				logger.Warnf("error returned from path discovery: %v", channelResult.error)
				continue
			}

			// get the context names from the parsed kubeconfig
			contexts, err := store.getContextsForPath(channelResult.kubeconfigPath)
			if err != nil {
				// do not throw error, try to parse the other files
				continue
			}

			allKubeconfigContextNames = append(allKubeconfigContextNames, contexts...)
			for _, context := range contexts {
				writeToContextToPathMapping(context, channelResult.kubeconfigPath)
			}
		}

		// write index file as soon as the path discovery is complete
		index := types.Index{
			Kind:                 store.getKind(),
			ContextToPathMapping: contextToPathMapping,
		}

		if err := writeIndex(index, indexFilePath); err != nil {
			logger.Warnf("failed to write index file to speed up future fuzzy searches: %v", err)
			return nil
		}

		state := types.IndexState{
			Kind:           store.getKind(),
			LastUpdateTime: time.Now().UTC(),
		}

		if err := writeIndexStoreState(state, indexStatePath); err != nil {
			logger.Warnf("failed to write index state file: %v", err)
		}

		return nil
	}()

	return showFuzzySearch(store, showPreview)
}

func showFuzzySearch(store store, showPreview bool) (string, string, error) {
	// display selection dialog for all the kubeconfig context names
	idx, err := fuzzyfinder.Find(
		&allKubeconfigContextNames,
		func(i int) string {
			return allKubeconfigContextNames[i]
		},
		fuzzyfinder.WithHotReload(),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if !showPreview || i == -1 {
				return ""
			}

			// read the content of the kubeconfig here and display
			currentContextName := allKubeconfigContextNames[i]
			kubeconfig, err := store.getSanitizedKubeconfigForContext(currentContextName)
			if err != nil {
				logger.Warnf("failed to show preview: %v", err)
				return ""
			}
			return kubeconfig
		}))

	if err != nil {
		return "", "", err
	}

	// map selection back to kubeconfig
	selectedContext := allKubeconfigContextNames[idx]
	kubeconfigPath := readFromContextToPathMapping(selectedContext)

	return kubeconfigPath, selectedContext, nil
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

func getContextsFromKubeconfig(path string, kubeconfig *types.KubeConfig) ([]string, error) {
	// get parent folder name
	parentFoldername := filepath.Base(filepath.Dir(path))
	return getContextNames(kubeconfig, parentFoldername), nil
}

func parseKubeconfig(data []byte) (*types.KubeConfig, error) {
	config := types.KubeConfig{}

	// unmarshal in a form that does not include the credentials
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal kubeconfig: %v", err)
	}
	return &config, nil
}

// sets the parent folder name to each context in the kubeconfig file
func getContextNames(config *types.KubeConfig, parentFoldername string) []string {
	var contextNames []string
	for _, context := range config.Contexts {
		split := strings.Split(context.Name, separator)
		if len(split) > 1 {
			// already has the directory name in there. override it in case it changed
			contextNames = append(contextNames, fmt.Sprintf("%s%s%s", parentFoldername, separator, split[len(split)-1]))
		} else {
			contextNames = append(contextNames, fmt.Sprintf("%s%s%s", parentFoldername, separator, context.Name))
		}
	}
	return contextNames
}

func setCurrentContext(kubeconfigData []byte, ctxnName string) (string, error) {
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
	tempDir := os.ExpandEnv(temporaryKubeconfigDir)
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
		return file.Name(), appendCurrentContext(file.Name(), currentContext)
	}

	return file.Name(), nil
}

func appendCurrentContext(kubeconfigPath, currentContext string) error {
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
