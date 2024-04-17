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

package store

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	gardenclient "github.com/danielfoehrkn/kubeswitch/pkg/store/gardener/copied_gardenctlv2"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
	"github.com/disiqueira/gotree"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	seedmanagementv1alpha1 "github.com/gardener/gardener/pkg/apis/seedmanagement/v1alpha1"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gardenerstore "github.com/danielfoehrkn/kubeswitch/pkg/store/gardener"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/state"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/danielfoehrkn/kubeswitch/types"
)

const (
	// CmNameClusterIdentity is the config map name that contains the gardener cluster identity
	CmNameClusterIdentity = "cluster-identity"
	// KeyClusterIdentity is the key in the cluster identity config map
	KeyClusterIdentity = CmNameClusterIdentity
	// AllNamespacesDenominator is a character that indicates that all Shoot clusters should be considered for the search
	AllNamespacesDenominator = "/"
	// defaultGardenloginConfigPath is the path to the default gardenlogin config
	defaultGardenloginConfigPath = "$HOME/.garden/gardenlogin.yaml"
)

// GardenloginConfig represents the config for the Gardenlogin-exec-provider that is
// required to work with the kubeconfig files obtained from the GardenConfig cluster
// If missing, this configuration is generated based on the Kubeswitch config
type GardenloginConfig struct {
	// Gardens is a list of known GardenConfig clusters
	Gardens []GardenConfig `yaml:"gardens"`
}

// GardenConfig holds the config of a garden cluster
type GardenConfig struct {
	// Identity is the cluster identity of the garden cluster.
	// See cluster-identity ConfigMap in kube-system namespace of the garden cluster
	Identity string `yaml:"identity"`

	// Kubeconfig holds the path for the kubeconfig of the garden cluster
	Kubeconfig string `yaml:"kubeconfig"`
}

// NewGardenerStore creates a new Gardener store
func NewGardenerStore(store types.KubeconfigStore, stateDir string) (*GardenerStore, error) {
	config, err := gardenerstore.GetStoreConfig(store)
	if err != nil {
		return nil, err
	}

	var landscapeName string
	if config != nil && config.LandscapeName != nil {
		landscapeName = *config.LandscapeName
	}

	return &GardenerStore{
		Logger:                   logrus.New().WithField("store", types.StoreKindGardener),
		KubeconfigStore:          store,
		Config:                   config,
		LandscapeName:            landscapeName,
		StateDirectory:           stateDir,
		PathToShootLock:          sync.RWMutex{},
		PathToManagedSeedLock:    sync.RWMutex{},
		CaSecretNameToSecretLock: sync.RWMutex{},
	}, nil
}

