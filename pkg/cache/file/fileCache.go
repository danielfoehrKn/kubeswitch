package file

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielfoehrkn/kubeswitch/pkg/cache"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
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

// TODO: pass CacheConfig to constructor instead of rawCfg
func New(upstream store.KubeconfigStore, rawCfg interface{}) (store.KubeconfigStore, error) {
	cfg, err := unmarshalFileCacheCfg(rawCfg)
	if err != nil {
		return nil, err
	}

	cfgStore := types.KubeconfigStore{}
	if len(cfg.Path) == 0 {
		return nil, fmt.Errorf("path for filesystem cache was not configured")
	}
	cfgStore.Paths = []string{cfg.Path}

	log := logrus.New().WithField("store", types.StoreKindFilesystem).WithField("cache", cacheKey)
	log.Logger.SetLevel(logrus.DebugLevel)

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

// create a filename for provided path
// the filename does not contain any folders or special characters
// As suffix the UID of the Upstream store is added and ".cache"
func (c *fileCache) filename(path string) string {
	filename := md5.Sum([]byte(path))
	return fmt.Sprintf("%x.%s.%s", filename, c.upstream.GetID(), kubeconfigSuffix)

}

// fileCache implements the store.KubeconfigStore interface and intercepts calls to the
// upstream store.
// First, it checks if the kubeconfig is already available.
// If not, it is loaded from the upstream store and stored.
func (c *fileCache) GetKubeconfigForPath(path string) ([]byte, error) {
	c.logger.Debugf("Looking for '%s'", path)

	// check if kubeconfig is already available in the cache
	cachedFile := c.filename(path)
	file := filepath.Join(c.cfg.Path, cachedFile)
	//TODO: Why is this needed????e
	file = strings.ReplaceAll(file, "~", "$HOME")
	file = os.ExpandEnv(file)

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
	c.logger.Debugf("StartSearch")
	c.upstream.StartSearch(channel)
}

func (c *fileCache) GetLogger() *logrus.Entry {
	return c.upstream.GetLogger()
}
func (c *fileCache) GetStoreConfig() types.KubeconfigStore {
	return c.upstream.GetStoreConfig()
}
