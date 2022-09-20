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
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/types"
)

func NewVaultStore(vaultAPIAddressFromFlag, vaultTokenFileName, kubeconfigName string, kubeconfigStore types.KubeconfigStore) (*VaultStore, error) {
	vaultStoreConfig := &types.StoreConfigVault{}
	if kubeconfigStore.Config != nil {
		buf, err := yaml.Marshal(kubeconfigStore.Config)
		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(buf, vaultStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vault config: %w", err)
		}
	}

	vaultAPI := vaultStoreConfig.VaultAPIAddress
	if len(vaultAPIAddressFromFlag) > 0 {
		vaultAPI = vaultAPIAddressFromFlag
	}

	vaultAddress := os.Getenv("VAULT_ADDR")
	if len(vaultAddress) > 0 {
		vaultAPI = vaultAddress
	}

	if len(vaultAPI) == 0 {
		return nil, fmt.Errorf("when using the vault kubeconfig store, the API address of the vault has to be provided either by command line argument \"vaultAPI\", via environment variable \"VAULT_ADDR\" or via SwitchConfig file")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var vaultToken string

	// https://www.vaultproject.io/docs/commands/token-helper
	tokenBytes, _ := os.ReadFile(fmt.Sprintf("%s/%s", home, vaultTokenFileName))
	if tokenBytes != nil {
		vaultToken = string(tokenBytes)
	}

	vaultTokenEnv := os.Getenv("VAULT_TOKEN")
	if len(vaultTokenEnv) > 0 {
		vaultToken = vaultTokenEnv
	}

	if len(vaultToken) == 0 {
		return nil, fmt.Errorf("when using the vault kubeconfig store, a vault API token must be provided. Per default, the token file in \"~.vault-token\" is used. The default token can be overriden via the environment variable \"VAULT_TOKEN\"")
	}

	vaultConfig := &vaultapi.Config{
		Address: vaultAPI,
	}
	client, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetToken(vaultToken)

	return &VaultStore{
		Logger:          logrus.New().WithField("store", types.StoreKindVault),
		KubeconfigName:  kubeconfigName,
		KubeconfigStore: kubeconfigStore,
		Client:          client,
	}, nil
}

func (s *VaultStore) GetID() string {
	id := "default"
	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}
	return fmt.Sprintf("%s.%s", types.StoreKindVault, id)
}

func (s *VaultStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	// for vault, the secret name itself contains the semantic information (not the key of the kv-pair of the vault secret)
	return filepath.Base(path)
}

func (s *VaultStore) GetKind() types.StoreKind {
	return types.StoreKindVault
}

func (s *VaultStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *VaultStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *VaultStore) recursivePathTraversal(wg *sync.WaitGroup, searchPath string, channel chan SearchResult) {
	defer wg.Done()

	secret, err := s.Client.Logical().List(searchPath)
	if err != nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}

	if secret == nil {
		s.Logger.Infof("No secrets found for path %s", searchPath)
		return
	}

	items := secret.Data["keys"].([]interface{})
	for _, item := range items {
		itemPath := fmt.Sprintf("%s/%s", strings.TrimSuffix(searchPath, "/"), item)
		if strings.HasSuffix(item.(string), "/") {
			// this is another folder
			wg.Add(1)
			go s.recursivePathTraversal(wg, itemPath, channel)
		} else if item != "" {
			// found an actual secret
			channel <- SearchResult{
				KubeconfigPath: itemPath,
				Error:          err,
			}
		}
	}
}

func (s *VaultStore) StartSearch(channel chan SearchResult) {
	wg := sync.WaitGroup{}
	// start multiple recursive searches from different root paths
	for _, path := range s.vaultPaths {
		s.Logger.Debugf("discovering secrets from vault under path %q", path)

		wg.Add(1)
		go s.recursivePathTraversal(&wg, path, channel)
	}
	wg.Wait()
}

func getBytesFromSecretValue(v interface{}) ([]byte, error) {
	data, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("failed to marshal value into string")
	}

	bytes := []byte(data)

	// check if it is base64 encode - if yes use the decoded version
	base64, err := base64.StdEncoding.DecodeString(data)
	if err == nil {
		bytes = base64
	}
	return bytes, nil
}

func (s *VaultStore) GetKubeconfigForPath(path string) ([]byte, error) {
	s.Logger.Debugf("vault: getting secret for path %q", path)

	secret, err := s.Client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("could not read secret with path '%s': %v", path, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("no kubeconfig found for path %s", path)
	}

	if len(secret.Data) != 1 {
		return nil, fmt.Errorf("cannot read kubeconfig from %q. Only support one entry in the secret", path)
	}

	for secretKey, data := range secret.Data {
		matched, err := filepath.Match(s.KubeconfigName, secretKey)
		if err != nil {
			return nil, err
		}
		if !matched {
			return nil, fmt.Errorf("cannot read kubeconfig from %q. Key %q does not match desired kubeconfig name", path, s.KubeconfigName)
		}

		bytes, err := getBytesFromSecretValue(data)
		if err != nil {
			return nil, fmt.Errorf("cannot read kubeconfig from %q: %v", path, err)
		}
		return bytes, nil
	}
	return nil, fmt.Errorf("should not happen")
}

func (s *VaultStore) VerifyKubeconfigPaths() error {
	var duplicatePath = make(map[string]*struct{})

	for _, path := range s.KubeconfigStore.Paths {
		// do not add duplicate paths
		if duplicatePath[path] != nil {
			continue
		}
		duplicatePath[path] = &struct{}{}

		_, err := s.Client.Logical().Read(path)
		if err != nil {
			return err
		}

		s.vaultPaths = append(s.vaultPaths, path)
	}
	return nil
}
