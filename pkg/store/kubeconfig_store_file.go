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

package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/karrick/godirwalk"
	"github.com/sirupsen/logrus"
)

func (s *FilesystemStore) GetKind() types.StoreKind {
	return types.StoreKindFilesystem
}

func (s *FilesystemStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *FilesystemStore) StartSearch(channel chan SearchResult) {
	for _, path := range s.kubeconfigFilepaths {
		channel <- SearchResult{
			KubeconfigPath: path,
			Error:          nil,
		}
	}

	wg := sync.WaitGroup{}
	for _, path := range s.kubeconfigDirectories {
		wg.Add(1)
		go s.searchDirectory(&wg, path, channel)
	}
	wg.Wait()
}

func (s *FilesystemStore) searchDirectory(wg *sync.WaitGroup, searchPath string, channel chan SearchResult) {
	defer wg.Done()

	if err := godirwalk.Walk(searchPath, &godirwalk.Options{
		Callback: func(osPathname string, _ *godirwalk.Dirent) error {
			fileName := filepath.Base(osPathname)
			matched, err := filepath.Match(s.KubeconfigName, fileName)
			if err != nil {
				return err
			}
			if matched {
				channel <- SearchResult{
					KubeconfigPath: osPathname,
					Error:          nil,
				}
			}
			return nil
		},
		Unsorted: false, // (optional) set true for faster yet non-deterministic enumeration
	}); err != nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("failed to find kubeconfig files in directory: %v", err),
		}
	}
}

func (s *FilesystemStore) GetKubeconfigForPath(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (s *FilesystemStore) VerifyKubeconfigPaths() error {
	var (
		duplicatePath              = make(map[string]*struct{})
		validKubeconfigFilepaths   []string
		validKubeconfigDirectories []string
		usr, _                     = user.Current()
		homeDir                    = usr.HomeDir
	)

	for _, path := range s.KubeconfigStore.Paths {
		// do not add duplicate paths
		if duplicatePath[path] != nil {
			continue
		}
		duplicatePath[path] = &struct{}{}

		kubeconfigPath := path
		if kubeconfigPath == "~" {
			kubeconfigPath = homeDir
		} else if strings.HasPrefix(kubeconfigPath, "~/") {
			// Use strings.HasPrefix so we don't match paths like
			// "/something/~/something/"
			kubeconfigPath = filepath.Join(homeDir, kubeconfigPath[2:])
		}

		info, err := os.Stat(kubeconfigPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("the configured kubeconfig directory %q does not exist", path)
		} else if err != nil {
			return fmt.Errorf("failed to read from the configured kubeconfig directory %q: %v", path, err)
		}

		if info.IsDir() {
			validKubeconfigDirectories = append(validKubeconfigDirectories, kubeconfigPath)
			continue
		}
		validKubeconfigFilepaths = append(validKubeconfigFilepaths, kubeconfigPath)
	}

	if len(validKubeconfigDirectories) == 0 && len(validKubeconfigFilepaths) == 0 {
		return fmt.Errorf("none of the %d specified kubeconfig path(s) exist. Either specifiy an existing path via flag '--kubeconfig-path' or in the switch config file", len(s.KubeconfigStore.Paths))
	}
	s.kubeconfigDirectories = validKubeconfigDirectories
	s.kubeconfigFilepaths = validKubeconfigFilepaths
	return nil
}
