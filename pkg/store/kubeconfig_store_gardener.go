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
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/danielfoehrkn/kubeswitch/pkg/subcommands/alias/state"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/danielfoehrkn/kubeswitch/types"
)

const (
	cmNameClusterIdentity = "cluster-identity"
	keyClusterIdentity    = cmNameClusterIdentity
)

// NewGardenerStore creates a new Gardener store
func NewGardenerStore(store types.KubeconfigStore, stateDir string) (*GardenerStore, error) {
	config, err := getStoreConfig(store)
	if err != nil {
		return nil, err
	}

	gardenClient, err := getGardenClient(config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cm := &corev1.ConfigMap{}
	if err := gardenClient.Get(ctx, client.ObjectKey{Name: cmNameClusterIdentity, Namespace: metav1.NamespaceSystem}, cm); err != nil {
		return nil, fmt.Errorf("unable to get gardener landscape identity from config map %s/%s: %w", metav1.NamespaceSystem, cmNameClusterIdentity, err)
	}

	identity, ok := cm.Data[keyClusterIdentity]
	if !ok {
		return nil, fmt.Errorf("unable to get gardener landscape identity from config map %s/%s: data key %q not found", metav1.NamespaceSystem, cmNameClusterIdentity, keyClusterIdentity)
	}

	var landscapeName string
	if config != nil && config.LandscapeName != nil {
		landscapeName = *config.LandscapeName
	}

	return &GardenerStore{
		Logger:            logrus.New().WithField("store", types.StoreKindGardener),
		Client:            gardenClient,
		KubeconfigStore:   store,
		Config:            config,
		LandscapeIdentity: identity,
		LandscapeName:     landscapeName,
		StateDirectory:    stateDir,
	}, nil
}

// ValidateGardenerStoreConfiguration validates the store configuration for Gardener
// returns the optional landscape name as well as the error list
func ValidateGardenerStoreConfiguration(path *field.Path, store types.KubeconfigStore) (*string, field.ErrorList) {
	var errors = field.ErrorList{}

	// always find the kubeconfigs of all Shoot on the landscape
	// in the future it could be restricted via paths to only certain namespaces
	if len(store.Paths) > 0 {
		errors = append(errors, field.Forbidden(path.Child("paths"), "specifying a path for the Gardener store is currently not supported"))
	}

	configPath := path.Child("config")
	if store.Config == nil {
		errors = append(errors, field.Required(configPath, "Missing configuration in the SwitchConfig file for the Gardener store"))
		return nil, errors
	}

	config, err := getStoreConfig(store)
	if err != nil {
		errors = append(errors, field.Invalid(configPath, store.Config, err.Error()))
		return nil, errors
	}

	if len(config.GardenerAPIKubeconfigPath) == 0 {
		errors = append(errors, field.Invalid(configPath.Child("gardenerAPIKubeconfigPath"), config.GardenerAPIKubeconfigPath, "The kubeconfig to the Gardener API server must be set"))
	}

	if config.LandscapeName != nil && len(*config.LandscapeName) == 0 {
		errors = append(errors, field.Invalid(configPath.Child("landscapeName"), *config.LandscapeName, "The optional Gardener landscape name must not be empty"))
	}

	return config.LandscapeName, errors
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

func (s *GardenerStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *GardenerStore) GetKubeconfigForPath(path string) ([]byte, error) {
	if getGardenKubeconfigPath(s.LandscapeIdentity) == path {
		if s.Config == nil || len(s.Config.GardenerAPIKubeconfigPath) == 0 {
			return nil, fmt.Errorf("cannot get garden kubeconfig. Field 'gardenerAPIKubeconfigPath' is not configured in the Gardener store configuration in the SwitchConfig file")
		}
		return ioutil.ReadFile(s.Config.GardenerAPIKubeconfigPath)
	}

	landscape, resource, name, namespace, err := parseIdentifier(path)
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
	case GardenerResourceSeed:
		s.Logger.Debugf("getting kubeconfig for seed %q", name)
		// at the moment, managed seeds can only refer to Shoots in the Garden namespace
		// hence, get the kubeconfig secret from there
		// we do not support managed seeds or external Seeds (that have a kubeconfig set) yet
		if err := s.Client.Get(ctx, client.ObjectKey{Namespace: "garden", Name: fmt.Sprintf("%s.kubeconfig", name)}, &kubeconfigSecret); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("kubeconfig secret for Seed %q not found", name)
			}
			return nil, fmt.Errorf("failed to get kubeconfig secret for Seed %q: %w", name, err)
		}
	case GardenerResourceShoot:
		var ok bool

		s.Logger.Debugf("getting kubeconfig for Shoot (%s/%s)", namespace, name)
		kubeconfigSecret, ok = s.ShootNameToKubeconfigSecret[getSecretIdentifier(namespace, name)]
		if !ok {
			if err := s.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: fmt.Sprintf("%s.kubeconfig", name)}, &kubeconfigSecret); err != nil {
				if apierrors.IsNotFound(err) {
					return nil, fmt.Errorf("kubeconfig secret for Shoot (%s/%s) not found", namespace, name)
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

func (s *GardenerStore) StartSearch(channel chan SearchResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: parallelize retrieval of secrets, shoot and projects via goroutines
	selector := labels.SelectorFromSet(labels.Set{"gardener.cloud/role": "kubeconfig"})
	secretList := &corev1.SecretList{}
	if err := s.Client.List(ctx, secretList, &client.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		channel <- SearchResult{
			Error: fmt.Errorf("failed to retrieve secrets from Gardener API: %v", err),
		}
		return
	}

	shootNameToSecret := getShootNameToSecret(s.Logger, secretList)
	// save to use later in GetKubeconfigForPath()
	s.ShootNameToKubeconfigSecret = shootNameToSecret
	s.Logger.Debugf("Found %d kubeconfigs", len(shootNameToSecret))

	shoots := &gardencorev1beta1.ShootList{}
	if err := s.Client.List(ctx, shoots, &client.ListOptions{Namespace: "garden"}); err != nil {
		// if err := s.Client.List(ctx, shoots, &client.ListOptions{}); err != nil {
		channel <- SearchResult{
			Error: fmt.Errorf("failed to list Shoots: %w", err),
		}
		return
	}
	s.Logger.Debugf("Found %d shoots", len(shoots.Items))

	projects := &gardencorev1beta1.ProjectList{}
	if err := s.Client.List(ctx, projects, &client.ListOptions{}); err != nil {
		channel <- SearchResult{
			Error: fmt.Errorf("failed to list projects: %w", err),
		}
		return
	}
	s.Logger.Debugf("Found %d projects", len(projects.Items))

	s.sendKubeconfigPaths(channel, shoots, buildNamespaceToProjectMap(projects))
}

func getGardenClient(config *types.StoreConfigGardener) (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(gardencorev1beta1.AddToScheme(scheme))

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.GardenerAPIKubeconfigPath},
		&clientcmd.ConfigOverrides{})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to create rest config: %v", err))
	}

	k8sclient, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to create garden client: %v", err))
	}
	return k8sclient, nil
}

