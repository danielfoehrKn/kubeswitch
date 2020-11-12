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

func (s *FileStore) CreateLandscapeDirectory(landscapeDirectory string) error {
	// create root directory
	err := os.Mkdir(landscapeDirectory, 0700)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create filesystem directory for kubeconfigs %q: %v", exportPath, err)
	}
	return nil
}

func (s *FileStore) GetPreviousIdentifiers(dir, landscape string) (sets.String, sets.String, error) {
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

func (s *FileStore) WriteKubeconfigFile(directory, kubeconfigName string, kubeconfigSecret corev1.Secret) error {
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

func (s *FileStore) CleanExistingKubeconfigs(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}