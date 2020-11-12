package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/utils/secrets"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = logrus.New()

	gardenKubeconfigPath        string
	exportPath                  string
	landscapeName               string
	clean                       bool
	shootKubeconfigName         string
	kubeconfigStore             string
	vaultAPIAddress             string
	vaultSecretEnginePathPrefix string

	rootCommand = &cobra.Command{
		Use:   "sync",
		Short: "Sync the kubeconfig of Shoot clusters to the local filesystem.",
		Long:  `Hook for the \"switch\" tool for Gardener landscapes to sync the kubeconfigs of Shoot clusters to the local filesystem.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			var store KubeconfigStore
			switch kubeconfigStore {
			case KubeconfigStoreFilesystem:
				store = &FileStore{}
			case KubeconfigStoreVault:
				vaultAddress := os.Getenv("VAULT_ADDR")
				vaultToken := vaultAPIAddress
				if len(vaultToken) == 0 {
					vaultToken = os.Getenv("VAULT_TOKEN")
				}
				if len(vaultToken) == 0 {
					return fmt.Errorf("for the vault kubeconfig store, the vault API address has to be provided wither by command line argument \"vaultAPI\" or via environment variable \"VAULT_ADDR\"")
				}

				config := &vaultapi.Config{
					Address: vaultAddress,
				}
				client, err := vaultapi.NewClient(config)
				if err != nil {
					return err
				}
				client.SetToken(vaultToken)

				store = &VaultStore{
					client:                      client,
					vaultSecretEnginePathPrefix: vaultSecretEnginePathPrefix,
				}
			default:
				return fmt.Errorf("unknown store %q", kubeconfigStore)
			}

			return runHook(store)
		},
	}
)

func init() {
	logger.SetOutput(os.Stdout)
	rootCommand.Flags().StringVar(
		&gardenKubeconfigPath,
		"garden-kubeconfig-path",
		"",
		"local directory path to the kubeconfig of the Garden cluster. The cluster has to contain the Shoot resources.")
	rootCommand.Flags().StringVar(
		&exportPath,
		"export-directory",
		"",
		"root of the path where the Shoot kubeconfig files are exported to. The path for exported kubeconfigs is: export-directory/<landscape-name>/shoots/seed-<seed-name>/<landscape-name>-shoot-<project-name>-<shoot-name>.")
	rootCommand.Flags().StringVar(
		&landscapeName,
		"landscape-name",
		"",
		"name of the Gardener landscape e.g \"dev\".")
	rootCommand.Flags().StringVar(
		&shootKubeconfigName,
		"export-kubeconfig-name",
		"config",
		"name for all the exported shoot cluster kubeconfig files.")
	rootCommand.Flags().BoolVar(
		&clean,
		"clean-directory",
		false,
		"clean the export path and all sub paths before exporting the new kubeconfig files. Used to prevent holding on to kubeconfigs of already deleted clusters.")
	rootCommand.Flags().StringVar(
		&kubeconfigStore,
		"store",
		"filesystem",
		"the storage for the kubeconfig files. Can be either \"filesystem\" or \"vault\"")
	rootCommand.Flags().StringVar(
		&vaultSecretEnginePathPrefix,
		"vaultSecretEnginePathPrefix",
		"",
		"the prefix to use for the vault secret engine when exporting the Gardener kubeconfigs. Only used for store \"vault\".")

}

const (
	KubeconfigStoreFilesystem = "filesystem"
	KubeconfigStoreVault      = "vault"
)

type KubeconfigStore interface {
	CreateLandscapeDirectory(dir string) error
	GetPreviousIdentifiers(dir, landscape string) (sets.String, sets.String, error)
	WriteKubeconfigFile(directory, kubeconfigName string, kubeconfigSecret corev1.Secret) error
	CleanExistingKubeconfigs(dir string) error
}

type FileStore struct {
}

type VaultStore struct {
	client                      *vaultapi.Client
	vaultSecretEnginePathPrefix string
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func runHook(store KubeconfigStore) error {
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
			logger.Warnf("Secret %s/%s does not contain a kubeconfig. Skipping.", secret.Namespace, secret.Name)
			continue
		}

		var shootName string
		if len(secret.ObjectMeta.OwnerReferences) == 0 || secret.ObjectMeta.OwnerReferences[0].Kind != "Shoot" {
			if !strings.Contains(secret.Namespace, ".kubeconfig") {
				logger.Warnf("Secret %s/%s could not be associated with any Shoot. Skipping.", secret.Namespace, secret.Name)
				continue
			}
			shootName = strings.Split(secret.Namespace, ".kubeconfig")[0]
		} else {
			shootName = secret.ObjectMeta.OwnerReferences[0].Name
		}
		shootNameToSecret[getSecretIdentifier(secret.Namespace, shootName)] = secret
	}

	logger.Infof("Found %d kubeconfigs", len(shootNameToSecret))

	shoots := &gardencorev1beta1.ShootList{}
	if err := k8sclient.List(ctx, shoots, &client.ListOptions{}); err != nil {
		return fmt.Errorf("Failed to  list secret objects: %v", err)
	}
	logger.Infof("Found %d shoots", len(shoots.Items))

	projects := &gardencorev1beta1.ProjectList{}
	if err := k8sclient.List(ctx, projects, &client.ListOptions{}); err != nil {
		return fmt.Errorf("failed to  list secret objects: %v", err)
	}
	logger.Infof("Found %d projects", len(projects.Items))

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

	oldShootIdentifiers, oldSeedIdentifiers, err := store.GetPreviousIdentifiers(exportPath, landscapeName)
	if err != nil {
		logger.Warnf("Failed to get existing kubeconfigs from the filesystem under path %q: %v", exportPath, err)
	}

	if clean {
		if err := store.CleanExistingKubeconfigs(landscapeDirectory); err != nil {
			logger.Warnf("Failed to clean existing kubeconfigs from the filesystem under path %q: %v", exportPath, err)
		}
	}

	// create root directory
	if err := store.CreateLandscapeDirectory(landscapeDirectory); err != nil {
		return fmt.Errorf("failed to create landscape directory %q: %v", landscapeDirectory, err)
	}

	for _, shoot := range shoots.Items {
		seedName := *shoot.Spec.SeedName
		if _, ok := seedNames[seedName]; !ok {
			seedNames[seedName] = struct{}{}
		}
		projectName := namespaceToProjectName[shoot.Namespace]
		if len(projectName) == 0 {
			logger.Warnf("Could not find project for Shoot (%s/%s). Skipping.", shoot.Namespace, shoot.Name)
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
			kubeconfigDirectory = getShootKubeconfigDirectory(exportPath, landscapeName, seedName, identifier)
		}

		var (
			secret      = corev1.Secret{}
			secretFound bool
		)
		secret, secretFound = shootNameToSecret[getSecretIdentifier(shoot.Namespace, shoot.Name)]
		if !secretFound {
			if err := k8sclient.Get(ctx, client.ObjectKey{Namespace: shoot.Namespace, Name: fmt.Sprintf("%s.kubeconfig", shoot.Name)}, &secret); err != nil {
				if apierrors.IsNotFound(err) {
					logger.Warnf("Could not find kubeconfig secret for Shoot (%s/%s). Skipping.", shoot.Namespace, shoot.Name)
					continue
				}
				logger.Errorf("Failed to get kubeconfig secret for Shoot (%s/%s). Skipping.", shoot.Namespace, shoot.Name)
				continue
			}
		}

		if err := store.WriteKubeconfigFile(kubeconfigDirectory, shootKubeconfigName, secret); err != nil {
			return fmt.Errorf("unable to write kubeconfig to path: %s: %v", kubeconfigDirectory, err)
		}
		if isShootedSeed {
			shootedSeedIdentifiers.Insert(identifier)
		} else {
			shootIdentifiers.Insert(identifier)
		}

		if shootIdentifiers.Len()%30 == 0 {
			logger.Infof("Wrote %d shoot kubeconfigs.", len(shootIdentifiers))
		}
	}

	// check which shoots are deleted and which are added
	addedShoots := shootIdentifiers.Difference(oldShootIdentifiers)
	removedShoots := oldShootIdentifiers.Difference(shootIdentifiers)
	fmt.Printf("\u001B[1;33m%s\u001B[0m: \n - Wrote kubeconfigs for \u001B[1;32m%d shoots\u001B[0m on \033[1;34m%d seeds\033[0m (%d shooted seeds) to directory %q.\n - \u001B[1;31mDeleted %d Shoots\u001B[0m. \n - \u001B[1;32mAdded %d Shoots\u001B[0m. \n", "Summary", shootIdentifiers.Len(), len(seedNames), len(shootedSeedNames), fmt.Sprintf("%s/%s", exportPath, landscapeName), len(removedShoots), len(addedShoots))

	// check which shooted seeds are deleted and which are added
	addedShootedSeeds := shootedSeedIdentifiers.Difference(oldSeedIdentifiers)
	removedShootedSeeds := oldSeedIdentifiers.Difference(shootedSeedIdentifiers)
	fmt.Printf(" - \u001B[1;31mDeleted %d Shooted Seeds\u001B[0m. \n - \u001B[1;32mAdded %d Shooted Seeds\u001B[0m.", len(removedShootedSeeds), len(addedShootedSeeds))
	fmt.Printf("\n \n")
	return nil
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
