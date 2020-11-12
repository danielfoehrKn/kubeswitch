package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
	"gopkg.in/yaml.v2"

	"github.com/danielfoehrkn/kubectlSwitch/types"
)

func (s *FileStore) getKind() types.StoreKind {
	return types.StoreKindFilesystem
}

func (s *FileStore) discoverPaths(searchPath string, kubeconfigName string, channel chan channelResult) {
	var kubeconfigPaths []string

	if err := godirwalk.Walk(searchPath, &godirwalk.Options{
		Callback: func(osPathname string, _ *godirwalk.Dirent) error {
			fileName := filepath.Base(osPathname)
			matched, err := filepath.Match(kubeconfigName, fileName)
			if err != nil {
				return err
			}
			if matched {
				kubeconfigPaths = append(kubeconfigPaths, osPathname)
				channel <- channelResult{
					kubeconfigPath: osPathname,
					error:          nil,
				}
			}
			return nil
		},
		Unsorted: false, // (optional) set true for faster yet non-deterministic enumeration
	}); err != nil {
		channel <- channelResult{
			kubeconfigPath: "",
			error:          fmt.Errorf("failed to find kubeconfig files in directory: %v", err),
		}
	}
}

func (s *FileStore) getContextsForPath(path string) ([]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file with path '%s': %v", path, err)
	}

	// parse into struct that does not contain the credentials
	config, err := parseKubeconfig(data)
	if err != nil {
		return nil, fmt.Errorf("could not parse Kubeconfig with path '%s': %v", path, err)
	}

	kubeconfigData, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("could not marshal kubeconfig with path '%s': %v", path, err)
	}

	// save kubeconfig content to in-memory map to avoid duplicate read operation in getSanitizedKubeconfigForContext
	writeToPathToKubeconfig(path, string(kubeconfigData))

	return getContextsFromKubeconfig(path, config)
}

func (s *FileStore) getSanitizedKubeconfigForContext(contextName string) (string, error) {
	path := readFromContextToPathMapping(contextName)

	// during first run without index, the files are already read in the getContextsForPath and save in-memory
	kubeconfig := readFromPathToKubeconfig(path)
	if len(kubeconfig) > 0 {
		return kubeconfig, nil
	}

	// kubeconfig not yet saved in in-memory map -> load from filesystem
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read file with path '%s': %v", path, err)
	}

	config, err := parseKubeconfig(data)
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

func (s *FileStore) getKubeconfigForPath(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (s *FileStore) checkPath(kubeconfigDirectory string) error {
	if _, err := os.Stat(kubeconfigDirectory); os.IsNotExist(err) {
		return fmt.Errorf("the kubeconfig directory %q does not exist", kubeconfigDirectory)
	}
	return nil
}