// InitializeGardenerStore initializes the store using the provided Gardener kubeconfig
// decoupled from the NewGardenerStore() to be called when starting the search to reduce
// time when the CLI can start showing the fuzzy search
func (s *GardenerStore) InitializeGardenerStore() error {
	var err error
	s.Client, err = gardenerstore.GetGardenClient(s.Config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cm := &corev1.ConfigMap{}
	if err := s.Client.Get(ctx, client.ObjectKey{Name: CmNameClusterIdentity, Namespace: metav1.NamespaceSystem}, cm); err != nil {
		return fmt.Errorf("unable to get gardener landscape identity from config map %s/%s: %w", metav1.NamespaceSystem, CmNameClusterIdentity, err)
	}

	identity, ok := cm.Data[KeyClusterIdentity]
	if !ok {
		return fmt.Errorf("unable to get gardener landscape identity from config map %s/%s: data key %q not found", metav1.NamespaceSystem, CmNameClusterIdentity, KeyClusterIdentity)
	}
	s.LandscapeIdentity = identity
	s.GardenClient = gardenclient.NewGardenClient(s.Client, s.LandscapeIdentity)

	// possibly concurrent access to file when multiple stores
	// For now, ignore the env variables: `GL_HOME` & `GL_CONFIG_NAME` that could be used to set alternative config directories
	gardenloginConfigPath := os.ExpandEnv(defaultGardenloginConfigPath)
	if _, err := os.Stat(gardenloginConfigPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// the default configuration does not exist. Write based on the Kubeswitch configuration file
		if err := writeGardenloginConfig(gardenloginConfigPath, &GardenloginConfig{Gardens: []GardenConfig{
			{
				Identity:   s.LandscapeIdentity,
				Kubeconfig: s.Config.GardenerAPIKubeconfigPath,
			},
		}}); err != nil {
			return fmt.Errorf("failed to write Gardenlogin config: %v", err)
		}
		return nil
	}

	// if already exists, check if contains an entry with the specified landscape identity
	gardenloginConfig, err := getGardenloginConfig(gardenloginConfigPath)
	if err != nil {
		return err
	}

	foundEntry := false
	for _, entry := range gardenloginConfig.Gardens {
		if entry.Identity == s.LandscapeIdentity {
			foundEntry = true
			break
		}
	}

	if !foundEntry {
		gardenloginConfig.Gardens = append(gardenloginConfig.Gardens, GardenConfig{
			Identity:   s.LandscapeIdentity,
			Kubeconfig: s.Config.GardenerAPIKubeconfigPath,
		})
		if err := writeGardenloginConfig(gardenloginConfigPath, gardenloginConfig); err != nil {
			return fmt.Errorf("failed to write Gardenlogin config: %v", err)
		}
	}

	return nil
}

// getGardenloginConfig returns the GardenloginConfig from the provided filepath
func getGardenloginConfig(path string) (*GardenloginConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &GardenloginConfig{}
	if len(bytes) == 0 {
		return config, nil
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal Gardenlogin config file with path '%s': %v", path, err)
	}
	return config, nil
}

// writeGardenloginConfig writes the given gardenlogin config to path
func writeGardenloginConfig(path string, config *GardenloginConfig) error {
	// creates or truncate/clean the existing file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	output, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	_, err = file.Write(output)
	if err != nil {
		return err
	}
	return nil
}

// StartSearch starts the search for Shoots and Managed Seeds
func (s *GardenerStore) StartSearch(channel chan SearchResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.InitializeGardenerStore(); err != nil {
		err := fmt.Errorf("failed to initialize store. This is most likely a problem with your provided kubeconfig: %v", err)
		channel <- SearchResult{
			Error: err,
		}
		return
	}

	// specifying no path equals to all namespaces being searched
	if len(s.KubeconfigStore.Paths) == 0 {
		s.KubeconfigStore.Paths = []string{AllNamespacesDenominator}
	}

	var (
		shootList  []gardencorev1beta1.Shoot
		secretList []corev1.Secret
	)
	for _, path := range s.KubeconfigStore.Paths {
		var (
			shoots      *gardencorev1beta1.ShootList
			listOptions = client.ListOptions{}
		)

		if path != AllNamespacesDenominator {
			listOptions.Namespace = path
		}

		shoots, err := s.GardenClient.ListShoots(ctx, &listOptions)
		if err != nil {
			channel <- SearchResult{
				Error: fmt.Errorf("failed to call list Shoots from the Gardener API for namespace %q: %v", path, err),
			}
			return
		}

		selector := labels.SelectorFromSet(labels.Set{"gardener.cloud/role": "ca-cluster"})
		listOptions.LabelSelector = selector
		secrets := &corev1.SecretList{}
		if err := s.Client.List(ctx, secrets, &listOptions); err != nil {
			channel <- SearchResult{
				Error: fmt.Errorf("failed to list CA secrets for namespace %q: %v", path, err),
			}
			return
		}

		shootList = append(shootList, shoots.Items...)
		secretList = append(secretList, secrets.Items...)

		if path == AllNamespacesDenominator {
			break
		}
	}

	managedSeeds := &seedmanagementv1alpha1.ManagedSeedList{}
	if err := s.Client.List(ctx, managedSeeds, &client.ListOptions{}); err != nil {
		// do not return here as many older Gardener installations do not have the
		// resource group for managed seeds yet
		s.Logger.Debugf("failed to list managed seeds: %v", err)
	}

	// for memoization
	s.CacheCaSecretNameToSecret = make(map[string]corev1.Secret, len(secretList))
	for _, secret := range secretList {
		s.writeCacheCaSecretNameToSecretLock(fmt.Sprintf("%s:%s", secret.Namespace, secret.Name), secret)
	}

	s.sendKubeconfigPaths(channel, shootList, managedSeeds.Items)
}

func (s *GardenerStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	// the Gardener store encodes the path with semantic information
	// <landscape-name>--shoot-<project-name>--<shoot-name>
	// just use this semantic information as a prefix & remove the double dashes
	return strings.ReplaceAll(path, "--", "-")
}

// IsInitialized checks if the store has been initialized already
func (s *GardenerStore) IsInitialized() bool {
	return s.Client != nil && len(s.LandscapeIdentity) > 0
}

func (s *GardenerStore) GetID() string {
	id := "default"

	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	} else if s.Config != nil && s.Config.LandscapeName != nil {
		id = *s.Config.LandscapeName
	}

	return fmt.Sprintf("%s.%s", types.StoreKindGardener, id)
}

