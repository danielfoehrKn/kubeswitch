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
	"fmt"

	"github.com/rancher/norman/clientbase"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/types"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

func NewRancherStore(store types.KubeconfigStore) (*RancherStore, error) {
	rancherStoreConfig := &types.StoreConfigRancher{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to process Rancher store config: %w", err)
		}

		err = yaml.Unmarshal(buf, rancherStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal Rancher config: %w", err)
		}
	}

	rancherAPIAddress := rancherStoreConfig.RancherAPIAddress
	if len(rancherAPIAddress) == 0 {
		return nil, fmt.Errorf("when using the Rancher kubeconfig store, the address of Rancher has to be provided via SwitchConfig file")
	}

	rancherToken := rancherStoreConfig.RancherToken
	if len(rancherToken) == 0 {
		return nil, fmt.Errorf("when using the Rancher kubeconfig store, a Rancher API token must be provided via SwitchConfig file")
	}

	return &RancherStore{
		Logger:          logrus.New().WithField("store", types.StoreKindRancher),
		KubeconfigStore: store,
		ClientOpts: &clientbase.ClientOpts{
			URL:      rancherAPIAddress,
			TokenKey: rancherToken,
		},
	}, nil
}

func (r *RancherStore) GetID() string {
	id := "default"
	if r.KubeconfigStore.ID != nil {
		id = *r.KubeconfigStore.ID
	}
	return fmt.Sprintf("%s.%s", types.StoreKindRancher, id)
}

func (r *RancherStore) GetContextPrefix(path string) string {
	if r.GetStoreConfig().ShowPrefix != nil && !*r.GetStoreConfig().ShowPrefix {
		return ""
	}
	return path
}

func (r *RancherStore) GetKind() types.StoreKind {
	return types.StoreKindRancher
}

func (r *RancherStore) GetStoreConfig() types.KubeconfigStore {
	return r.KubeconfigStore
}

func (r *RancherStore) GetLogger() *logrus.Entry {
	return r.Logger
}

// initClient initializes the Rancher client
// It is called once at the beginning of the search and every time a kubenfig is requested
// It is a NOOP if the client is already initialized
func (r *RancherStore) initClient() error {
	if r.Client != nil { // already initialized
		return nil
	}

	client, err := managementClient.NewClient(r.ClientOpts)
	if err != nil {
		return fmt.Errorf("failed to create Rancher client: %w", err)
	}

	r.Client = client
	return nil
}

func (r *RancherStore) StartSearch(channel chan SearchResult) {
	r.Logger.Debug("Rancher: start search")

	if err := r.initClient(); err != nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("failed to initialize Rancher client: %w", err),
		}
		return
	}

	cluster, err := r.Client.Cluster.ListAll(nil)
	if err != nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}
	for _, v := range cluster.Data {
		id := v.ID
		if id == "local" {
			// rancher uses "local" as id for its base cluster
			// if multiple rancher stores are used, this always leads to conflicts with the local cluster.
			// As a workaround the id of the store is used for the local cluster
			id = r.GetID()
		}
		channel <- SearchResult{
			KubeconfigPath: id,
			Error:          nil,
		}
	}
}

func (r *RancherStore) GetKubeconfigForPath(path string, _ map[string]string) ([]byte, error) {
	r.Logger.Debugf("Rancher: getting secret for path %q", path)

	if err := r.initClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize Rancher client: %w", err)
	}

	clusterID := path
	if clusterID == r.GetID() {
		// local cluster was replaced in StartSearch; restore original id
		clusterID = "local"
	}

	cluster, err := r.Client.Cluster.ByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster '%s': %w", path, err)
	}

	kubeconfig, err := r.Client.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig for cluster '%s': %w", path, err)
	}
	return []byte(kubeconfig.Config), nil
}

func (r *RancherStore) VerifyKubeconfigPaths() error {
	return nil
}
