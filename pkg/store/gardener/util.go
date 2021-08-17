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

package gardener

import (
	"fmt"
	"log"
	"os"
	"strings"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	seedmanagementv1alpha1 "github.com/gardener/gardener/pkg/apis/seedmanagement/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/danielfoehrkn/kubeswitch/types"
)

type GardenerResource string

const (
	GardenerResourceShoot GardenerResource = "Shoot"
	GardenerResourceSeed  GardenerResource = "Seed"
)

// GetStoreConfig unmarshalls to the Gardener store config from the configuration
func GetStoreConfig(store types.KubeconfigStore) (*types.StoreConfigGardener, error) {
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

func GetGardenClient(config *types.StoreConfigGardener) (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(gardencorev1beta1.AddToScheme(scheme))
	utilruntime.Must(seedmanagementv1alpha1.AddToScheme(scheme))

	gardenerAPIKubeconfigPath := strings.ReplaceAll(config.GardenerAPIKubeconfigPath, "~", "$HOME")
	gardenerAPIKubeconfigPath = os.ExpandEnv(gardenerAPIKubeconfigPath)

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: gardenerAPIKubeconfigPath},
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

// GetGardenKubeconfigPath gets the kubeconfig path for the kubeconfig that is configured
// in the SwitchConfig and points to the Gardener API
func GetGardenKubeconfigPath(landscapeIdentity string) string {
	return fmt.Sprintf("%s-garden", landscapeIdentity)
}

// <namespace>-<name>
func GetSecretIdentifier(namespace string, shootName string) string {
	return fmt.Sprintf("%s/%s", namespace, shootName)
}

// <landscape>--seed--<seed-name>
func GetSeedIdentifier(landscape, seedName string) string {
	return fmt.Sprintf("%s--seed--%s", landscape, seedName)
}

// <landscape>--shoot--<project-name>--<shoot-name>
func GetShootIdentifier(landscape, project, shoot string) string {
	return fmt.Sprintf("%s--shoot--%s--%s", landscape, project, shoot)
}

// ParseIdentifier takes a kubeconfig identifier and
// returns the
// 1) the landscape identity or name
// 2) type of the Gardener resource (shoot/seed)
// 3) name of the resource
// 4) optionally the namespace
// 5) optionally the project name
func ParseIdentifier(path string) (string, GardenerResource, string, string, string, error) {
	split := strings.Split(path, "--")
	switch len(split) {
	case 4:
		if !strings.Contains(path, "shoot") {
			return "", "", "", "", "", fmt.Errorf("cannot parse kubeconfig path %q", path)
		}

		// TODO something is off here, I should be getting the project name only right??
		projectName := "garden"
		namespace := "garden"
		if split[2] != "garden" { // e.g d060239
			namespace = fmt.Sprintf("garden-%s", split[2])
			projectName = split[2]
		}

		return split[0], GardenerResourceShoot, split[3], namespace, projectName, nil
	case 3:
		if !strings.Contains(path, "seed") {
			return "", "", "", "", "", fmt.Errorf("cannot parse kubeconfig path: %q", path)
		}
		return split[0], GardenerResourceSeed, split[2], "", "", nil

	default:
		return "", "", "", "", "", fmt.Errorf("cannot parse kubeconfig path: %q", path)
	}
}

// takes a slice of client.ObjectList and casts each client.ObjectList to corev1.SecretList
// returns a map[string]corev1.Secret where the key is the secret <namespace>-<name> and the value the secret resource
func GetSecretNamespaceNameToSecret(log *logrus.Entry, secretSlice []client.ObjectList) map[string]corev1.Secret {
	shootNameToSecret := make(map[string]corev1.Secret)
	// we have a list of lists
	for _, objectList := range secretSlice {
		list := objectList.(*corev1.SecretList)
		for _, secret := range list.Items {
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
			shootNameToSecret[GetSecretIdentifier(secret.Namespace, shootName)] = secret
		}
	}
	return shootNameToSecret
}

// isShootedSeed determines if this Shoot is a Shooted seed based on an annotation
func IsShootedSeed(shoot gardencorev1beta1.Shoot) bool {
	if shoot.Namespace == v1beta1constants.GardenNamespace && shoot.Annotations != nil {
		return shoot.Annotations[v1beta1constants.AnnotationShootUseAsSeed] != ""
	}
	return false
}
