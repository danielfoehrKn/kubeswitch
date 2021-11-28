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
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/disiqueira/gotree"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/danielfoehrkn/kubeswitch/types"
)

func init() {
	utilruntime.Must(apiv1.AddToScheme(scheme))
}

// NewAzureStore creates a new Azure store
func NewAzureStore(store types.KubeconfigStore, stateDir string) (*AzureStore, error) {
	storeConfig := &types.StoreConfigAzure{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(buf, storeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal Azure config: %w", err)
		}
	}

	return &AzureStore{
		Logger:             logrus.New().WithField("store", types.StoreKindAzure),
		KubeconfigStore:    store,
		Config:             storeConfig,
		StateDirectory:     stateDir,
		DiscoveredClusters: make(map[string]*armcontainerservice.ManagedCluster),
	}, nil
}

// InitializeAzureStore initializes the Azure store
func (s *AzureStore) InitializeAzureStore() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("obtaining Azure credentials failed: %v", err)
	}

	// TODO: upgrading to version github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice v0.1.0
	// var options *arm.ClientOptions
	// if s.Config.Endpoint != nil {
	// 	options = &arm.ClientOptions{
	// 		Host: arm.Endpoint(*s.Config.Endpoint),
	// 	}
	// }
	// s.AksClient = armcontainerservice.NewManagedClustersClient(*s.Config.SubscriptionID, cred, options)

	endpoint := "https://management.azure.com/"
	if s.Config.Endpoint != nil {
		endpoint = *s.Config.Endpoint
	}

	con := arm.NewConnection(endpoint, cred, nil)
	s.AksClient = armcontainerservice.NewManagedClustersClient(con, *s.Config.SubscriptionID)

	s.Logger.Debugf("Authenticated to subscription %s", *s.Config.SubscriptionID)
	return nil
}

// StartSearch starts the search for AKS clusters
// Limitation: Two seperate subscriptions should not have the same (resource_group, cluster-name) touple
func (s *AzureStore) StartSearch(channel chan SearchResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.InitializeAzureStore(); err != nil {
		err := fmt.Errorf("failed to initialize store: %w", err)
		channel <- SearchResult{
			Error: err,
		}
		return
	}

	// uses dedicated lists per resource group
	if len(s.Config.ResourceGroups) > 0 {
		// TODO: optimize using goroutines to hide I/O latency
		for _, resourceGroup := range s.Config.ResourceGroups {
			pager := s.AksClient.ListByResourceGroup(resourceGroup, nil)
			if pager.Err() != nil {
				handleAzureError(channel, pager.Err())
				return
			}

			for pager.NextPage(ctx) {
				s.Logger.Debugf("next page found for resource group %q", resourceGroup)
				s.returnSearchResultsForClusters(channel, pager.PageResponse().ManagedClusterListResult.Value, &resourceGroup)
			}

			if pager.Err() != nil {
				handleAzureError(channel, pager.Err())
				return
			}
		}

		s.Logger.Debugf("search done for AKS resource groups")
		return
	}

	pager := s.AksClient.List(nil)
	if pager.Err() != nil {
		handleAzureError(channel, pager.Err())
		return
	}

	for pager.NextPage(ctx) {
		s.Logger.Debugf("next page found")
		s.returnSearchResultsForClusters(channel, pager.PageResponse().ManagedClusterListResult.Value, nil)
	}
	s.Logger.Debugf("search done for AKS")
}

func handleAzureError(channel chan SearchResult, err error) {
	if err, ok := err.(armcontainerservice.CloudError); ok && err.InnerError != nil {
		// TODO: if 401 is returned, execute `az cli` to re-authenticate
		// similar to gcp
		channel <- SearchResult{
			Error: fmt.Errorf("AKS returned an error listing AKS clusters: %w", err),
		}
		return
	}

	if err != nil {
		channel <- SearchResult{
			Error: fmt.Errorf("failed to list AKS clusters: %w", err),
		}
		return
	}
}

func (s *AzureStore) returnSearchResultsForClusters(channel chan SearchResult, managedClusters []*armcontainerservice.ManagedCluster, resourceGroup *string) {
	for _, cluster := range managedClusters {
		s.Logger.Debugf("Found cluster with name %q and id %q", *cluster.Name, *cluster.ID)
		if cluster.Name == nil {
			continue
		}

		if cluster.Resource.Name == nil {
			s.Logger.Debugf("resource name for cluster %q not set", *cluster.Resource.Name)
			continue
		}

		if resourceGroup == nil {
			if cluster.ID == nil {
				// this should not happen
				continue
			}

			// This is a hack: parse the resource group from the ID
			// there is unfortunately currently no easy way to get the resource group of an AKS cluster as the go-sdk does not expose that field :/
			//  - /subscriptions/<subscription-id>/resourcegroups/kubeswitch/providers/Microsoft.ContainerService/managedClusters/kubeswitch_test
			split := strings.Split(*cluster.ID, "/")
			if len(split) <= 4 {
				s.Logger.Debugf("Unable to obtain resource group for cluster %q from cluster ID  %q", *cluster.Resource.Name, *cluster.ID)
				continue
			}

			resourceGroup = &split[4]
			s.Logger.Debugf("Obtained resource group %s", *resourceGroup)
		}

		kubeconfigPath := getAzureKubeconfigPath(*resourceGroup, *cluster.Name)
		s.insertIntoClusterCache(kubeconfigPath, cluster)

		channel <- SearchResult{
			KubeconfigPath: kubeconfigPath,
			Error:          nil,
		}
	}
}

