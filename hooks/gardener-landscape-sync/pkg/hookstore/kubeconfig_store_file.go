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

package hookstore

import (
	"fmt"
	"os"

	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/danielfoehrkn/kubeswitch/types"
)

func (s *FileStore) GetKind() types.StoreKind {
	return types.StoreKindFilesystem
}

func (s *FileStore) CreateLandscapeDirectory(landscapeDirectory string) error {
	// create root directory
	err := os.Mkdir(landscapeDirectory, 0700)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create filesystem directory for kubeconfigs %q: %v", landscapeDirectory, err)
	}
	return nil
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

	kubeconfig := kubeconfigSecret.Data[secrets.DataKeyKubeconfig]
	_, err = file.Write(kubeconfig)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileStore) CleanExistingKubeconfigs(log *logrus.Entry, dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}
