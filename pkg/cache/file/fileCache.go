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

package file

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielfoehrkn/kubeswitch/pkg/cache"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const cacheKey = "filesystem"
const kubeconfigSuffix = "cache"

func init() {
	cache.Register(cacheKey, New)
}

func New(upstream store.KubeconfigStore, ccfg *types.Cache) (store.KubeconfigStore, error) {
	if ccfg == nil {
		return nil, fmt.Errorf("cache config must be provided for file cache")
	}
	cfg, err := unmarshalFileCacheCfg(ccfg.Config)
	if err != nil {
		return nil, err
	}

	cfgStore := types.KubeconfigStore{}
	if len(cfg.Path) == 0 {
		return nil, fmt.Errorf("path for filesystem cache was not configured")
	}
	cfgStore.Paths = []string{cfg.Path}

	log := logrus.New().WithField("store", types.StoreKindFilesystem).WithField("cache", cacheKey)

	return &fileCache{
		upstream: upstream,
		cfg:      cfg,
		logger:   log,
	}, nil
}

type fileCache struct {
	upstream store.KubeconfigStore
	cfg      fileCacheCfg
	logger   *logrus.Entry
}

func unmarshalFileCacheCfg(cfg interface{}) (fileCacheCfg, error) {
	var fileCacheCfg fileCacheCfg
	if cfg == nil {
		return fileCacheCfg, fmt.Errorf("cache is not configured")
	}
	buf, err := yaml.Marshal(cfg)
	if err != nil {
		return fileCacheCfg, fmt.Errorf("failed to marshal cache config: %w", err)
	}
	err = yaml.Unmarshal(buf, &fileCacheCfg)
	if err != nil {
		return fileCacheCfg, fmt.Errorf("cache config is invalid: %w", err)
	}
	return fileCacheCfg, nil
}

type fileCacheCfg struct {
	// Path to store the kubeconfigs in.
	Path string `yaml:"path"`
}

// hash for provided path
// the hash does not contain any folders or special characters and is safe to use as filename
func (c *fileCache) hash(path string) string {
	filename := md5.Sum([]byte(path))
	return fmt.Sprintf("%x", filename)
}

// suffix contains the UID of the Upstream store with a suffix kubeconfigSuffix"
func (c *fileCache) suffix() string {
	return fmt.Sprintf(".%s.%s", c.upstream.GetID(), kubeconfigSuffix)
}

// GetKubeconfigForPath returns the kubeconfig for the given path.
// First, it checks if the kubeconfig is already available in cache.
// If not, it is loaded from the upstream store and stored in cache
func (c *fileCache) GetKubeconfigForPath(path string) ([]byte, error) {
	c.logger.Debugf("Looking for '%s'", path)

	// check if kubeconfig is already available in the cache
	cacheFilename := fmt.Sprintf("%s%s", c.hash(path), c.suffix())
	file := filepath.Join(c.cfg.Path, cacheFilename)
	file = util.ExpandEnv(file)

	k, err := kubeconfigutil.NewKubeconfigForPath(file)
	if err == nil { // return cached kubeconfig if found
		c.logger.Debugf("kubeconfig found in cache '%s'", path)
		return k.GetBytes()
	}
	c.logger.Debugf("kubeconfig not found in cache '%s'", path)
	// kubeconfig not found in cache, load from upstream store
	kubeconfig, err := c.upstream.GetKubeconfigForPath(path)
	if err != nil { // if the upstream returns an error, the result is not cached
		return kubeconfig, err
	}

	// store the kubeconfig in the cache
	k, err = kubeconfigutil.New(kubeconfig, file, false)
	if err != nil {
		c.logger.Debugf("failure '%s' , %s", path, err)
		return nil, fmt.Errorf("failed to store kubeconfig in cache: %w", err)
	}
	_, err = k.WriteKubeconfigFile()
	return kubeconfig, err
}

// Flush cache by deleting all files in the cache directory
func (c *fileCache) Flush() (int, error) {
	path := util.ExpandEnv(c.cfg.Path)
	files, _ := os.ReadDir(path)
	deleted := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), c.suffix()) {
			continue
		}
		err := os.Remove(filepath.Join(path, f.Name()))
		if err != nil {
			return deleted, fmt.Errorf("failed to delete file '%s': %w", f.Name(), err)
		}
		deleted++
	}
	return deleted, nil
}

// passthru requests to the upstream store

func (c *fileCache) GetID() string {
	return c.upstream.GetID()
}

func (c *fileCache) GetKind() types.StoreKind {
	return c.upstream.GetKind()
}

func (c *fileCache) GetContextPrefix(path string) string {
	return c.upstream.GetContextPrefix(path)
}

func (c *fileCache) VerifyKubeconfigPaths() error {
	return c.upstream.VerifyKubeconfigPaths()
}

func (c *fileCache) StartSearch(channel chan store.SearchResult) {
	c.upstream.StartSearch(channel)
}

func (c *fileCache) GetLogger() *logrus.Entry {
	return c.upstream.GetLogger()
}
func (c *fileCache) GetStoreConfig() types.KubeconfigStore {
	return c.upstream.GetStoreConfig()
}
