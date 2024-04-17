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
	"os"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/smithy-go/logging"
	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/disiqueira/gotree"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func NewEKSStore(store types.KubeconfigStore, stateDir string) (*EKSStore, error) {
	eksStoreConfig := &types.StoreConfigEKS{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(buf, eksStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal eks config: %w", err)
		}

		if eksStoreConfig.Region == nil {
			defaultregion, ok := os.LookupEnv("AWS_DEFAULT_REGION")
			if !ok {
				return nil, fmt.Errorf("failed to set aws region from config or environment")
			}
			eksStoreConfig.Region = &defaultregion
		}
	} else {
		profile, ok := os.LookupEnv("AWS_PROFILE")
		if !ok {
			return nil, fmt.Errorf("failed to set aws profile from config or environment")
		}

		region, ok := os.LookupEnv("AWS_REGION")
		defaultregion, ok2 := os.LookupEnv("AWS_DEFAULT_REGION")
		if !ok && !ok2 {
			return nil, fmt.Errorf("failed to set aws region from config or environment")
		}
		if ok2 {
			region = defaultregion
		}

		eksStoreConfig.Profile = profile
		eksStoreConfig.Region = &region
	}

	if len(eksStoreConfig.Profile) == 0 {
		return nil, fmt.Errorf("profile is required")
	}
	if eksStoreConfig.Region == nil || len(*eksStoreConfig.Region) == 0 {
		return nil, fmt.Errorf("region is required")
	}

	return &EKSStore{
		KubeconfigStore:    store,
		Config:             eksStoreConfig,
		StateDirectory:     stateDir,
		DiscoveredClusters: make(map[string]*awsekstypes.Cluster),
	}, nil
}

func (s *EKSStore) InitializeEKSStore() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithLogger(AWSLogrusBridgeLogger{Logger: s.GetLogger()}),
	}

	optFns = append(optFns, awsconfig.WithRegion(*s.Config.Region))
	optFns = append(optFns, awsconfig.WithSharedConfigProfile(s.Config.Profile))

	cfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return err
	}

	s.Client = awseks.NewFromConfig(cfg)

	return nil
}

func (s *EKSStore) IsInitialized() bool {
	return s.Client != nil && s.Config != nil
}

func (s *EKSStore) GetID() string {
	id := "default"

	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}

	return fmt.Sprintf("%s.%s", types.StoreKindEKS, id)
}

func (s *EKSStore) GetKind() types.StoreKind {
	return types.StoreKindEKS
}

func (s *EKSStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *EKSStore) GetLogger() *logrus.Entry {
	if s.Logger == nil {
		s.Logger = logrus.WithField("store", s.GetID())
	}
	return s.Logger
}

func (s *EKSStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	return strings.ReplaceAll(path, "--", "-")
}

func (s *EKSStore) VerifyKubeconfigPaths() error {
	// NOOP
	return nil
}

func (s *EKSStore) StartSearch(channel chan SearchResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.InitializeEKSStore(); err != nil {
		err := fmt.Errorf("failed to initialize store. This is most likely a problem with your provided aws credentials: %v", err)
		channel <- SearchResult{
			Error: err,
		}
		return
	}

	opts := &awseks.ListClustersInput{}
	pager := awseks.NewListClustersPaginator(s.Client, opts)
	for pager.HasMorePages() {
		s.GetLogger().Debugf("next page found")
		resp, err := pager.NextPage(ctx)
		if err != nil {
			channel <- SearchResult{
				Error: err,
			}
			return
		}

		for _, clusterName := range resp.Clusters {
			// kubeconfig path used to uniquely identify this cluster
			// eks_<profile>--<region>--<eks-cluster-name>
			kubeconfigPath := fmt.Sprintf("eks_%s--%s--%s", s.Config.Profile, *s.Config.Region, clusterName)

			channel <- SearchResult{
				KubeconfigPath: kubeconfigPath,
				Error:          nil,
			}
		}
	}
	s.GetLogger().Debugf("Search done for EKS")
}

// ParseIdentifier takes a kubeconfig identifier and
// returns the
// 1) the EKS resource group
// 2) the name of the EKS cluster
func parseEksIdentifier(path string) (string, string, string, error) {
	split := strings.Split(path, "--")
	switch len(split) {
	case 3:
		return strings.TrimPrefix(split[0], "eks_"), split[1], split[2], nil
	default:
		return "", "", "", fmt.Errorf("unable to parse kubeconfig path: %q", path)
	}
}

