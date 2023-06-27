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
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
	paths "path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/vault/api"
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

	engineversion := vaultStoreConfig.VaultEngineVersion
	if len(engineversion) == 0 {
		engineversion = "v1"
	}

	vaultKeyKubeconfig := vaultStoreConfig.VaultKeyKubeconfig
	if len(vaultKeyKubeconfig) == 0 {
		vaultKeyKubeconfig = "config"
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
		Logger:             logrus.New().WithField("store", types.StoreKindVault),
		KubeconfigName:     kubeconfigName,
		KubeconfigStore:    kubeconfigStore,
		VaultKeyKubeconfig: vaultKeyKubeconfig,
		Client:             client,
		EngineVersion:      engineversion,
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

// recursivePathTraversal dfs-traverses the secrets tree rooted at the given path
// and calls the `visit` functor for each of the directory and leaf paths.
// Note: for kv-v2, a "metadata" path is expected and "metadata" paths will be
// returned in the visit functor.
func (s *VaultStore) recursivePathTraversal(wg *sync.WaitGroup, ctx context.Context, client *api.Client, path string, visit func(path string, directory bool) error) {
	defer wg.Done()

	resp, err := client.Logical().ListWithContext(ctx, path)
	if err != nil {
		s.Logger.Errorf("could not list %q path: %s", path, err)
		return
	}

	if resp == nil || resp.Data == nil {
		s.Logger.Errorf("no value found at %q: %s", path, err)
		return
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		s.Logger.Errorf("unexpected list response at %q", path)
		return
	}

	keysRawSlice, ok := keysRaw.([]interface{})
	if !ok {
		s.Logger.Errorf("unexpected list response type %T at %q", keysRaw, path)
		return
	}

	keys := make([]string, 0, len(keysRawSlice))

	for _, keyRaw := range keysRawSlice {
		key, ok := keyRaw.(string)
		if !ok {
			s.Logger.Errorf("unexpected key type %T at %q", keyRaw, path)
			return
		}
		keys = append(keys, key)
	}

	// sort the keys for a deterministic output
	sort.Strings(keys)

	for _, key := range keys {
		// the keys are relative to the current path: combine them
		child := paths.Join(path, key)

		if strings.HasSuffix(key, "/") {
			// visit the directory
			if err := visit(child, true); err != nil {
				return
			}

			// this is not a leaf node: we need to go deeper...
			wg.Add(1)
			go s.recursivePathTraversal(wg, ctx, client, child, visit)
		} else {
			// this is a leaf node: add it to the list
			if err := visit(child, false); err != nil {
				return
			}
		}
	}
}

func (s *VaultStore) StartSearch(channel chan SearchResult) {
	wg := sync.WaitGroup{}
	// start multiple recursive searches from different root paths
	for _, path := range s.vaultPaths {
		// Checking secret engine version. If it's v2, we should shim /metadata/
		// to secret path if necessary.
		var secretsPath string
		if s.EngineVersion == "v2" {
			mountPath := strings.Split(path, "/")[0]
			secretsPath = shimKvV2ListPath(path, mountPath)
		} else {
			secretsPath = path
		}
		s.Logger.Debugf("discovering secrets from vault under path %q", secretsPath)

		wg.Add(1)
		go s.recursivePathTraversal(&wg, context.Background(), s.Client, secretsPath, func(path string, directory bool) error {
			// found an actual secret, but remove "metadata/" from the path
			rawPath := shimKVv2Metadata(path)
			channel <- SearchResult{
				KubeconfigPath: rawPath,
				Error:          nil,
			}
			s.Logger.Debugf("Found %s", rawPath)
			return nil
		})
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

	// Checking secret engine version. If it's v2, we should shim /metadata/
	// to secret path if necessary.
	var secretsPath string
	if s.EngineVersion == "v2" {
		mountPath := strings.Split(path, "/")[0]
		secretsPath = shimKVv2Path(path, mountPath)
	} else {
		secretsPath = path
	}

	s.Logger.Debugf("vault: getting secret for path %q", secretsPath)
	secret, err := s.Client.Logical().Read(secretsPath)
	if err != nil {
		return nil, fmt.Errorf("could not read secret with path '%s': %v", secretsPath, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("no kubeconfig found for path %s", secretsPath)
	}

	if (s.EngineVersion == "v1") && len(secret.Data) != 1 {
		return nil, fmt.Errorf("cannot read kubeconfig from %q. Only support one entry in the secret if we using v1", secretsPath)
	}

	if s.EngineVersion == "v1" {
		for secretKey, data := range secret.Data {
			matched, err := filepath.Match(s.KubeconfigName, secretKey)
			if err != nil {
				return nil, err
			}
			if !matched {
				return nil, fmt.Errorf("cannot read kubeconfig from %q. Key %q does not match desired kubeconfig name", secretsPath, s.KubeconfigName)
			}

			if matched {
				bytes, err := getBytesFromSecretValue(data)
				if err != nil {
					return nil, fmt.Errorf("cannot read kubeconfig from %q: %v", secretsPath, err)
				}
				return bytes, nil
			}
		}
	} else {
		if secret.Data["data"] == nil {
			return nil, fmt.Errorf("cannot read kubeconfig from %q. Secret is empty.", secretsPath)
		}
		value, ok := secret.Data["data"].(map[string]interface{})[s.VaultKeyKubeconfig]
		if ok {
			bytes, err := getBytesFromSecretValue(value)
			if err != nil {
				return nil, fmt.Errorf("cannot read kubeconfig from %q: %v", secretsPath, err)
			}
			return bytes, nil
		}
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

		// Checking secret engine version. If it's v2, we should shim /metadata/
		// to secret path if necessary.
		var secretsPath string
		if s.EngineVersion == "v2" {
			mountPath := strings.Split(path, "/")[0]
			secretsPath = shimKvV2ListPath(path, mountPath)
		} else {
			secretsPath = path
		}

		_, err := s.Client.Logical().Read(secretsPath)
		if err != nil {
			return err
		}

		s.vaultPaths = append(s.vaultPaths, secretsPath)
	}
	return nil
}

// shimKVv2Path aligns the supported legacy path to KV v2 specs by inserting
// /data/ into the path for reading secrets. Paths for metadata are not modified.
func shimKVv2Path(rawPath, mountPath string) string {
	switch {
	case rawPath == mountPath, rawPath == strings.TrimSuffix(mountPath, "/"):
		return path.Join(mountPath, "data")
	default:
		p := strings.TrimPrefix(rawPath, mountPath)

		// Only add /data/ prefix to the path if neither /data/ or /metadata/ are
		// present.
		if strings.HasPrefix(p, "data/") || strings.HasPrefix(p, "metadata/") {
			return rawPath
		}
		return path.Join(mountPath, "data", p)
	}
}

// shimKvV2ListPath aligns the supported legacy path to KV v2 specs by inserting
// /metadata/ into the path for listing secrets. Paths with /metadata/ are not modified.
func shimKvV2ListPath(rawPath, mountPath string) string {
	mountPath = strings.TrimSuffix(mountPath, "/")

	if strings.HasPrefix(rawPath, path.Join(mountPath, "metadata")) {
		// It doesn't need modifying.
		return rawPath
	}

	switch {
	case rawPath == mountPath:
		return path.Join(mountPath, "metadata")
	default:
		rawPath = strings.TrimPrefix(rawPath, mountPath)
		return path.Join(mountPath, "metadata", rawPath)
	}
}

// shimKVv2Metadata removes metadata/ from the path
func shimKVv2Metadata(path string) string {
	return strings.Replace(path, "metadata/", "", -1)
}
