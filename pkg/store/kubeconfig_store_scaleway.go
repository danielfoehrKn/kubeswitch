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

	"github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/types"
)

func NewScalewayStore(store types.KubeconfigStore) (*ScalewayStore, error) {
	scalewayStoreConfig := &types.StoreConfigScaleway{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to process Scaleway store config: %w", err)
		}

		err = yaml.Unmarshal(buf, scalewayStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal Scaleway config: %w", err)
		}
	}
	logger := logrus.New().WithField("store", types.StoreKindScaleway)

	scalewayAccessKey := scalewayStoreConfig.ScalewayAccessKey
	if len(scalewayAccessKey) == 0 {
		return nil, fmt.Errorf("When using the Scaleway kubeconfig store, the access key for Scaleway has to be provided via a SwitchConfig file")
	}
	scalewayOrganizationID := scalewayStoreConfig.ScalewayOrganizationID
	if len(scalewayOrganizationID) == 0 {
		return nil, fmt.Errorf("When using the Scaleway kubeconfig store, the organization ID for Scaleway has to be provided via a SwitchConfig file")
	}
	scalewaySecretKey := scalewayStoreConfig.ScalewaySecretKey
	if len(scalewaySecretKey) == 0 {
		return nil, fmt.Errorf("When using the Scaleway kubeconfig store, the secret key for Scaleway has to be provided via a SwitchConfig file")
	}
	scalewayRegion := scalewayStoreConfig.ScalewayRegion
	if len(scalewayRegion) == 0 {
		logger.Warning("No region specified for scaleway, using default 'fr-par'")
		scalewayRegion = "fr-par"
	}

	client, err := scw.NewClient(
		scw.WithDefaultOrganizationID(scalewayOrganizationID),
		scw.WithAuth(scalewayAccessKey, scalewaySecretKey),
		scw.WithDefaultRegion(scw.Region(scalewayRegion)),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Scaleway client due to error: %w", err)
	}

	return &ScalewayStore{
		Logger:             logger,
		KubeconfigStore:    store,
		Client:             client,
		DiscoveredClusters: make(map[string]ScalewayKube),
	}, nil
}

type ScalewayKube struct {
	ID      string
	Name    string
	Project string
}

func (s *ScalewayStore) GetID() string {
	id := "default"
	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}
	return fmt.Sprintf("%s.%s", types.StoreKindScaleway, id)
}

func (s *ScalewayStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	if s.GetStoreConfig().ID != nil {
		return *s.GetStoreConfig().ID
	}

	return string(types.StoreKindScaleway)
}

func (r *ScalewayStore) GetKind() types.StoreKind {
	return types.StoreKindScaleway
}

func (s *ScalewayStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *ScalewayStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *ScalewayStore) StartSearch(channel chan SearchResult) {
	s.Logger.Debug("Scaleway: start search")

	papi := account.NewProjectAPI(s.Client)
	if papi == nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("Failed to create scaleway project API"),
		}
		return
	}
	pres, err := papi.ListProjects(
		&account.ProjectAPIListProjectsRequest{},
	)
	if err != nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("Could no list projects in Scaleway err: %w", err),
		}
		return
	}
	// list projects

	kapi := k8s.NewAPI(s.Client)
	if kapi == nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("Failed to create Kubernetes API instance for scaleway err: %w", err),
		}
		return
	}

	for _, project := range pres.Projects {
		cres, err := kapi.ListClusters(&k8s.ListClustersRequest{ProjectID: &project.ID})
		if err != nil {
			channel <- SearchResult{
				KubeconfigPath: "",
				Error:          fmt.Errorf("Failed to retrieve Kubernetes cluster for project %v err: %w", project.Name, err),
			}
			return
		}
		if cres.TotalCount == 0 {
			s.Logger.Debug("No k8s clusters in project", project.Name)
			continue
		}
		for _, cluster := range cres.Clusters {
			s.DiscoveredClusters[cluster.ID] = ScalewayKube{ID: cluster.ID, Name: cluster.Name, Project: project.ID}
			channel <- SearchResult{
				KubeconfigPath: cluster.Name,
				Error:          nil,
			}
		}
	}
}

func (s *ScalewayStore) GetKubeconfigForPath(path string) ([]byte, error) {
	s.Logger.Debugf("Scaleway: getting secret for path %q", path)

	var cluster ScalewayKube
	for _, c := range s.DiscoveredClusters {
		if c.Name == path {
			cluster = c
		}
	}

	kapi := k8s.NewAPI(s.Client)

	config, err := kapi.GetClusterKubeConfig(&k8s.GetClusterKubeConfigRequest{
		ClusterID: cluster.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig for cluster '%s': %w", path, err)
	}
	return config.GetRaw(), nil
}

func (r *ScalewayStore) VerifyKubeconfigPaths() error {
	return nil
}
