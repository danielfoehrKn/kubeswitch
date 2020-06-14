package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/danielfoehrkn/kubectlSwitch/types"

	"github.com/ktr0731/go-fuzzyfinder"
	"gopkg.in/yaml.v2"
)

const separator = "/"
const kubeconfigCurrentContext = "current-context:"
const temporaryKubeconfigDir = "$HOME/.kube/.switch_tmp"

func Switcher(kubeconfigDirectory, kubeconfigName string, showPreview bool) error {
	kubeconfigPaths, err := findFilePathsInDirectory(kubeconfigDirectory, kubeconfigName)
	if err != nil {
		return err
	}

	if len(kubeconfigPaths) == 0 {
		return fmt.Errorf("no kubeconfig files with name %q found in directory %q", kubeconfigName, kubeconfigDirectory)
	}

	contextToPathMapping := make(map[string]string)
	pathToKubeconfig := make(map[string]string)
	var allKubeconfigContextNames []string
	for _, path := range kubeconfigPaths {
		kubeconfig, err := parseKubeconfig(path)
		if err != nil {
			// do not throw error, try to parse the other files
			continue
		}

		// marshal the sanitized kubeconfig for the preview
		kubeconfigData, err := yaml.Marshal(kubeconfig)
		if err != nil {
			// do not throw error, try to parse the other files
			continue
		}
		pathToKubeconfig[path] = string(kubeconfigData)

		// get the context names form the parsed kubeconfig
		contexts, err := getContextsFromKubeconfig(path, kubeconfig)
		if err != nil {
			// do not throw error, try to parse the other files
			continue
		}

		allKubeconfigContextNames = append(allKubeconfigContextNames, contexts...)
		for _, context := range contexts {
			contextToPathMapping[context] = path
		}
	}
	sort.Strings(allKubeconfigContextNames)

	// display selection dialog for all the kubeconfig context names
	idx, err := fuzzyfinder.Find(
		allKubeconfigContextNames,
		func(i int) string {
			return allKubeconfigContextNames[i]
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if !showPreview || i == -1{
				return ""
			}

			// read the content of the kubeconfig here and display
			currentContextName := allKubeconfigContextNames[i]
			kubeconfigPath := contextToPathMapping[currentContextName]
			kubeconfig := pathToKubeconfig[kubeconfigPath]
			return kubeconfig
		}))
	if err != nil {
		log.Fatal(err)
	}

	// map selection back to kubeconfig
	selectedContext := allKubeconfigContextNames[idx]
	kubeconfigPath := contextToPathMapping[selectedContext]

	// remove the folder name from the context name
	split := strings.Split(selectedContext, separator)
	selectedContext = split[len(split)-1]

	tempKubeconfigPath, err := setCurrentContext(kubeconfigPath, selectedContext)
	if err != nil {
		log.Fatalf("failed to write current context to kubeconfig: %v", err)
	}

	// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG ENV Variable
	fmt.Print(tempKubeconfigPath)

	return nil
}

func getContextsFromKubeconfig(path string, kubeconfig *types.KubeConfig) ([]string, error) {
	// get parent folder name
	parentFoldername := filepath.Base(filepath.Dir(path))
	return getContextNames(kubeconfig, parentFoldername), nil
}

func parseKubeconfig(path string) (*types.KubeConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file with path '%s': %v", path, err)
	}
	config := types.KubeConfig{}

	// unmarshal in a form that does not include the credentials
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal kubeconfig with path '%s': %v", path, err)
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

func setCurrentContext(kubeconfigPath string, ctxnName string) (string, error) {
	currentContext := fmt.Sprintf("%s %s", kubeconfigCurrentContext, ctxnName)

	input, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	foundCurrentContext := false
	for i, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.Contains(line, kubeconfigCurrentContext) {
			foundCurrentContext = true
			lines[i] = currentContext
		}
	}

	output := strings.Join(lines, "\n")
	tempDir := os.ExpandEnv(temporaryKubeconfigDir)
	err = os.Mkdir(tempDir, 0700)
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

func findFilePathsInDirectory(dir, kubeconfigName string) ([]string, error) {
	var kubeconfigPaths []string
	if err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == kubeconfigName{
				kubeconfigPaths = append(kubeconfigPaths, path)
			}
			return nil
		}); err != nil {
		return nil, fmt.Errorf("failed to find kubeconfig files in directory: %v", err)
	}
	return kubeconfigPaths, nil
}
