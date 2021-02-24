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

package pkg

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/danielfoehrkn/kubectlSwitch/hooks/gardener-landscape-sync/pkg/hookstore"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/index"
)

func RunHook(log *logrus.Entry, kubeconfigStore hookstore.KubeconfigStore, clean bool, shootKubeconfigName, gardenKubeconfigPath, exportPath, landscapeName, stateDir string) error {
	if len(gardenKubeconfigPath) == 0 {
		return fmt.Errorf("must set the path to the kubeconfig of the Garden cluster")
	}
	if len(exportPath) == 0 {
		return fmt.Errorf("must set the export directory")
	}
	if len(landscapeName) == 0 {
		return fmt.Errorf("must provide a landscape name")
	}

	landscapeDirectory := fmt.Sprintf("%s/%s", exportPath, landscapeName)

	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(gardencorev1beta1.AddToScheme(scheme))

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: gardenKubeconfigPath},
		&clientcmd.ConfigOverrides{})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("unable to create rest config: %v", err))
	}

	k8sclient, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("unable to create kubernetes client: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	selector := labels.SelectorFromSet(labels.Set{"gardener.cloud/role": "kubeconfig"})
	secretList := &corev1.SecretList{}
	if err := k8sclient.List(ctx, secretList, &client.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		return fmt.Errorf("failed to list secret objects: %v", err)
	}

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

	log.Infof("Found %d kubeconfigs", len(shootNameToSecret))

	shoots := &gardencorev1beta1.ShootList{}
	if err := k8sclient.List(ctx, shoots, &client.ListOptions{}); err != nil {
		return fmt.Errorf("Failed to  list secret objects: %v", err)
	}
	log.Infof("Found %d shoots", len(shoots.Items))

	projects := &gardencorev1beta1.ProjectList{}
	if err := k8sclient.List(ctx, projects, &client.ListOptions{}); err != nil {
		return fmt.Errorf("failed to  list secret objects: %v", err)
	}
	log.Infof("Found %d projects", len(projects.Items))

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

	shootIdentifiers := sets.NewString()
	shootedSeedIdentifiers := sets.NewString()

	seedNames := map[string]struct{}{}
	shootedSeedNames := map[string]struct{}{}

	searchIndex, err := index.New(log, kubeconfigStore.GetKind(), stateDir)
	if err != nil {
		return err
	}

	oldShootIdentifiers, oldSeedIdentifiers, err := GetPreviousIdentifiers(searchIndex, landscapeName)
	if err != nil {
		log.Warnf("Failed to get existing kubeconfigs from the filesystem under path %q: %v", exportPath, err)
	}

	if clean {
		if err := kubeconfigStore.CleanExistingKubeconfigs(log, landscapeDirectory); err != nil {
			log.Warnf("Failed to clean existing kubeconfigs from the filesystem under path %q: %v", exportPath, err)
		}
	}

	// create root directory
	if err := kubeconfigStore.CreateLandscapeDirectory(landscapeDirectory); err != nil {
		return fmt.Errorf("failed to create landscape directory %q: %v", landscapeDirectory, err)
	}

	for _, shoot := range shoots.Items {
		seedName := shoot.Spec.SeedName
		if seedName == nil {
			continue
		}

		if _, ok := seedNames[*seedName]; !ok {
			seedNames[*seedName] = struct{}{}
		}
		projectName := namespaceToProjectName[shoot.Namespace]
		if len(projectName) == 0 {
			log.Warnf("Could not find project for Shoot (%s/%s). Skipping.", shoot.Namespace, shoot.Name)
			continue
		}

		var (
			identifier          string
			kubeconfigDirectory string
		)
		// check for shooted seed
		isShootedSeed := isShootedSeed(shoot)
		if isShootedSeed {
			identifier = getSeedIdentifier(landscapeName, shoot.Name)
			kubeconfigDirectory = getSeedKubeconfigDirectory(exportPath, landscapeName, identifier)
			if _, ok := shootedSeedNames[shoot.Name]; !ok {
				shootedSeedNames[shoot.Name] = struct{}{}
			}
		} else {
			identifier = getShootIdentifier(landscapeName, projectName, shoot.Name)
			kubeconfigDirectory = getShootKubeconfigDirectory(exportPath, landscapeName, *seedName, identifier)
		}

		var (
			secret      = corev1.Secret{}
			secretFound bool
		)
		secret, secretFound = shootNameToSecret[getSecretIdentifier(shoot.Namespace, shoot.Name)]
		if !secretFound {
			if err := k8sclient.Get(ctx, client.ObjectKey{Namespace: shoot.Namespace, Name: fmt.Sprintf("%s.kubeconfig", shoot.Name)}, &secret); err != nil {
				if apierrors.IsNotFound(err) {
					log.Warnf("Could not find kubeconfig secret for Shoot (%s/%s). Skipping.", shoot.Namespace, shoot.Name)
					continue
				}
				log.Errorf("Failed to get kubeconfig secret for Shoot (%s/%s). Skipping.", shoot.Namespace, shoot.Name)
				continue
			}
		}

		if err := kubeconfigStore.WriteKubeconfigFile(kubeconfigDirectory, shootKubeconfigName, secret); err != nil {
			return fmt.Errorf("unable to write kubeconfig to path: %s: %v", kubeconfigDirectory, err)
		}
		if isShootedSeed {
			shootedSeedIdentifiers.Insert(identifier)
		} else {
			shootIdentifiers.Insert(identifier)
		}

		if shootIdentifiers.Len()%30 == 0 {
			log.Infof("Wrote %d shoot kubeconfigs.", len(shootIdentifiers))
		}
	}

	// delete search index for the corresponding store kind to make sure next search gets updated results
	if err := searchIndex.Delete(); err != nil {
		log.Warnf("failed to delete the search index for store kind %q. Next search might be based on outdated information.", kubeconfigStore.GetKind())
		return err
	}

	// check which shoots are deleted and which are added
	addedShoots := shootIdentifiers.Difference(oldShootIdentifiers)
	removedShoots := oldShootIdentifiers.Difference(shootIdentifiers)
	fmt.Printf("\u001B[1;33m%s\u001B[0m: \n - Wrote kubeconfigs for \u001B[1;32m%d shoots\u001B[0m on \033[1;34m%d seeds\033[0m (%d shooted seeds) to the %s with path %q.\n - \u001B[1;31mDeleted %d Shoots\u001B[0m. \n - \u001B[1;32mAdded %d Shoots\u001B[0m. \n", "Summary", shootIdentifiers.Len(), len(seedNames), len(shootedSeedNames), kubeconfigStore.GetKind(), fmt.Sprintf("%s/%s", exportPath, landscapeName), len(removedShoots), len(addedShoots))

	// check which shooted seeds are deleted and which are added
	addedShootedSeeds := shootedSeedIdentifiers.Difference(oldSeedIdentifiers)
	removedShootedSeeds := oldSeedIdentifiers.Difference(shootedSeedIdentifiers)
	fmt.Printf(" - \u001B[1;31mDeleted %d Shooted Seeds\u001B[0m. \n - \u001B[1;32mAdded %d Shooted Seeds\u001B[0m.", len(removedShootedSeeds), len(addedShootedSeeds))
	fmt.Printf("\n \n")
	return nil
}

