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
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	eks "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/danielfoehrkn/kubeswitch/types"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	gkev1 "google.golang.org/api/container/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SearchResult is a full kubeconfig path discovered from the kubeconfig store
// given the contained kubeconfig path, the store knows how to retrieve and return the
// actual kubeconfig
type SearchResult struct {
	KubeconfigPath string
	Error          error
}

type KubeconfigStore interface {
	// GetID returns the unique store ID
	// should be
	// - "<store kind>.default" if the kubeconfigStore.ID is not set
	// - "<store kind>.<id>" if the kubeconfigStore.ID is set
	GetID() string

	// GetKind returns the store kind (e.g., filesystem)
	GetKind() types.StoreKind

	// GetContextPrefix returns the prefix for the kubeconfig context names displayed in the search result
	// includes the path to the kubeconfig in the backing store because some stores compute the prefix based on that
	GetContextPrefix(path string) string

	// VerifyKubeconfigPaths verifies that the configured search paths are valid
	// can also include additional preprocessing
	VerifyKubeconfigPaths() error

	// StartSearch starts the search over the configured search paths
	// and populates the results via the given channel
	StartSearch(channel chan SearchResult)

	// GetKubeconfigForPath returns the byte representation of the kubeconfig
	// the kubeconfig has to fetch the kubeconfig from its backing store (e.g., uses the HTTP API)
	GetKubeconfigForPath(path string) ([]byte, error)

	// GetLogger returns the logger of the store
	GetLogger() *logrus.Entry

	// GetStoreConfig returns the store's configuration from the switch config file
	GetStoreConfig() types.KubeconfigStore
}

// Previewer can be optionally implemented by stores to show custom preview content
// before the kubeconfig
type Previewer interface {
	GetSearchPreview(path string) (string, error)
}

type FilesystemStore struct {
	Logger                *logrus.Entry
	KubeconfigStore       types.KubeconfigStore
	KubeconfigName        string
	kubeconfigDirectories []string
	kubeconfigFilepaths   []string
}

type VaultStore struct {
	Logger          *logrus.Entry
	KubeconfigStore types.KubeconfigStore
	Client          *vaultapi.Client
	KubeconfigName  string
	vaultPaths      []string
}

type GardenerStore struct {
	Logger            *logrus.Entry
	KubeconfigStore   types.KubeconfigStore
	Client            client.Client
	Config            *types.StoreConfigGardener
	LandscapeIdentity string
	LandscapeName     string
	StateDirectory    string
	// if a search against the Gardener API has been executed, this is filled with
	// all the discovered secrets.
	// This way we can save some requests against the API when getting the kubeconfig later
	SecretNamespaceNameToSecret map[string]corev1.Secret
}

type EKSStore struct {
	Logger          *logrus.Entry
	KubeconfigStore types.KubeconfigStore
	Client          *awseks.Client
	Config          *types.StoreConfigEKS
	// DiscoveredClusters maps the kubeconfig path (az_<resource-group>--<cluster-name>) -> cluster
	// This is a cache for the clusters discovered during the initial search for kubeconfig paths
	// when not using a search index
	DiscoveredClusters map[string]*eks.Cluster
	StateDirectory     string
}

type GKEStore struct {
	Logger          *logrus.Entry
	KubeconfigStore types.KubeconfigStore
	GkeClient       *gkev1.Service
	Config          *types.StoreConfigGKE
	// DiscoveredClusters maps the kubeconfig path (gke--project-name--clusterName) -> cluster
	// This is a cache for the clusters discovered during the initial search for kubeconfig paths
	// when not using a search index
	DiscoveredClusters map[string]*gkev1.Cluster
	// ProjectNameToID contains a mapping projectName -> project ID
	// used to construct the kubeconfig path containing the project name instead of a technical project id
	ProjectNameToID map[string]string
	StateDirectory  string
}

type AzureStore struct {
	Logger *logrus.Entry
	// DiscoveredClustersMutex is a mutex allow many reads, one write mutex to synchronize writes
	// to the DiscoveredClusters map.
	// This can happen when a goroutine still discovers clusters while another goroutine computes the preview for a missing cluster.
	DiscoveredClustersMutex sync.RWMutex
	KubeconfigStore         types.KubeconfigStore
	AksClient               *armcontainerservice.ManagedClustersClient
	Config                  *types.StoreConfigAzure
	// DiscoveredClusters maps the kubeconfig path (az_<resource-group>--<cluster-name>) -> cluster
	// This is a cache for the clusters discovered during the initial search for kubeconfig paths
	// when not using a search index
	DiscoveredClusters map[string]*armcontainerservice.ManagedCluster
	StateDirectory     string
}
