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
	"io/ioutil"
	"strings"
	"time"

	"github.com/disiqueira/gotree"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	seedmanagementv1alpha1 "github.com/gardener/gardener/pkg/apis/seedmanagement/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gardenerstore "github.com/danielfoehrkn/kubeswitch/pkg/store/gardener"
	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/state"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/danielfoehrkn/kubeswitch/types"
)

const (
	CM_NAME_CLUSTER_IDENTITY   = "cluster-identity"
	KEY_CLUSTER_IDENTITY       = CM_NAME_CLUSTER_IDENTITY
	ALL_NAMESPACES_DENOMINATOR = "/"
)

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
		Logger:          logrus.New().WithField("store", types.StoreKindGardener),
		KubeconfigStore: store,
		Config:          config,
		LandscapeName:   landscapeName,
		StateDirectory:  stateDir,
	}, nil
}

// InitializeGardenerStore initializes the store using the provided Gardener kubeconfig
// decoupled from the NewGardenerStore() to be called when starting the search to reduce
// time when the CLI can start showing the fuzzy search
func (s *GardenerStore) InitializeGardenerStore() error {
	gardenClient, err := gardenerstore.GetGardenClient(s.Config)
	if err != nil {
		return err
	}
	s.Client = gardenClient

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cm := &corev1.ConfigMap{}
	if err := gardenClient.Get(ctx, client.ObjectKey{Name: CM_NAME_CLUSTER_IDENTITY, Namespace: metav1.NamespaceSystem}, cm); err != nil {
		return fmt.Errorf("unable to get gardener landscape identity from config map %s/%s: %w", metav1.NamespaceSystem, CM_NAME_CLUSTER_IDENTITY, err)
	}

	identity, ok := cm.Data[KEY_CLUSTER_IDENTITY]
	if !ok {
		return fmt.Errorf("unable to get gardener landscape identity from config map %s/%s: data key %q not found", metav1.NamespaceSystem, CM_NAME_CLUSTER_IDENTITY, KEY_CLUSTER_IDENTITY)
	}
	s.LandscapeIdentity = identity
	return nil
}

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

	for _, p := range s.KubeconfigStore.Paths {
		if p == ALL_NAMESPACES_DENOMINATOR {
			continue
		}

		ns := corev1.Namespace{}
		if err := s.Client.Get(ctx, client.ObjectKey{Name: p}, &ns); err != nil {
			if apierrors.IsNotFound(err) {
				err = fmt.Errorf("namespace %q does not exist. Please check your path configuration : %w", p, err)
			}
			channel <- SearchResult{
				Error: err,
			}
			return
		}
	}

	kubeconfigSecretSelector := labels.SelectorFromSet(labels.Set{"gardener.cloud/role": "kubeconfig"})
	secretResult, err := s.getListForPaths(ctx, &corev1.SecretList{}, &kubeconfigSecretSelector)
	if err != nil {
		channel <- SearchResult{
			Error: fmt.Errorf("failed to call Gardener API: %v", err),
		}
		return
	}

	shootNameToSecret := gardenerstore.GetSecretNamespaceNameToSecret(s.Logger, secretResult)
	// save to use later in GetKubeconfigForPath()
	s.SecretNamespaceNameToSecret = shootNameToSecret
	s.Logger.Debugf("Found %d kubeconfigs", len(shootNameToSecret))

	shootResult, err := s.getListForPaths(ctx, &gardencorev1beta1.ShootList{}, nil)
	if err != nil {
		channel <- SearchResult{
			Error: fmt.Errorf("failed to call Gardener API: %v", err),
		}
		return
	}

	managedSeeds := &seedmanagementv1alpha1.ManagedSeedList{}
	if err := s.Client.List(ctx, managedSeeds, &client.ListOptions{}); err != nil {
		// do not return here as many older Gardener installations do not have the
		// resource group for managed seeds yet
		s.Logger.Debugf("failed to list managed seeds: %v", err)
	}

	s.sendKubeconfigPaths(channel, shootResult, managedSeeds)
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

