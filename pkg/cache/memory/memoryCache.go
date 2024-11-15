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

package memory

import (
	"github.com/danielfoehrkn/kubeswitch/pkg/cache"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/sirupsen/logrus"
)

func init() {
	cache.Register("memory", New)
}

func New(upstream storetypes.KubeconfigStore, _ *types.Cache) (storetypes.KubeconfigStore, error) {
	return &memoryCache{
		upstream: upstream,
		cache:    make(map[string][]byte),
	}, nil
}

type memoryCache struct {
	upstream storetypes.KubeconfigStore
	cache    map[string][]byte
}

// GetKubeconfigForPath implements the store.KubeconfigStore interface.
// It is a wrapper around a KubeConfigCache.
// It intercepts calls to GetKubeconfigForPath and caches the result in memory.
func (c *memoryCache) GetKubeconfigForPath(path string, tags map[string]string) ([]byte, error) {
	if val, ok := c.cache[path]; ok {
		c.GetLogger().Debugf("GetKubeconfigForPath: %s found in cache", path)
		return val, nil
	}
	c.GetLogger().Debugf("GetKubeconfigForPath: %s not cached", path)
	kube, err := c.upstream.GetKubeconfigForPath(path, tags)
	if err != nil {
		return kube, err
	}
	c.cache[path] = kube
	return kube, nil
}

func (c *memoryCache) GetID() string {
	return c.upstream.GetID()
}

func (c *memoryCache) GetKind() types.StoreKind {
	return c.upstream.GetKind()
}

func (c *memoryCache) GetContextPrefix(path string) string {
	return c.upstream.GetContextPrefix(path)
}

func (c *memoryCache) VerifyKubeconfigPaths() error {
	return c.upstream.VerifyKubeconfigPaths()
}

func (c *memoryCache) StartSearch(channel chan storetypes.SearchResult) {
	c.upstream.StartSearch(channel)
}

func (c *memoryCache) GetLogger() *logrus.Entry {
	return c.upstream.GetLogger()
}

func (c *memoryCache) GetStoreConfig() types.KubeconfigStore {
	return c.upstream.GetStoreConfig()
}

func (c *memoryCache) GetSearchPreview(path string, optionalTags map[string]string) (string, error) {
	previewer, ok := c.upstream.(storetypes.Previewer)
	if !ok {
		// if the wrapped store is not a previewer, simply return an empty string, hence causing no visual distortion
		return "", nil
	}

	return previewer.GetSearchPreview(path, optionalTags)
}