func (s *GardenerStore) GetKind() types.StoreKind {
	return types.StoreKindGardener
}

func (s *GardenerStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *GardenerStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *GardenerStore) GetControlplaneKubeconfigForShoot(shootName, project string) ([]byte, *string, error) {
	if !s.IsInitialized() {
		if err := s.InitializeGardenerStore(); err != nil {
			return nil, nil, fmt.Errorf("failed to initialize Gardener store: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shootNamespace := "garden"

	// for possible private seeds that exist outside the "garden" namespace
	if project != "garden" {
		shootNamespace = fmt.Sprintf("garden-%s", project)
	}

	// we get the Shoot for the managed Seed
	shoot, err := s.GardenClient.GetShoot(ctx, shootNamespace, shootName)
	if err != nil {
		return nil, nil, err
	}

	if shoot.Spec.SeedName == nil || *shoot.Spec.SeedName == "" {
		return nil, nil, fmt.Errorf("shoot %q has not yet been assigned to a seed", shootName)
	}

	if shoot.Status.TechnicalID == "" {
		return nil, nil, fmt.Errorf("no technicalID has been assigned to the shoot %q yet", shootName)
	}

	// this actually tries to find a ManagedSeed with the given name in the garden ns
	// then uses the Shoot referenced in the Managed seed to obtain the kubeconfig for the managed seed's Shoot
	clientConfig, err := s.GardenClient.GetSeedClientConfig(ctx, *shoot.Spec.SeedName)
	if err != nil {
		return nil, nil, err
	}

	// the shoot.Status.TechnicalID is the same as this shoot's controlplane namespace in the seed
	clientConfig, err = gardenerstore.ClientConfigWithNamespace(clientConfig, shoot.Status.TechnicalID)
	if err != nil {
		return nil, nil, err
	}

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, nil, err
	}

	bytes, err := clientcmd.Write(rawConfig)
	if err != nil {
		return nil, nil, err
	}

	config, err := kubeconfigutil.NewKubeconfig(bytes)
	if err != nil {
		return nil, nil, err
	}

	contextNames, err := config.GetContextNames()
	if err != nil {
		return nil, nil, err
	}

	for _, name := range contextNames {
		if strings.HasSuffix(name, "-internal") {
			if err := config.RemoveContext(name); err != nil {
				return nil, nil, fmt.Errorf("unable to remove internal kubeconfig context: %v", err)
			}
			break
		}
	}

	// add meta information to kubeconfig (ignored by kubectl)
	if err := config.SetGardenerStoreMetaInformation(s.LandscapeIdentity, string(gardenerstore.GardenerResourceSeed), "garden", *shoot.Spec.SeedName); err != nil {
		return nil, nil, err
	}

	bytes, err = config.GetBytes()
	if err != nil {
		return nil, nil, err
	}

	return bytes, shoot.Spec.SeedName, nil
}

func (s *GardenerStore) GetKubeconfigForPath(path string, _ map[string]string) ([]byte, error) {
	if !s.IsInitialized() {
		if err := s.InitializeGardenerStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize Gardener store: %w", err)
		}
	}

	if gardenerstore.GetGardenKubeconfigPath(s.LandscapeIdentity) == path {
		if s.Config == nil || len(s.Config.GardenerAPIKubeconfigPath) == 0 {
			return nil, fmt.Errorf("cannot get garden kubeconfig. Field 'gardenerAPIKubeconfigPath' is not configured in the Gardener store configuration in the SwitchConfig file")
		}
		return os.ReadFile(s.Config.GardenerAPIKubeconfigPath)
	}

	landscape, resource, name, namespace, gardenerProjectName, err := gardenerstore.ParseIdentifier(path)
	if err != nil {
		return nil, err
	}

	if landscape != s.LandscapeName && landscape != s.LandscapeIdentity {
		return nil, fmt.Errorf("unknown Gardener landscape %q", landscape)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var clientConfig clientcmd.ClientConfig
	switch resource {
	case gardenerstore.GardenerResourceSeed:
		managedSeed, ok := s.readFromCachePathToManagedSeed(path)

		// we know the namespace and name for the Shoot for the ManagedSeed
		// so we can use that knowledge to get the correct index to get the corresponding Shoot from the cache
		namespace = "garden"
		if ok {
			path = gardenerstore.GetShootIdentifier(landscape, "garden", managedSeed.Spec.Shoot.Name)
		}
		fallthrough
	case gardenerstore.GardenerResourceShoot:
		s.Logger.Debugf("Getting kubeconfig for %s (%s/%s)", resource, namespace, name)

		shoot, _ := s.readFromCachePathToShoot(path)
		caClusterSecretName := fmt.Sprintf("%s:%s.%s", namespace, name, gardenclient.ShootProjectSecretSuffixCACluster)
		caSecret, _ := s.readFromCacheCaSecretNameToSecretLock(caClusterSecretName)

		clientConfig, err = s.GardenClient.GetShootClientConfig(ctx, namespace, name, shoot, caSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Shoot kubeconfig: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown Gardener resource %q", resource)
	}

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}

	bytes, err := clientcmd.Write(rawConfig)
	if err != nil {
		return nil, err
	}

	config, err := kubeconfigutil.NewKubeconfig(bytes)
	if err != nil {
		return nil, err
	}

	contextNames, err := config.GetContextNames()
	if err != nil {
		return nil, err
	}

	for _, name := range contextNames {
		if strings.HasSuffix(name, "-internal") {
			if err := config.RemoveContext(name); err != nil {
				return nil, fmt.Errorf("unable to remove internal kubeconfig context: %v", err)
			}
			break
		}
	}

	// add meta information to kubeconfig (ignored by kubectl)
	// this allows the "controlplane" command to unambiguously determine the Shoot for this Gardener store
	if err := config.SetGardenerStoreMetaInformation(s.LandscapeIdentity, string(resource), gardenerProjectName, name); err != nil {
		return nil, err
	}

	return config.GetBytes()
}

func (s *GardenerStore) GetSearchPreview(path string, optionalTags map[string]string) (string, error) {
	// To improve UX, we return an error immediately and load the store in the background
	if !s.IsInitialized() {
		go func() {
			if err := s.InitializeGardenerStore(); err != nil {
				s.Logger.Debugf("failed to initialize Gardener store: %v", err)
			}
		}()
		return "", fmt.Errorf("gardener store is not initalized yet")
	}

	landscapeName := fmt.Sprintf("%s: %s", "Gardener landscape", s.LandscapeIdentity)
	if len(s.LandscapeName) > 0 {
		landscapeName = fmt.Sprintf("%s: %s", "Gardener landscape", s.LandscapeName)
	}

	if gardenerstore.GetGardenKubeconfigPath(s.LandscapeIdentity) == path {
		asciTree := gotree.New(fmt.Sprintf("%s (*)", landscapeName))
		return asciTree.Print(), nil
	}

	asciTree := gotree.New(landscapeName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, resource, name, namespace, projectName, err := gardenerstore.ParseIdentifier(path)
	if err != nil {
		return "", err
	}

	switch resource {
	case gardenerstore.GardenerResourceSeed:
		asciTree.Add(fmt.Sprintf("Seed: %s (*)", name))
		return asciTree.Print(), nil
	case gardenerstore.GardenerResourceShoot:
		asciTree.Add(fmt.Sprintf("Project: %s", projectName))

		shoot := &gardencorev1beta1.Shoot{}
		if err := s.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, shoot); err != nil {
			if apierrors.IsNotFound(err) {
				return "", fmt.Errorf("kubeconfig secret for %s (%s/%s) not found", resource, namespace, name)
			}
			return "", fmt.Errorf("failed to get kubeconfig secret for Shoot (%s/%s): %w", namespace, name, err)
		}

		asciSeed := gotree.New("Seed: not scheduled yet")
		if shoot.Status.SeedName != nil {
			asciSeed = gotree.New(fmt.Sprintf("Seed: %s", *shoot.Status.SeedName))
		}
		asciSeed.Add(fmt.Sprintf("Shoot: %s (*)", shoot.Name))
		asciTree.AddTree(asciSeed)
		return asciTree.Print(), err
	default:
		return "", fmt.Errorf("unknown Gardener resource %q", resource)
	}
}

func (s *GardenerStore) sendKubeconfigPaths(channel chan SearchResult, shoots []gardencorev1beta1.Shoot, managedSeeds []seedmanagementv1alpha1.ManagedSeed) {
	var landscapeName = s.LandscapeIdentity

	// first, send the garden context name configured in the switch config
	// the GetKubeconfigForPath() knows that this is a "special" path getting
	// the kubeconfig from the filesystem (set in SwitchConfig for the GardenerStore) instead of
	// from the Gardener API
	gardenKubeconfigPath := gardenerstore.GetGardenKubeconfigPath(s.LandscapeIdentity)
	channel <- SearchResult{
		KubeconfigPath: gardenKubeconfigPath,
		Error:          nil,
	}

	// all search result use the landscape name instead of the identity if configured
	// e.g dev-shoot-<shoot-name>
	if len(s.LandscapeName) > 0 {
		landscapeName = *s.Config.LandscapeName

		err := s.createGardenKubeconfigAlias(gardenKubeconfigPath)
		if err != nil {
			s.Logger.Warnf("failed to write alias %s for context name %s", fmt.Sprintf("%s-garden", landscapeName), fmt.Sprintf("%s-garden", s.LandscapeIdentity))
		}
	}

	s.CachePathToManagedSeed = make(map[string]seedmanagementv1alpha1.ManagedSeed, len(managedSeeds))
	s.CachePathToShoot = make(map[string]gardencorev1beta1.Shoot, len(shoots))

	shootNamesManagedSeed := make(map[string]struct{})
	for _, managedSeed := range managedSeeds {
		// shoots referenced by managed Seeds are assumed to be in the garden namespace
		shootNamesManagedSeed[fmt.Sprintf("garden:%s", managedSeed.Spec.Shoot.Name)] = struct{}{}
		// currently the name of the Seed resource of a manged Seed is ALWAYS the managed Seed's name
		kubeconfigPath := gardenerstore.GetSeedIdentifier(landscapeName, managedSeed.Name)

		// for memoization
		s.writeCachePathToManagedSeed(kubeconfigPath, managedSeed)
	}

	// loop over all Shoots/ShootedSeeds and construct and send their kubeconfig paths as search result
	for _, shoot := range shoots {
		seedName := shoot.Spec.SeedName
		if seedName == nil {
			// shoots that are not scheduled to Seed yet do not have a control plane
			continue
		}

		var projectName string
		if shoot.Namespace == "garden" {
			projectName = "garden"
		} else {
			// need to deduct the project name from the Shoot namespace garden-<projectname>
			split := strings.SplitN(shoot.Namespace, "garden-", 2)
			switch len(split) {
			case 2:
				projectName = split[1]
			default:
				continue
			}
		}

		var kubeconfigPath string

		// check if the shoot is a managed seed
		// check that the Shoot is not already added through the managed Seed to avoid duplicates
		if gardenerstore.IsManagedSeed(shoot) {
			// seed resource of a Shooted seed should have the same name as the Seed
			kubeconfigPath = gardenerstore.GetSeedIdentifier(landscapeName, shoot.Name)
		} else {
			kubeconfigPath = gardenerstore.GetShootIdentifier(landscapeName, projectName, shoot.Name)
		}

		// for memoization
		s.writeCachePathToShoot(kubeconfigPath, shoot)

		_, isAlreadyReferencedByManagedSeed := shootNamesManagedSeed[fmt.Sprintf("%s:%s", shoot.Namespace, shoot.Name)]
		if isAlreadyReferencedByManagedSeed {
			continue
		}

		channel <- SearchResult{
			KubeconfigPath: kubeconfigPath,
			Error:          nil,
		}
	}

	// the reason why the paths for managed seeds are populated here in the end (even though they are available before),
	// is so that the corresponding Shoot resource for the ManagedSeed is already available the cache s.CachePathToShoot[]
	// when populating the path. This avoids cache misses.
	s.PathToManagedSeedLock.RLock()
	for pathForSeed := range s.CachePathToManagedSeed {
		channel <- SearchResult{
			KubeconfigPath: pathForSeed,
			Error:          nil,
		}
	}
	s.PathToManagedSeedLock.RUnlock()
}

func (s *GardenerStore) createGardenKubeconfigAlias(gardenKubeconfigPath string) error {
	bytes, err := s.GetKubeconfigForPath(gardenKubeconfigPath, nil)
	if err != nil {
		return err
	}

	// get context name from the virtual garden kubeconfig
	_, contexts, err := util.GetContextsNamesFromKubeconfig(bytes, s.GetContextPrefix(gardenKubeconfigPath))
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig context names for path %q: %v", gardenKubeconfigPath, err)
	}

	if len(contexts) == 0 {
		return fmt.Errorf("no context names found")
	}

	// create an additional alias for the garden context name
	a, err := state.GetDefaultAlias(s.StateDirectory)
	if err != nil {
		return err
	}

	gardenContextName := contexts[0]
	// alias sap-landscape-dev-garden/virtual-garden with sap-landscape-dev-garden
	// in order to get to the garden API by just 'switch sap-landscape-dev-garden'
	// which can be extracted from the cluster-identity cm in the Shoot
	if _, err := a.WriteAlias(gardenKubeconfigPath, gardenContextName); err != nil {
		return err
	}
	return nil
}

func (s *GardenerStore) VerifyKubeconfigPaths() error {
	// NOOP as we do not allow any paths to be configured for the Gardener store
	// searches through all namespaces
	return nil
}

func (s *GardenerStore) writeCachePathToShoot(key string, value gardencorev1beta1.Shoot) {
	s.PathToShootLock.Lock()
	defer s.PathToShootLock.Unlock()
	s.CachePathToShoot[key] = value
}

func (s *GardenerStore) readFromCachePathToShoot(key string) (gardencorev1beta1.Shoot, bool) {
	s.PathToShootLock.RLock()
	defer s.PathToShootLock.RUnlock()
	shoot, ok := s.CachePathToShoot[key]
	return shoot, ok
}

func (s *GardenerStore) writeCachePathToManagedSeed(key string, value seedmanagementv1alpha1.ManagedSeed) {
	s.PathToManagedSeedLock.Lock()
	defer s.PathToManagedSeedLock.Unlock()
	s.CachePathToManagedSeed[key] = value
}

func (s *GardenerStore) readFromCachePathToManagedSeed(key string) (seedmanagementv1alpha1.ManagedSeed, bool) {
	s.PathToManagedSeedLock.RLock()
	defer s.PathToManagedSeedLock.RUnlock()
	managedSeed, ok := s.CachePathToManagedSeed[key]
	return managedSeed, ok
}

func (s *GardenerStore) writeCacheCaSecretNameToSecretLock(key string, value corev1.Secret) {
	s.CaSecretNameToSecretLock.Lock()
	defer s.CaSecretNameToSecretLock.Unlock()
	s.CacheCaSecretNameToSecret[key] = value
}

func (s *GardenerStore) readFromCacheCaSecretNameToSecretLock(key string) (corev1.Secret, bool) {
	s.CaSecretNameToSecretLock.RLock()
	defer s.CaSecretNameToSecretLock.RUnlock()
	secret, ok := s.CacheCaSecretNameToSecret[key]
	return secret, ok
}
