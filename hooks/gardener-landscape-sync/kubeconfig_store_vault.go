package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gardener/gardener/pkg/utils/secrets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// NOOP
func (s *VaultStore) CreateLandscapeDirectory(landscapeDirectory string) error {
	return nil
}

func (s *VaultStore) GetPreviousIdentifiers(dir, landscape string) (sets.String, sets.String, error) {
	shootIdentifiers := sets.NewString()
	seedIdentifiers := sets.NewString()
	directory := fmt.Sprintf("%s/%s", dir, landscape)

	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return shootIdentifiers, seedIdentifiers, nil
		}
		return nil, nil, err
	}

	if err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// parent directories of directories with the kubeconfig are created with a uniform prefix
			if info.IsDir() && strings.Contains(info.Name(), fmt.Sprintf("%s-shoot-", landscape)) {
				shootIdentifiers.Insert(info.Name())
			}
			if info.IsDir() && strings.Contains(path, "shooted-seeds") && strings.Contains(info.Name(), fmt.Sprintf("%s-seed-", landscape)) {
				seedIdentifiers.Insert(info.Name())
			}
			return nil
		}); err != nil {
		return nil, nil, fmt.Errorf("failed to find kubeconfig files in directory: %v", err)
	}
	return shootIdentifiers, seedIdentifiers, nil
}

func (s *VaultStore) WriteKubeconfigFile(directory, kubeconfigName string, kubeconfigSecret corev1.Secret) error {
	var path = directory
	if len(s.vaultSecretEnginePathPrefix) > 0 {
		path = fmt.Sprintf("%s/%s", s.vaultSecretEnginePathPrefix, directory)
	}
	kubeconfigData, _ := kubeconfigSecret.Data[secrets.DataKeyKubeconfig]

	_, err := s.client.Logical().Write(path, map[string]interface{}{
		kubeconfigName: kubeconfigData,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *VaultStore) CleanExistingKubeconfigs(dir string) error {
	return nil
}
