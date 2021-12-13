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

package ns

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	// namespaceSubdirectory is the name of the sub-directory within the switch state containing the
	// namespaces index for each context
	namespaceSubdirectory = "namespace"
)

type NamespaceCache struct {
	cacheFilepath string
	content       []string
}

// NewNamespaceCache creates a new NamespaceCache
func NewNamespaceCache(stateDirectory string, contextName string) (*NamespaceCache, error) {
	namespaceStateDirectory := fmt.Sprintf("%s/%s", stateDirectory, namespaceSubdirectory)
	if _, err := os.Stat(namespaceStateDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(namespaceStateDirectory, 0755); err != nil {
			return nil, err
		}
	}

	// sanitize from / that separates the prefix from the actual context name
	contextName = strings.ReplaceAll(contextName, "/", "")
	cacheFilepath := fmt.Sprintf("%s/%s", namespaceStateDirectory, contextName)

	i := NamespaceCache{
		cacheFilepath: cacheFilepath,
	}

	indexFromFile, err := i.loadFromFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	i.content = indexFromFile
	return &i, nil
}

func (i *NamespaceCache) HasContent() bool {
	return i.content != nil || len(i.content) == 0
}

func (i *NamespaceCache) GetContent() []string {
	if i.content == nil {
		return []string{}
	}
	return i.content
}

// LoadIndexFromFile takes a filename and reads the contents into a []string.
func (i *NamespaceCache) loadFromFile() ([]string, error) {
	// an index file is not required. Its ok if it does not exist.
	if _, err := os.Stat(i.cacheFilepath); err != nil {
		return nil, err
	}

	file, err := os.Open(i.cacheFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append([]string{scanner.Text()}, lines...)
	}
	return lines, scanner.Err()
}

func (i *NamespaceCache) Write(toWrite []string) error {
	// creates or truncate/clean the existing file
	file, err := os.Create(i.cacheFilepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// print values to f, one per line
	for _, value := range toWrite {
		_, err = fmt.Fprintln(file, value)
		if err != nil {
			return err
		}
	}

	return nil
}