// getStoreConfig unmarshalls to the Gardener store config from the configuration
func getStoreConfig(store types.KubeconfigStore) (*types.StoreConfigGardener, error) {
	if store.Config == nil {
		return nil, fmt.Errorf("providing a configuration for the Gardener store is required. Please configure your SwitchConfig file properly")
	}

	storeConfig := &types.StoreConfigGardener{}
	buf, err := yaml.Marshal(store.Config)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(buf, storeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config for the Gardener kubeconfig store: %w", err)
	}
	return storeConfig, nil
}

func (s *GardenerStore) sendKubeconfigPaths(channel chan SearchResult, shoots *gardencorev1beta1.ShootList, namespaceToProjectName map[string]string) {
	var (
		shootIdentifiers       = sets.NewString()
		shootedSeedIdentifiers = sets.NewString()
		landscapeName          = s.LandscapeIdentity
	)

	// first, send the garden context name configured in the switch config
	// the GetKubeconfigForPath() knows that this is a "special" path getting
	// the kubeconfig from the filesystem (set in SwitchConfig for the GardenerStore) instead of
	// from the Gardener API
	gardenKubeconfigPath := getGardenKubeconfigPath(s.LandscapeIdentity)
	channel <- SearchResult{
		KubeconfigPath: gardenKubeconfigPath,
		Error:          nil,
	}

	// all search result use the landscape name instead of the identity if configured
	// e.g dev-shoot-<shoot-name>
	if len(s.LandscapeName) > 0 {
		landscapeName = *s.Config.LandscapeName

		err := s.createGardenKubeconfigAlias(landscapeName, gardenKubeconfigPath)
		if err != nil {
			s.Logger.Warnf("failed to write alias %s for context name %s", fmt.Sprintf("%s-garden", landscapeName), fmt.Sprintf("%s-garden", s.LandscapeIdentity))
		}
	}

	// loop over all Shoots/ShootedSeeds and construct and send their kubeconfig paths as search result
	for _, shoot := range shoots.Items {
		seedName := shoot.Spec.SeedName
		if seedName == nil {
			continue
		}

		projectName := namespaceToProjectName[shoot.Namespace]
		if len(projectName) == 0 {
			s.Logger.Warnf("Could not find project for Shoot (%s/%s). Skipping.", shoot.Namespace, shoot.Name)
			continue
		}

		var kubeconfigPath string

		// TODO: include managed seed
		// check for shooted seed
		if isShootedSeed(shoot) {
			kubeconfigPath = getSeedIdentifier(landscapeName, shoot.Name)
			shootedSeedIdentifiers.Insert(kubeconfigPath)
		} else {
			kubeconfigPath = getShootIdentifier(landscapeName, projectName, shoot.Name)
			shootIdentifiers.Insert(kubeconfigPath)
		}

		channel <- SearchResult{
			KubeconfigPath: kubeconfigPath,
			Error:          nil,
		}
	}
}

func getGardenKubeconfigPath(landscapeIdentity string) string {
	return fmt.Sprintf("%s-garden", landscapeIdentity)
}