// getListForPaths takes a context an object list (e.g secretsList)
// returns an object list for each configured paths (namespaces)
func (s *GardenerStore) getListForPaths(ctx context.Context, list client.ObjectList, selector *labels.Selector) ([]client.ObjectList, error) {
	// specifying no path equals to all namespaces being searched
	if len(s.KubeconfigStore.Paths) == 0 {
		s.KubeconfigStore.Paths = []string{ALL_NAMESPACES_DENOMINATOR}
	}

	resultList := make([]client.ObjectList, len(s.KubeconfigStore.Paths))
	for i, path := range s.KubeconfigStore.Paths {
		var (
			l           = list
			listOptions = client.ListOptions{}
		)

		if path != ALL_NAMESPACES_DENOMINATOR {
			listOptions.Namespace = path
		}

		if selector != nil {
			listOptions.LabelSelector = *selector
		}

		if err := s.Client.List(ctx, l, &listOptions); err != nil {
			return nil, err
		}
		copy := l.DeepCopyObject()
		resultList[i] = copy.(client.ObjectList)

		if path == ALL_NAMESPACES_DENOMINATOR {
			break
		}
	}
	return resultList, nil
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

func (s *GardenerStore) GetKubeconfigForPath(path string) ([]byte, error) {
	if !s.IsInitialized() {
		if err := s.InitializeGardenerStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize Gardener store: %w", err)
		}
	}

	if gardenerstore.GetGardenKubeconfigPath(s.LandscapeIdentity) == path {
		if s.Config == nil || len(s.Config.GardenerAPIKubeconfigPath) == 0 {
			return nil, fmt.Errorf("cannot get garden kubeconfig. Field 'gardenerAPIKubeconfigPath' is not configured in the Gardener store configuration in the SwitchConfig file")
		}
		return ioutil.ReadFile(s.Config.GardenerAPIKubeconfigPath)
	}

	landscape, resource, name, namespace, _, err := gardenerstore.ParseIdentifier(path)
	if err != nil {
		return nil, err
	}

	if landscape != s.LandscapeName && landscape != s.LandscapeIdentity {
		return nil, fmt.Errorf("unknown Gardener landscape %q", landscape)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kubeconfigSecret := corev1.Secret{}

	switch resource {
	case gardenerstore.GardenerResourceSeed:
		// For Seeds:
		// at the moment, managed seeds can only refer to Shoots in the Garden namespace
		// we do not support external Seeds (that have a kubeconfig set)

		namespace = "garden"
		fallthrough
	case gardenerstore.GardenerResourceShoot:
		var ok bool

		s.Logger.Debugf("getting kubeconfig for %s (%s/%s)", resource, namespace, name)

		kubeconfigSecret, ok = s.SecretNamespaceNameToSecret[gardenerstore.GetSecretIdentifier(namespace, name)]
		if !ok {
			if err := s.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: fmt.Sprintf("%s.kubeconfig", name)}, &kubeconfigSecret); err != nil {
				if apierrors.IsNotFound(err) {
					return nil, fmt.Errorf("kubeconfig secret for %s (%s/%s) not found", resource, namespace, name)
				}
				return nil, fmt.Errorf("failed to get kubeconfig secret for Shoot (%s/%s): %w", namespace, name, err)
			}
		}
	default:
		return nil, fmt.Errorf("unknown Gardener resource %q", resource)
	}

	value, found := kubeconfigSecret.Data[secrets.DataKeyKubeconfig]
	if !found {
		return nil, fmt.Errorf("kubeconfig secret for Shoot (%s/%s) does not contain a kubeconfig", namespace, name)
	}

	return value, nil
}

func (s *GardenerStore) GetSearchPreview(path string) (string, error) {
	if !s.IsInitialized() {
		if err := s.InitializeGardenerStore(); err != nil {
			return "", fmt.Errorf("failed to initialize Gardener store: %w", err)
		}
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

func (s *GardenerStore) sendKubeconfigPaths(channel chan SearchResult, shoots []client.ObjectList, managedSeeds *seedmanagementv1alpha1.ManagedSeedList) {
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

	shootNamesManagedSeed := make(map[string]struct{})
	for _, managedSeed := range managedSeeds.Items {
		// shoots referenced by managed Seeds are assumed to be in the garden namespace
		shootNamesManagedSeed[fmt.Sprintf("garden:%s", managedSeed.Spec.Shoot.Name)] = struct{}{}
		// currently the name of the Seed resource of a manged Seed is ALWAYS the managed resource name
		kubeconfigPath := gardenerstore.GetSeedIdentifier(landscapeName, managedSeed.Name)
		channel <- SearchResult{
			KubeconfigPath: kubeconfigPath,
			Error:          nil,
		}
	}

	// loop over all Shoots/ShootedSeeds and construct and send their kubeconfig paths as search result
	for _, objectList := range shoots {
		shootList := objectList.(*gardencorev1beta1.ShootList)
		for _, shoot := range shootList.Items {
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

			_, isAlreadyReferencedByManagedSeed := shootNamesManagedSeed[fmt.Sprintf("%s:%s", shoot.Namespace, shoot.Name)]
			if isAlreadyReferencedByManagedSeed {
				continue
			}
			// check for shooted seed annotation
			// check that the Shoot is not already added through the managed Seed to avoid duplicates
			if gardenerstore.IsShootedSeed(shoot) {
				// seed resource of a Shooted seed should have the same name as the Seed
				kubeconfigPath = gardenerstore.GetSeedIdentifier(landscapeName, shoot.Name)
			} else {
				kubeconfigPath = gardenerstore.GetShootIdentifier(landscapeName, projectName, shoot.Name)
			}

			channel <- SearchResult{
				KubeconfigPath: kubeconfigPath,
				Error:          nil,
			}
		}
	}
}

func (s *GardenerStore) createGardenKubeconfigAlias(gardenKubeconfigPath string) error {
	bytes, err := s.GetKubeconfigForPath(gardenKubeconfigPath)
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
