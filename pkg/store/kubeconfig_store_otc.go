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
	"github.com/disiqueira/gotree"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/clusters"
	"strings"
	"time"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack"

	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func NewOTCStore(store types.KubeconfigStore, _ string) (*OTCStore, error) {
	otcStoreConfig := &types.StoreConfigOTC{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(buf, otcStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal OTC config: %w", err)
		}
	}

	return &OTCStore{
		KubeconfigStore: store,
		Config:          otcStoreConfig,
	}, nil
}

func (s *OTCStore) InitializeOTCStore() error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var names []string
	if *s.Config.Cloud != "" {
		names = []string{*s.Config.Cloud}
	}
	env := openstack.NewEnv("OTC_")
	cloud, err := env.Cloud(names...)
	if err != nil {
		return fmt.Errorf("error getting cloud: %w", err)
	}

	// Cannot do this - there is a bug in gophertelekomcloud that STS key is not loaded from config
	// client, err := openstack.AuthenticatedClientFromCloud(cloud)
	// if err != nil {
	// 	return fmt.Errorf("error getting client: %w", err)
	// }

	opts, err := openstack.AuthOptionsFromInfo(&cloud.AuthInfo, cloud.AuthType)
	if err != nil {
		return fmt.Errorf("failed to convert AuthInfo to AuthOptsBuilder with Env vars: %s", err)
	}

	// fix the issue from upstream
	if aksk, ok := opts.(golangsdk.AKSKAuthOptions); ok {
		if aksk.SecurityToken == "" && cloud.AuthInfo.SecurityToken != "" {
			aksk.SecurityToken = cloud.AuthInfo.SecurityToken
			opts = aksk
		}
	}

	// finally get the client
	client, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return fmt.Errorf("failed to authenticate client: %s", err)
	}

	cceClient, err := openstack.NewCCE(client, golangsdk.EndpointOpts{})
	if err != nil {
		return fmt.Errorf("Error getting cce: %+v\n", err)
	}

	s.Client = cceClient
	return nil
}

func (s *OTCStore) IsInitialized() bool {
	return s.Client != nil && s.Config != nil
}

func (s *OTCStore) GetID() string {
	id := "default"

	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}

	return fmt.Sprintf("%s.%s", types.StoreKindOTC, id)
}

func (s *OTCStore) GetKind() types.StoreKind {
	return types.StoreKindOTC
}

func (s *OTCStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *OTCStore) GetLogger() *logrus.Entry {
	if s.Logger == nil {
		s.Logger = logrus.WithField("store", s.GetID())
	}
	return s.Logger
}

func (s *OTCStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	return strings.ReplaceAll(path, "--", "-")
}

func (s *OTCStore) VerifyKubeconfigPaths() error {
	// NOOP
	return nil
}

func (s *OTCStore) StartSearch(channel chan storetypes.SearchResult) {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.InitializeOTCStore(); err != nil {
		err := fmt.Errorf("failed to initialize store. This is most likely a problem with your provided OTC credentials: %v", err)
		channel <- storetypes.SearchResult{
			Error: err,
		}
		return
	}

	refinedClusters, err := clusters.List(s.Client, clusters.ListOpts{})
	if err != nil {
		channel <- storetypes.SearchResult{
			Error: fmt.Errorf("Error getting clusters: %+v\n", err),
		}
	}

	for _, cluster := range refinedClusters {
		kubeconfigPath := fmt.Sprintf("otc_%s--%s", *s.Config.Cloud, cluster.Metadata.Name)
		channel <- storetypes.SearchResult{
			KubeconfigPath: kubeconfigPath,
			Error:          nil,
			Tags: map[string]string{
				"id":      cluster.Metadata.Id,
				"name":    cluster.Metadata.Name,
				"version": cluster.Spec.Version,
				"type":    cluster.Spec.Type,
				"flavor":  cluster.Spec.Flavor,
				"status":  cluster.Status.Phase,
			},
		}
	}
	s.GetLogger().Debugf("Search done for OTC")
}

func (s *OTCStore) GetKubeconfigForPath(_ string, tags map[string]string) ([]byte, error) {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !s.IsInitialized() {
		if err := s.InitializeOTCStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize OTC store: %w", err)
		}
	}

	expiryOpts := clusters.ExpirationOpts{
		Duration: -1,
	}

	kubeconfig, err := clusters.GetCertWithExpiration(s.Client, tags["id"], expiryOpts).ExtractMap()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve cluster kubeconfig: %w", err)
	}

	kubeconfig = s.convertKubeconfig(tags["name"], kubeconfig)

	configYaml, err := yaml.Marshal(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal cluster kubeconfig: %w", err)
	}
	return configYaml, nil
}

func (s *OTCStore) convertKubeconfig(name string, kubeconfig map[string]interface{}) map[string]interface{} {
	if s.Config.SelectContext == nil {
		return kubeconfig
	}

	contexts, ok := kubeconfig["contexts"].([]interface{})
	if !ok {
		s.GetLogger().Warnf("contexts in kubeconfig is not an array, skipping context selection")
		return kubeconfig
	}

	var selectedContext = make(map[string]interface{})
	for _, ctx := range contexts {
		ctxMap, ok := ctx.(map[string]interface{})
		if !ok {
			s.GetLogger().Warnf("context %v is not a map, skipping", ctx)
			continue
		}

		if ctxMap["name"] == *s.Config.SelectContext {
			selectedContext = ctxMap
			break
		}
	}

	if selectedContext == nil {
		s.GetLogger().Warnf("selected context %q not found in kubeconfig", *s.Config.SelectContext)
		return kubeconfig
	}

	selectedContext["name"] = name
	kubeconfig["contexts"] = []interface{}{selectedContext}
	return kubeconfig
}

func (s *OTCStore) GetSearchPreview(_ string, tags map[string]string) (string, error) {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !s.IsInitialized() {
		if err := s.InitializeOTCStore(); err != nil {
			return "", fmt.Errorf("failed to initialize OTC store: %w", err)
		}
	}

	asciTree := gotree.New(tags["name"])

	asciTree.Add(fmt.Sprintf("ID: %s", tags["id"]))
	asciTree.Add(fmt.Sprintf("Cloud: %s", *s.Config.Cloud))
	asciTree.Add(fmt.Sprintf("Version: %s", tags["version"]))
	asciTree.Add(fmt.Sprintf("Type: %s", tags["type"]))
	asciTree.Add(fmt.Sprintf("Flavor: %s", tags["flavor"]))
	asciTree.Add(fmt.Sprintf("Status: %s", tags["status"]))

	return asciTree.Print(), nil
}