func (s *EKSStore) GetKubeconfigForPath(path string, _ map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !s.IsInitialized() {
		if err := s.InitializeEKSStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize EKS store: %w", err)
		}
	}
	_, _, clusterName, err := parseEksIdentifier(path)
	if err != nil {
		return nil, err
	}

	cluster := s.DiscoveredClusters[path]
	if cluster == nil {
		resp, err := s.Client.DescribeCluster(ctx, &awseks.DescribeClusterInput{Name: &clusterName})
		if err != nil {
			return nil, err
		}
		s.DiscoveredClusters[path] = resp.Cluster
		cluster = resp.Cluster
	}

	// context name does not include the location or the account as this information is already included in the path (different to gcloud)
	contextName := fmt.Sprintf("eks_%s", *cluster.Name)

	// need to provide a CA certificate in the kubeconfig (if not using insecure configuration)
	if cluster.CertificateAuthority == nil || cluster.CertificateAuthority.Data == nil {
		return nil, fmt.Errorf("cluster CA certificate not found for cluster=%s", *cluster.Arn)
	}

	kubeconfig := &types.KubeConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Config",
		},
		Clusters: []types.KubeCluster{{
			Name: contextName,
			Cluster: types.Cluster{
				CertificateAuthorityData: *cluster.CertificateAuthority.Data,
				Server:                   *cluster.Endpoint,
			},
		}},
		CurrentContext: contextName,
		Contexts: []types.KubeContext{
			{
				Name: contextName,
				Context: types.Context{
					Cluster: contextName,
					User:    contextName,
				},
			},
		},
		Users: []types.KubeUser{
			{
				Name: contextName,
				User: types.User{
					ExecProvider: &types.ExecProvider{
						APIVersion: "client.authentication.k8s.io/v1beta1",
						Command:    "aws",
						Args: []string{
							"--region",
							*s.Config.Region,
							"eks",
							"get-token",
							"--cluster-name",
							*cluster.Name,
						},
						Env: []types.EnvMap{
							{Name: "AWS_PROFILE", Value: s.Config.Profile},
						},
					},
				},
			},
		},
	}

	bytes, err := yaml.Marshal(kubeconfig)

	return bytes, err
}

func (s *EKSStore) GetSearchPreview(path string, optionalTags map[string]string) (string, error) {
	if !s.IsInitialized() {
		// this takes too long, initialize concurrently
		go func() {
			if err := s.InitializeEKSStore(); err != nil {
				s.Logger.Debugf("failed to initialize store: %v", err)
			}
		}()
		return "", fmt.Errorf("eks store is not initalized yet")
	}

	// low timeout to not pile up many requests, but timeout fast
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	profile, region, clusterName, err := parseEksIdentifier(path)
	if err != nil {
		return "", err
	}

	// the cluster should be in the cache, but do not fail if it is not
	cluster := s.DiscoveredClusters[path]

	// cluster has not been discovered from the EKS API yet
	// this is the case when a search index is used
	if cluster == nil {
		// The name of the cluster to retrieve.
		// we can safely use the client, as we know the store has been previously initialized
		resp, err := s.Client.DescribeCluster(ctx, &awseks.DescribeClusterInput{Name: &clusterName})
		if err != nil {
			return "", fmt.Errorf("failed to get Eks cluster with name %q : %w", clusterName, err)
		}
		cluster = resp.Cluster
		s.DiscoveredClusters[path] = cluster
	}

	asciTree := gotree.New(clusterName)

	if cluster.Version != nil {
		asciTree.Add(fmt.Sprintf("Kubernetes Version: %s", *cluster.Version))
	}
	if cluster.PlatformVersion != nil {
		asciTree.Add(fmt.Sprintf("Platform Version: %s", *cluster.PlatformVersion))
	}

	asciTree.Add(fmt.Sprintf("Status: %s", cluster.Status))
	asciTree.Add(fmt.Sprintf("AWS Profile: %s", profile))
	asciTree.Add(fmt.Sprintf("Region: %s", region))

	return asciTree.Print(), nil
}

// AWSLogrusBridgeLogger is a Logger implementation that wraps the standard library logger, and delegates logging to it's
// Printf method.
type AWSLogrusBridgeLogger struct {
	Logger *logrus.Entry
}

// Logf logs the given classification and message to the underlying logger.
func (s AWSLogrusBridgeLogger) Logf(classification logging.Classification, format string, v ...interface{}) {
	level, err := logrus.ParseLevel(string(classification))
	if err != nil {
		level = logrus.DebugLevel
	}
	s.Logger.Logf(level, format, v...)
}
