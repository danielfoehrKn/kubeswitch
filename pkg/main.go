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

	"github.com/danielfoehrkn/kubectlSwitch/pkg/config"
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
	allKubeconfigContextNames []string

	contextToPathMapping     = make(map[string]string)
	contextToPathMappingLock = sync.RWMutex{}

	pathToKubeconfig     = make(map[string]string)
	pathToKubeconfigLock = sync.RWMutex{}
)

func Switcher(log *logrus.Entry, kubeconfigStore store.KubeconfigStore, configPath string, stateDir string, showPreview bool) error {
	if err := kubeconfigStore.CheckRootPath(); err != nil {
		return err
	}

	searchIndex, err := index.New(log, kubeconfigStore.GetKind(), stateDir)
	if err != nil {
		return err
	}

	shouldReadFromIndex := false
	if searchIndex.HasContent() && searchIndex.HasKind(kubeconfigStore.GetKind()) {
		switchConfig, err := config.LoadConfigFromFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to check if should read from index file: %v", err)
		}

		// found an index for the correct Store kind
		// check if should use existing index or not
		shouldReadFromIndex, err = searchIndex.ShouldBeUsed(switchConfig)
		if err != nil {
			return err
		}
	}

	var kubeconfigPath, selectedContext string
	if shouldReadFromIndex {
		kubeconfigPath, selectedContext, err = fuzzySearchFromIndex(log, kubeconfigStore, *searchIndex, showPreview)
		if err != nil {
			return err
		}
	} else {
		kubeconfigPath, selectedContext, err = fuzzySearch(log, kubeconfigStore, *searchIndex, showPreview)
		if err != nil {
			return err
		}
	}

	// remove the folder name from the context name
	split := strings.Split(selectedContext, "/")
	selectedContext = split[len(split)-1]

	kubeconfigData, err := kubeconfigStore.GetKubeconfigForPath(log, kubeconfigPath)
	if err != nil {
		return err
	}

	tempKubeconfigPath, err := setCurrentContextOnTemporaryKubeconfigFile(kubeconfigData, selectedContext)
	if err != nil {
		return fmt.Errorf("failed to write current context to temporary kubeconfig: %v", err)
	}

	// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG ENV Variable
	fmt.Print(tempKubeconfigPath)

	return nil
}

func fuzzySearchFromIndex(log *logrus.Entry, store store.KubeconfigStore, searchIndex index.SearchIndex, showPreview bool) (string, string, error) {
	contextToPathMapping = searchIndex.GetContent()

	// build all KubeconfigContextNames from pre-computed index
	allKubeconfigContextNames = make([]string, len(contextToPathMapping))
	i := 0
	for k := range contextToPathMapping {
		allKubeconfigContextNames[i] = k
		i++
	}

	return showFuzzySearch(log, store, showPreview)
}

func fuzzySearch(log *logrus.Entry, kubeconfigStore store.KubeconfigStore, searchIndex index.SearchIndex, showPreview bool) (string, string, error) {
	var channel = make(chan store.PathDiscoveryResult)

	go func() {
		// only close when directory search is over, otherwise send on closed channel
		defer close(channel)
		kubeconfigStore.DiscoverPaths(log, channel)
	}()

	go func() error {
		for channelResult := range channel {
			if channelResult.Error != nil {
				log.Warnf("error returned from path discovery: %v", channelResult.Error)
				continue
			}

			// get the context names from the parsed kubeconfig
			contexts, err := getContextsForKubeconfigPath(log, kubeconfigStore, channelResult.KubeconfigPath)
			if err != nil {
				// do not throw Error, try to parse the other files
				// this will happen a lot when using vault as storage because the secrets key value needs to matche the desired kubeconfig name
				// this however cannot be checked without retrieving the actual secret (path discovery is only list operation)
				continue
			}

			allKubeconfigContextNames = append(allKubeconfigContextNames, contexts...)
			for _, context := range contexts {
				writeToContextToPathMapping(context, channelResult.KubeconfigPath)
			}
		}

		// write index file as soon as the path discovery is complete
		index := types.Index{
			Kind:                 kubeconfigStore.GetKind(),
			ContextToPathMapping: contextToPathMapping,
		}

		if err := searchIndex.Write(index); err != nil {
			log.Warnf("failed to write index file to speed up future fuzzy searches: %v", err)
			return nil
		}

		indexStateToWrite := types.IndexState{
			Kind:           kubeconfigStore.GetKind(),
			LastUpdateTime: time.Now().UTC(),
		}

		if err := searchIndex.WriteState(indexStateToWrite); err != nil {
			log.Warnf("failed to write index state file: %v", err)
		}

		return nil
	}()

	return showFuzzySearch(log, kubeconfigStore, showPreview)
}

func showFuzzySearch(log *logrus.Entry, kubeconfigStore store.KubeconfigStore, showPreview bool) (string, string, error) {
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
			kubeconfig, err := getSanitizedKubeconfigForContext(log, kubeconfigStore, currentContextName)
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
	selectedContext := allKubeconfigContextNames[idx]
	kubeconfigPath := readFromContextToPathMapping(selectedContext)

	return kubeconfigPath, selectedContext, nil
}

func getContextsForKubeconfigPath(log *logrus.Entry, kubeconfigStore store.KubeconfigStore, kubeconfigPath string) ([]string, error) {
	bytes, err := kubeconfigStore.GetKubeconfigForPath(log, kubeconfigPath)
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

	// save kubeconfig content to in-memory map to avoid duplicate read operation in getSanitizedKubeconfigForContext
	writeToPathToKubeconfig(kubeconfigPath, string(kubeconfigData))
	return getContextsFromKubeconfig(kubeconfigStore.GetKind(), kubeconfigPath, config)
}

func getSanitizedKubeconfigForContext(log *logrus.Entry, kubeconfigStore store.KubeconfigStore, contextName string) (string, error) {
	path := readFromContextToPathMapping(contextName)

	// during first run without index, the files are already read in the getContextsForKubeconfigPath and save in-memory
	kubeconfig := readFromPathToKubeconfig(path)
	if len(kubeconfig) > 0 {
		return kubeconfig, nil
	}

	data, err := kubeconfigStore.GetKubeconfigForPath(log, path)
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

	// save kubeconfig content to in-memory map to avoid duplicate read operation in getSanitizedKubeconfigForContext
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
