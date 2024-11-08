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
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/linode/linodego"
	"github.com/sirupsen/logrus"
)

func NewAkamaiStore(store types.KubeconfigStore) (*AkamaiStore, error) {
	akamaiStoreConfig := &types.StoreConfigAkamai{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to process akamai store config: %w", err)
		}

		err = yaml.Unmarshal(buf, akamaiStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal akami config: %w", err)
		}
	}

	return &AkamaiStore{
		Logger:          logrus.New().WithField("store", types.StoreKindAkamai),
		KubeconfigStore: store,
		Config:          akamaiStoreConfig,
	}, nil
}

// InitializeAkamaiStore the Akamai client
func (s *AkamaiStore) InitializeAkamaiStore() error {
	// use environment variables if token is not set
	token := s.Config.LinodeToken
	if token == "" {
		envToken, ok := os.LookupEnv("LINODE_TOKEN")
		if !ok {
			return fmt.Errorf("linode token not set")
		}
		token = envToken
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linodeClient := linodego.NewClient(oauth2Client)

	s.Client = &linodeClient

	return nil
}

// GetID returns the unique store ID
func (s *AkamaiStore) GetID() string {
	return fmt.Sprintf("%s.default", s.GetKind())
}

func (s *AkamaiStore) GetKind() types.StoreKind {
	return types.StoreKindAkamai
}

func (s *AkamaiStore) GetContextPrefix(path string) string {
	return fmt.Sprintf("%s/%s", s.GetKind(), path)
}

func (s *AkamaiStore) VerifyKubeconfigPaths() error {
	// NOOP
	return nil
}

func (s *AkamaiStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *AkamaiStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *AkamaiStore) StartSearch(channel chan SearchResult) {
	s.Logger.Debug("Akamai: start search")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.InitializeAkamaiStore(); err != nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}

	// list linode instances
	instances, err := s.Client.ListLKEClusters(ctx, nil)
	if err != nil {
		channel <- SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}

	for _, instance := range instances {
		channel <- SearchResult{
			KubeconfigPath: instance.Label,
			Tags: map[string]string{
				"clusterID": strconv.Itoa(instance.ID),
				"region":    instance.Region,
			},
		}
	}
}

func (s *AkamaiStore) GetKubeconfigForPath(path string, tags map[string]string) ([]byte, error) {
	s.Logger.Debugf("Akamai: get kubeconfig for path %s", path)

	// initialize client
	if err := s.InitializeAkamaiStore(); err != nil {
		return nil, err
	}

	clusterID, err := strconv.Atoi(tags["clusterID"])
	if err != nil {
		return nil, fmt.Errorf("failed to get clusterID: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// get kubeconfig
	LKEkubeconfig, err := s.Client.GetLKEClusterKubeconfig(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// decode base64 kubeconfig
	kubeconfig, err := base64.StdEncoding.DecodeString(LKEkubeconfig.KubeConfig)
	if err != nil {
		return nil, err
	}

	return kubeconfig, nil
}
