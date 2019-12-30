package main

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

const SEPERATOR = "/"


/* ATTENTION: The fuzzyfinder (dialog for selection) cannot be displayed in Goland IDE.
	1) build the project an run it : make build
 	2) ./kubectlSwitch  <pathToKubectlConfig1> <pathToKubectlConfig2>
*/

func main() {
	// arguments: all kubeconfig filepaths
	args := os.Args[1:]
	if len(args) == 0 || len(args[0]) == 0{
		return
	}

	var paths []string

	if len(args) == 1 {
		// split the arguments into paths
		paths = strings.Split(args[0], " ")
	} else {
		// already handed over several paths as arguments
		paths = args
	}

	contextToPathMapping := make(map[string]string)
	var allKubeconfigContextNames []string
	for _, path := range paths {
		contexts, err := getContextsFromKubeconfig(path)
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
			if i == -1 {
				return ""
			}
			// read the content of the kubeconfig here and display
			currentContextName := allKubeconfigContextNames[i]
			kubeconfigPath := contextToPathMapping[currentContextName]
			data, err := ioutil.ReadFile(kubeconfigPath)
			if err != nil {
				return ""
			}
			return string(data)
		}))
	if err != nil {
		log.Fatal(err)
	}

	// map selection back to Kubeconfig
	currentContextName := allKubeconfigContextNames[idx]
	kubeconfigPath := contextToPathMapping[currentContextName]

	// set current context in this kubeconfig
	config, err := parseKubeconfig(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	// current context for selection was in format <parentFolder>/<name>   -> we only need the name
	split := strings.Split(currentContextName, SEPERATOR)
	config.CurrentContext = split[len(split) -1]

	//migration from folder/name to name
	for i, context := range config.Contexts {
		split := strings.Split(context.Name, SEPERATOR)
		var s string
		if len(split) > 0 {
			s = split[len(split)-1]
		} else {
			s = split[0]
		}
		config.Contexts[i].Name = s
	}

	// write back the currentContext to the kubeconfig
	bytes, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = ioutil.WriteFile(kubeconfigPath, bytes, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// set the selected KubeconfigPath as the KUBECONFIG Environment variable
	if err := os.Setenv("KUBECONFIG", kubeconfigPath); err != nil {
		log.Fatalf("error setting env variable: %v", err)
	}

	// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG ENV Variable
	fmt.Print(kubeconfigPath)
}

func getContextsFromKubeconfig(path string) ([]string, error) {
	config, err := parseKubeconfig(path)
	if err != nil {
		return nil, err
	}
	// get parent folder name
	parentFoldername := filepath.Base(filepath.Dir(path))
	return getContextNames(config, parentFoldername), nil
}

func parseKubeconfig(path string) (*types.KubeConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file with path '%s': %v", path, err)
	}
	config := types.KubeConfig{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, fmt.Errorf("could not get contexts from kubeconfig with path '%s': %v", path, err)
	}
	return &config, nil
}

// sets the parent folder name to each context in the kubeconfig file
func getContextNames(config *types.KubeConfig, parentFoldername string) []string {
	var contextNames []string
	for _, context := range config.Contexts {
		split := strings.Split(context.Name, SEPERATOR)
		if len(split) > 1 {
			// already has the directory name in there. override it in case it changed
			contextNames = append(contextNames, fmt.Sprintf("%s%s%s", parentFoldername, SEPERATOR, split[len(split) -1]))
		} else {
			contextNames = append(contextNames, fmt.Sprintf("%s%s%s", parentFoldername, SEPERATOR, context.Name))
		}
	}
	return contextNames
}