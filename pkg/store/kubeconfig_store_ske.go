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
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	stackitconfig "github.com/stackitcloud/stackit-sdk-go/core/config"
	ske "github.com/stackitcloud/stackit-sdk-go/services/ske" //nolint:staticcheck
	"gopkg.in/yaml.v3"

	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
)

// stackitCLIRoundTripper is an http.RoundTripper that obtains a fresh bearer token
// by invoking the STACKIT CLI (`stackit auth get-access-token`) on every request.
// The CLI manages the token lifecycle (refresh, expiry) transparently, so this
// works as long as the user has previously run `stackit auth login`.
type stackitCLIRoundTripper struct {
	next http.RoundTripper
}

func (rt *stackitCLIRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(req.Context(), "stackit", "auth", "get-access-token")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("stackit auth get-access-token failed: %w: %s", err, stderr.String())
	}
	token := strings.TrimSpace(stdout.String())
	if token == "" {
		return nil, fmt.Errorf("stackit auth get-access-token returned an empty token")
	}

	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+token)

	next := rt.next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(reqClone)
}

const skeDefaultRegion = "eu01"

// NewSKEStore creates a new SKE (STACKIT Kubernetes Engine) store.
func NewSKEStore(store types.KubeconfigStore) (*SKEStore, error) {
	skeStoreConfig := &types.StoreConfigSKE{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to process SKE store config: %w", err)
		}
		if err = yaml.Unmarshal(buf, skeStoreConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal SKE store config: %w", err)
		}
	}

	if skeStoreConfig.ProjectID == "" {
		return nil, fmt.Errorf("when using the SKE kubeconfig store, a STACKIT project ID must be provided via the store config (projectID)")
	}

	if skeStoreConfig.Region == "" {
		skeStoreConfig.Region = skeDefaultRegion
	}

	// Build client options
	var opts []stackitconfig.ConfigurationOption
	if skeStoreConfig.UseStackitCLIAuth {
		// Verify the CLI is available
		if _, err := exec.LookPath("stackit"); err != nil {
			return nil, fmt.Errorf("useStackitCLIAuth is enabled but the 'stackit' CLI was not found in PATH: %w", err)
		}
		opts = append(opts, stackitconfig.WithCustomAuth(&stackitCLIRoundTripper{}))
	} else if skeStoreConfig.ServiceAccountToken != "" {
		opts = append(opts, stackitconfig.WithToken(skeStoreConfig.ServiceAccountToken))
	} else if skeStoreConfig.ServiceAccountKeyPath != "" {
		opts = append(opts, stackitconfig.WithServiceAccountKeyPath(skeStoreConfig.ServiceAccountKeyPath))
	}
	// If none of the above, the SDK will use environment variables or ~/.stackit/credentials.json

	client, err := ske.NewAPIClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize STACKIT SKE client: %w", err)
	}

	logger := logrus.New().WithField("store", types.StoreKindSKE)

	return &SKEStore{
		Logger:          logger,
		KubeconfigStore: store,
		Client:          client,
		Config:          skeStoreConfig,
	}, nil
}

func (s *SKEStore) GetID() string {
	id := "default"
	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}
	return fmt.Sprintf("%s.%s", types.StoreKindSKE, id)
}

func (s *SKEStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	storePrefix := string(types.StoreKindSKE)
	if s.GetStoreConfig().ID != nil {
		storePrefix = *s.GetStoreConfig().ID
	}

	projectLabel := s.Config.ProjectID
	if s.Config.ProjectName != "" {
		projectLabel = s.Config.ProjectName
	}

	return storePrefix + "/" + projectLabel
}

func (s *SKEStore) GetKind() types.StoreKind {
	return types.StoreKindSKE
}

func (s *SKEStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *SKEStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *SKEStore) StartSearch(channel chan storetypes.SearchResult) {
	s.Logger.Debug("SKE: start search")

	ctx := context.Background()
	resp, err := s.Client.ListClustersExecute(ctx, s.Config.ProjectID, s.Config.Region) //nolint:staticcheck
	if err != nil {
		channel <- storetypes.SearchResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("SKE: failed to list clusters for project %q in region %q: %w", s.Config.ProjectID, s.Config.Region, err),
		}
		return
	}

	if resp == nil || !resp.HasItems() {
		s.Logger.Debugf("SKE: no clusters found in project %q region %q", s.Config.ProjectID, s.Config.Region)
		return
	}

	projectLabel := s.Config.ProjectID
	if s.Config.ProjectName != "" {
		projectLabel = s.Config.ProjectName
	}
	s.Logger.Debugf("SKE: using project label %q for display", projectLabel)

	for _, cluster := range resp.GetItems() {
		name := cluster.GetName()
		if name == "" {
			continue
		}
		channel <- storetypes.SearchResult{
			KubeconfigPath: name,
			Error:          nil,
		}
	}
}

func (s *SKEStore) GetKubeconfigForPath(clusterName string, _ map[string]string) ([]byte, error) {
	s.Logger.Debugf("SKE: getting kubeconfig for cluster %q", clusterName)

	ctx := context.Background()
	kubeconfigResp, err := s.Client.CreateKubeconfig(ctx, s.Config.ProjectID, s.Config.Region, clusterName). //nolint:staticcheck
															CreateKubeconfigPayload(*ske.NewCreateKubeconfigPayload()).
															Execute() //nolint:staticcheck
	if err != nil {
		return nil, fmt.Errorf("SKE: failed to create kubeconfig for cluster %q: %w", clusterName, err)
	}

	if kubeconfigResp == nil || !kubeconfigResp.HasKubeconfig() {
		return nil, fmt.Errorf("SKE: received empty kubeconfig for cluster %q", clusterName)
	}

	return []byte(kubeconfigResp.GetKubeconfig()), nil
}

func (s *SKEStore) VerifyKubeconfigPaths() error {
	return nil
}
