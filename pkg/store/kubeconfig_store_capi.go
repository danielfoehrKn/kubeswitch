// Copyright 2024 The Kubeswitch authors
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
	"time"

	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	utilkubeconfig "sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewCapiStore(store types.KubeconfigStore, stateDir string) (*CapiStore, error) {
	storeConfig := &types.StoreConfigCapi{}
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

	return &CapiStore{
		KubeconfigStore: store,
		Logger:          logrus.New().WithField("store", types.StoreKindCapi),
		Config:          storeConfig,
	}, nil
}

func (s *CapiStore) InitializeCapiStore() error {
	s.Logger.Info("Initializing CAPI client")
	k8sclient, err := s.getCapiClient()
	if err != nil {
		return err
	}
	s.Client = k8sclient

	return nil
}

// GetID returns the unique store ID
func (s *CapiStore) GetID() string {
	return fmt.Sprintf("%s.%s", types.StoreKindCapi, *s.KubeconfigStore.ID)
}

// GetKind returns the store kind
func (s *CapiStore) GetKind() types.StoreKind {
	return types.StoreKindCapi
}

// GetContextPrefix returns the context prefix
func (s *CapiStore) GetContextPrefix(path string) string {
	return string(types.StoreKindCapi)
}

// VerifyKubeconfigPaths verifies the kubeconfig paths
func (s *CapiStore) VerifyKubeconfigPaths() error {
	return nil
}

func (s *CapiStore) getCapiClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(clusterv1beta1.AddToScheme(scheme))

	// client from s.Config.KubeconfigPath
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: s.Config.KubeconfigPath},
		&clientcmd.ConfigOverrides{})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create rest config: %v", err)
	}

	k8sclient, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create client: %v", err)
	}
	return k8sclient, nil
}

// StartSearch starts the search over the configured search paths
func (s *CapiStore) StartSearch(channel chan storetypes.SearchResult) {
	s.Logger.Debug("CAPI: start search")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// initialize CAPI client
	if err := s.InitializeCapiStore(); err != nil {
		channel <- storetypes.SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}

	// list clusters
	clusters := &clusterv1beta1.ClusterList{}
	err := s.Client.List(ctx, clusters)
	if err != nil {
		channel <- storetypes.SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}

	for _, cluster := range clusters.Items {
		s.Logger.Debug("CAPI: found cluster", "name", cluster.Name, "namespace", cluster.Namespace)

		channel <- storetypes.SearchResult{
			KubeconfigPath: fmt.Sprintf("%s-%s", cluster.Namespace, cluster.Name),
			Error:          nil,
			Tags: map[string]string{
				"namespace": cluster.Namespace,
				"name":      cluster.Name,
			},
		}
	}
}

// GetKubeconfigForPath returns the kubeconfig for the path
func (s *CapiStore) GetKubeconfigForPath(path string, tags map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	s.Logger.Debug("CAPI: GetKubeconfigForPath", "path", path)

	obj := client.ObjectKey{
		Namespace: tags["namespace"],
		Name:      tags["name"],
	}
	dataBytes, err := utilkubeconfig.FromSecret(ctx, s.Client, obj)
	if err != nil {
		s.Logger.Debug("CAPI: GetKubeconfigForPath", "error", err)
		return nil, err
	}
	return dataBytes, nil
}

func (s *CapiStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *CapiStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}