func getAzureKubeconfigPath(resourceGroup, clusterName string) string {
	return fmt.Sprintf("az_%s--%s", resourceGroup, clusterName)
}

func (s *AzureStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	// the GKE store encodes the path with semantic information
	// <project-name>--<location>--<cluster-name>
	// just use this semantic information as a prefix & remove the double dashes
	return strings.ReplaceAll(path, "--", "-")
}

// IsInitialized checks if the store has been initialized already
func (s *AzureStore) IsInitialized() bool {
	return s.AksClient != nil && s.Config != nil
}

func (s *AzureStore) GetID() string {
	id := "default"

	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}

	return fmt.Sprintf("%s.%s", types.StoreKindAzure, id)
}

func (s *AzureStore) GetKind() types.StoreKind {
	return types.StoreKindAzure
}

func (s *AzureStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *AzureStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *AzureStore) GetKubeconfigForPath(path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if !s.IsInitialized() {
		if err := s.InitializeAzureStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize Azure store: %w", err)
		}
	}
	resourceGroup, clusterName, err := parseAzureIdentifier(path)
	if err != nil {
		return nil, err
	}

	s.Logger.Debugf("AKS: GetKubeconfigForPath for group : %q and cluster: %q", resourceGroup, clusterName)

	// Documentation: support user credentials in the future
	resp, err := s.AksClient.ListClusterAdminCredentials(ctx, resourceGroup, clusterName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain kubeconfig for AKS cluster %q in resource group %q: %w", clusterName, resourceGroup, err)
	}

	for _, kubeconfig := range resp.Kubeconfigs {
		if kubeconfig != nil && len(kubeconfig.Value) > 0 {
			return kubeconfig.Value, err
		}
	}
	return nil, fmt.Errorf("no admin kubeconfig found for AKS cluster %q in resource group %q", clusterName, resourceGroup)
}

func (s *AzureStore) VerifyKubeconfigPaths() error {
	// NOOP
	return nil
}

// ParseIdentifier takes a kubeconfig identifier and
// returns the
// 1) the Azure resource group
// 2) the name of the AKS cluster
func parseAzureIdentifier(path string) (string, string, error) {
	split := strings.Split(path, "--")
	switch len(split) {
	case 2:
		return strings.TrimPrefix(split[0], "az_"), split[1], nil
	default:
		return "", "", fmt.Errorf("unable to parse kubeconfig path: %q", path)
	}
}

func (s *AzureStore) GetSearchPreview(path string) (string, error) {
	if !s.IsInitialized() {
		// this takes too long, initialize concurrently
		go func() {
			if err := s.InitializeAzureStore(); err != nil {
				s.Logger.Debugf("failed to initialize store: %v", err)
			}
		}()
		return "", fmt.Errorf("azure store is not initalized yet")
	}

	// low timeout to not pile up many requests, but timeout fast
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resourceGroup, clusterName, err := parseAzureIdentifier(path)
	if err != nil {
		return "", err
	}

	// the cluster should be in the cache, but do not fail if it is not
	cluster := s.readFromClusterCache(path)

	// cluster has not been discovered from the AKS API yet
	// this is the case when a search index is used
	if cluster == nil {
		// The name (resource_group, cluster) of the cluster to retrieve.
		// we can safely use the client, as we know the store has been previously initialized
		resp, err := s.AksClient.Get(ctx, resourceGroup, clusterName, nil)
		if err != nil {
			return "", fmt.Errorf("failed to get Azure cluster with name %q : %w", clusterName, err)
		}
		cluster = &resp.ManagedCluster
		s.insertIntoClusterCache(path, cluster)
	}

	asciTree := gotree.New(clusterName)

	if cluster.Properties.KubernetesVersion != nil {
		asciTree.Add(fmt.Sprintf("Kubernetes Version: %s", *cluster.Properties.KubernetesVersion))
	}

	if cluster.Properties.ProvisioningState != nil && cluster.Properties.PowerState != nil {
		asciTree.Add(fmt.Sprintf("Status: %s(%s)", *cluster.Properties.ProvisioningState, *cluster.Properties.PowerState.Code))
	}

	asciTree.Add(fmt.Sprintf("Resource group: %s", resourceGroup))

	if cluster.Location != nil {
		asciTree.Add(fmt.Sprintf("Location: %s", *cluster.Location))
	}

	asciTree.Add(fmt.Sprintf("Subscription ID: %s", *s.Config.SubscriptionID))

	return asciTree.Print(), nil
}

func (s *AzureStore) readFromClusterCache(key string) *armcontainerservice.ManagedCluster {
	s.DiscoveredClustersMutex.RLock()
	defer s.DiscoveredClustersMutex.RUnlock()
	return s.DiscoveredClusters[key]
}

func (s *AzureStore) insertIntoClusterCache(key string, value *armcontainerservice.ManagedCluster) {
	s.DiscoveredClustersMutex.Lock()
	defer s.DiscoveredClustersMutex.Unlock()
	s.DiscoveredClusters[key] = value
}
