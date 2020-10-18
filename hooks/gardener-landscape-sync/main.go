package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/utils/secrets"
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

	gardenKubeconfigPath string
	exportDirectory      string
	landscapeName        string
	cleanDirectory       bool
	shootKubeconfigName  string

	rootCommand    = &cobra.Command{
		Use:   "sync",
		Short: "Sync the kubeconfig of Shoot clusters to the local filesystem.",
		Long: `Hook for the \"switch\" tool for Gardener landscapes to sync the kubeconfigs of Shoot clusters to the local filesystem.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHook()
		},
	}
)

func init() {
	logger.SetOutput(os.Stdout)
	rootCommand.Flags().StringVar(
		&gardenKubeconfigPath,
		"garden-kubeconfig-path",
		"",
		"path to the kubeconfig of the Garden cluster. The cluster has to contain the Shoot resources.")
	rootCommand.Flags().StringVar(
		&exportDirectory,
		"export-directory",
		"",
		"root of the directory where the Shoot kubeconfig files are exported to. The path for exported kubeconfigs is: export-directory/<landscape-name>/shoots/seed-<seed-name>/<landscape-name>-shoot-<project-name>-<shoot-name>.")
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
		&cleanDirectory,
		"clean-directory",
		false,
		"clean the export directory and all subdirectories before exporting the new kubeconfig files. Used to prevent holding on to kubeconfigs of already deleted clusters.")
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func runHook() error {
	if len(gardenKubeconfigPath) == 0 {
		return fmt.Errorf("must set the path to the kubeconfig of the Garden cluster")
	}
	if len(exportDirectory) == 0 {
		return fmt.Errorf("must set the export directory")
	}
	if len(landscapeName) == 0 {
		return fmt.Errorf("must provide a landscape name")
	}

	landscapeDirectory := fmt.Sprintf("%s/%s", exportDirectory, landscapeName)


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

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
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
		if  _, exists := secret.Data[secrets.DataKeyKubeconfig]; !exists {
			logger.Warnf("Secret %s/%s does not contain a kubeconfig. Skipping.", secret.Namespace, secret.Name)
			continue
		}

		var shootName string
		if len(secret.ObjectMeta.OwnerReferences) == 0 || secret.ObjectMeta.OwnerReferences[0].Kind != "Shoot"{
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
	seedNames := map[string]struct{}{}

	oldShootIdentifiers, err := getPreviousShootIdentifiersFromFilesystem(exportDirectory, landscapeName)
	if err != nil {
		logger.Warnf("Failed to get existing kubeconfigs from the filesystem under path %q: %v", exportDirectory, err)
	}

	if cleanDirectory {
		if err := cleanExistingKubeconfigs(landscapeDirectory); err != nil {
			logger.Warnf("Failed to clean existing kubeconfigs from the filesystem under path %q: %v", exportDirectory, err)
		}
	}

	// create root directory
	err = os.Mkdir(landscapeDirectory, 0700)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create directory for kubeconfigs %q: %v", exportDirectory, err)
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

		shootIdentifier := getShootIdentifier(landscapeName, projectName, shoot.Name)

		kubeconfigDirectory := getKubeconfigDirectory(exportDirectory, landscapeName, seedName, shootIdentifier, shootKubeconfigName)

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

		if err := writeKubeconfigFile(kubeconfigDirectory, shootKubeconfigName, secret); err != nil {
			return fmt.Errorf("unable to write kubeconfig to path: %s: %v", kubeconfigDirectory, err)
		}
		shootIdentifiers.Insert(shootIdentifier)

		if len(shootIdentifiers)%30 == 0 {
			logger.Infof("Wrote %d kubeconfigs.", len(shootIdentifiers))
		}
	}

	// check which shoots are deleted and which are added
	addedShoots := shootIdentifiers.Difference(oldShootIdentifiers)
	removedShoots := oldShootIdentifiers.Difference(shootIdentifiers)
	logger.Infof("Summary: Wrote kubeconfigs for %d shoots on %d seeds to directory %q. Deleted %d and added %d Shoots.", shootIdentifiers.Len(), len(seedNames), fmt.Sprintf("%s/%s", exportDirectory, landscapeName), len(removedShoots), len(addedShoots))
	return nil
}

func cleanExistingKubeconfigs(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}

func getSecretIdentifier(namespace string, shootName string) string {
	return fmt.Sprintf("%s/%s", namespace, shootName)
}

// <landscape>-shoot-<project-name>-<shoot-name>
func getShootIdentifier(landscape, project, shoot string) string {
	return fmt.Sprintf("%s-shoot-%s-%s", landscape, project, shoot)
}

func getKubeconfigDirectory(rootDirectory, landscape, seedName , identifier, kubeconfigName string) string {
	// <landscape>/shoots/<seed>/<landscape>-shoot-<project-name>-<shoot-name>
	return fmt.Sprintf("%s/%s/shoots/seed-%s/%s", rootDirectory, landscape, seedName, identifier)
}

func writeKubeconfigFile(directory, kubeconfigName string, kubeconfigSecret corev1.Secret) error {
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create directory %q: %v", directory, err)
	}

	filepath := fmt.Sprintf("%s/%s", directory, kubeconfigName)
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	kubeconfig, _ := kubeconfigSecret.Data[secrets.DataKeyKubeconfig]
	_, err = file.Write(kubeconfig)
	if err != nil {
		return err
	}
	return nil
}

func getPreviousShootIdentifiersFromFilesystem(dir, landscape string) (sets.String, error) {
	shootIdentifiers := sets.NewString()
	directory := fmt.Sprintf("%s/%s", dir, landscape)

	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return shootIdentifiers, nil
		}
		return nil, err
	}

	if err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// parent directories of directories with the kubeconfig are created with a uniform prefix
			if info.IsDir() && strings.Contains(info.Name(), fmt.Sprintf("%s-shoot-", landscape)){
				shootIdentifiers.Insert(info.Name())
			}
			return nil
		}); err != nil {
		return nil, fmt.Errorf("failed to find kubeconfig files in directory: %v", err)
	}
	return shootIdentifiers, nil
}