func (s *GardenerStore) createGardenKubeconfigAlias(landscapeName string, gardenKubeconfigPath string) error {
	bytes, err := s.GetKubeconfigForPath(gardenKubeconfigPath)
	if err != nil {
		return err
	}

	// get context name from the virtual garden kubeconfig
	_, contexts, err := util.GetContextsForKubeconfigPath(bytes, types.StoreKindGardener, gardenKubeconfigPath)
	if err != nil {
		return err
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
	if err := a.WriteAlias(gardenKubeconfigPath, gardenContextName); err != nil {
		return err
	}
	return nil
}

func (s *GardenerStore) VerifyKubeconfigPaths() error {
	if len(s.KubeconfigStore.Paths) == 1 && s.KubeconfigStore.Paths[0] == "/" {
		return nil
	}

	// we do not list all namespaces here as Gardener can have several thousands of namespaces
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, path := range s.KubeconfigStore.Paths {
		ns := &corev1.Namespace{}
		if err := s.Client.Get(ctx, client.ObjectKey{Name: path}, ns); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("failed to retrieve namespaces from Gardener API: %w", err)
			}
			return fmt.Errorf("configured namespace %q does not exist in the Gardener API", path)
		}
	}
	return nil
}

func getShootNameToSecret(log *logrus.Entry, secretList *corev1.SecretList) map[string]corev1.Secret {
	shootNameToSecret := make(map[string]corev1.Secret, len(secretList.Items))
	for _, secret := range secretList.Items {
		if _, exists := secret.Data[secrets.DataKeyKubeconfig]; !exists {
			log.Warnf("Secret %s/%s does not contain a kubeconfig. Skipping.", secret.Namespace, secret.Name)
			continue
		}

		var shootName string
		if len(secret.ObjectMeta.OwnerReferences) == 0 || secret.ObjectMeta.OwnerReferences[0].Kind != "Shoot" {
			if !strings.Contains(secret.Namespace, ".kubeconfig") {
				log.Warnf("Secret %s/%s could not be associated with any Shoot. Skipping.", secret.Namespace, secret.Name)
				continue
			}
			shootName = strings.Split(secret.Namespace, ".kubeconfig")[0]
		} else {
			shootName = secret.ObjectMeta.OwnerReferences[0].Name
		}
		shootNameToSecret[getSecretIdentifier(secret.Namespace, shootName)] = secret
	}
	return shootNameToSecret
}

func buildNamespaceToProjectMap(projects *gardencorev1beta1.ProjectList) map[string]string {
	namespaceToProjectName := make(map[string]string, len(projects.Items))
	for _, project := range projects.Items {
		namespace := project.Spec.Namespace
		if namespace == nil {
			continue
		}
		if _, ok := namespaceToProjectName[*namespace]; !ok {
			namespaceToProjectName[*namespace] = project.Name
		}
	}
	return namespaceToProjectName
}

// isShootedSeed determines if this Shoot is a Shooted seed based on an annotation
// TODO: also support managed Seeds
func isShootedSeed(shoot gardencorev1beta1.Shoot) bool {
	if shoot.Namespace == v1beta1constants.GardenNamespace && shoot.Annotations != nil {
		_, ok := v1beta1constants.GetShootUseAsSeedAnnotation(shoot.Annotations)
		return ok
	}
	return false
}

type GardenerResource string

const (
	GardenerResourceShoot GardenerResource = "Shoot"
	GardenerResourceSeed  GardenerResource = "Seed"
)

// <namespace>-<name>
func getSecretIdentifier(namespace string, shootName string) string {
	return fmt.Sprintf("%s/%s", namespace, shootName)
}

// <landscape>--seed--<seed-name>
func getSeedIdentifier(landscape, shoot string) string {
	return fmt.Sprintf("%s--seed--%s", landscape, shoot)
}

// <landscape>--shoot--<project-name>--<shoot-name>
func getShootIdentifier(landscape, project, shoot string) string {
	return fmt.Sprintf("%s--shoot--%s--%s", landscape, project, shoot)
}

// parseIdentifier takes a kubeconfig identifier and
// returns the
// 1) the landscape identity or name
// 1) type of the Gardener resource (shoot/seed)
// 2) name of the resource
// 3) optionally the namespace
func parseIdentifier(path string) (string, GardenerResource, string, string, error) {
	split := strings.Split(path, "--")
	switch len(split) {
	case 4:
		if !strings.Contains(path, "shoot") {
			return "", "", "", "", fmt.Errorf("cannot parse kubeconfig path %q", path)
		}
		return split[0], GardenerResourceShoot, split[3], fmt.Sprintf("garden-%s", split[2]), nil
	case 3:
		if !strings.Contains(path, "seed") {
			return "", "", "", "", fmt.Errorf("cannot parse kubeconfig path: %q", path)
		}
		return split[0], GardenerResourceSeed, split[2], "", nil

	default:
		return "", "", "", "", fmt.Errorf("cannot parse kubeconfig path: %q", path)
	}
}