func GetPreviousIdentifiers(searchIndex *index.SearchIndex, landscape string) (sets.String, sets.String, error) {
	shootIdentifiers := sets.NewString()
	seedIdentifiers := sets.NewString()

	if !searchIndex.HasContent() {
		return sets.String{}, sets.String{}, nil
	}

	for _, kubeconfigFilepath := range searchIndex.GetContent() {
		parentDirectory := filepath.Dir(kubeconfigFilepath)
		name := filepath.Base(kubeconfigFilepath)
		// directories are created with a uniform prefix
		if strings.Contains(kubeconfigFilepath, fmt.Sprintf("%s-shoot-", landscape)) {
			shootIdentifiers.Insert(name)
		}
		// shooted seeds always are in a sub-directory "shooted-seeds"
		if strings.Contains(parentDirectory, "shooted-seeds") && strings.Contains(name, fmt.Sprintf("%s-seed-", landscape)) {
			seedIdentifiers.Insert(name)
		}
	}
	return shootIdentifiers, seedIdentifiers, nil
}

func isShootedSeed(shoot gardencorev1beta1.Shoot) bool {
	if shoot.Namespace == v1beta1constants.GardenNamespace && shoot.Annotations != nil {
		_, ok := v1beta1constants.GetShootUseAsSeedAnnotation(shoot.Annotations)
		return ok
	}
	return false
}

func getSecretIdentifier(namespace string, shootName string) string {
	return fmt.Sprintf("%s/%s", namespace, shootName)
}

// <landscape>-shoot-<project-name>-<shoot-name>
func getShootIdentifier(landscape, project, shoot string) string {
	return fmt.Sprintf("%s-shoot-%s-%s", landscape, project, shoot)
}

func getShootKubeconfigDirectory(rootDirectory, landscape, seedName, identifier string) string {
	// <landscape>/shoots/<seed>/<landscape>-shoot-<project-name>-<shoot-name>
	return fmt.Sprintf("%s/%s/shoots/seed-%s/%s", rootDirectory, landscape, seedName, identifier)
}

// <landscape>-seed-<seed-name>
func getSeedIdentifier(landscape, shoot string) string {
	return fmt.Sprintf("%s-seed-%s", landscape, shoot)
}

func getSeedKubeconfigDirectory(rootDirectory, landscape, identifier string) string {
	// <landscape>/seeds/<landscape>-seed-<seed-name>
	return fmt.Sprintf("%s/%s/shooted-seeds/%s", rootDirectory, landscape, identifier)
}
