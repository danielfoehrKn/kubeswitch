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
	"sync"

	"github.com/danielfoehrkn/kubeswitch/pkg/store/doks"
	"github.com/disiqueira/gotree"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/danielfoehrkn/kubeswitch/types"
	"github.com/digitalocean/doctl/do"
	"github.com/digitalocean/godo"
)

const (
	// tagDOKSClusterID is the tag that contains the remembered DOKS cluster ID in the tag metadata nad is required to obtain the kubeconfig from the DO API
	tagDOKSClusterID = "id"
	// tagDOKSClusterID is the tag that contains the doctl context (account with different access token) in which the DOKS cluster resides
	tagDoctlContextName = "ctx"
	// tagDOKSClusterID is the tag that contains the region of the DOKS cluster
	tagRegion = "region"
	// tagVersion is the tag that contains the K8s version of the DOKS cluster
	tagVersion = "version"
	// tagNodePools is the tag that contains the node pools of the DOKS cluster
	tagNodePools = "pools"
	// tagDOKSClusterID is the tag that contains the non-identifying DOKS cluster name
	tagDOKSClusterName = "name"
)

// NewDigitalOceanStore creates a new DigitalOcean store
func NewDigitalOceanStore(store types.KubeconfigStore) (*DigitalOceanStore, error) {
	doctlConfig, err := doks.GetDoctlConfiguration()
	// as the DO store is enabled by default to provide a seamless experience when already using `doctl`, it is perfectly fine that the doctl config file does not exist (the user might simply not use `doctl`)
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to load doctl config file")
	}

	return &DigitalOceanStore{
		Logger:          logrus.New().WithField("store", types.StoreKindDigitalOcean),
		KubeconfigStore: store,
		Config:          *doctlConfig,
	}, nil
}

// InitializeDigitalOceanStore initializes the DigitalOcean store with digital ocean clients
func (d *DigitalOceanStore) InitializeDigitalOceanStore() error {
	contextToKubernetesService := make(map[string]do.KubernetesService)
	accessToken := d.Config.DefaultAuthContextAccessToken
	defaultContextClient, err := d.getDoClient(accessToken)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to intialize the client for the default digital ocean account/context (context: %s)", d.Config.DefaultContextName))
	}

	contextToKubernetesService[d.Config.DefaultContextName] = do.NewKubernetesService(defaultContextClient)
	d.Logger.Debugf("Created digital ocean client for context: %s", d.Config.DefaultContextName)

	// if there are multiple contexts configured
	for doctlContextName, token := range d.Config.AuthContexts {
		doClient, err := d.getDoClient(token)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to intialize digital ocean client (context: %s)", d.Config.DefaultContextName))
		}

		contextToKubernetesService[doctlContextName] = do.NewKubernetesService(doClient)
		d.Logger.Debugf("Created digital ocean client for context: %s", doctlContextName)
	}
	d.ContextToKubernetesService = contextToKubernetesService
	return nil
}

// getDoClient creates the digital ocean client for a given access token
// inspired by: https://github.com/digitalocean/doctl/blob/7f1c9db38d19cd1104dc96537c00c6436768955a/doit.go#L235
func (d *DigitalOceanStore) getDoClient(accessToken string) (*godo.Client, error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)

	args := []godo.ClientOpt{
		godo.SetUserAgent("kubeswitch-client"),
	}

	if d.Config.HttpRetryMax > 0 {
		retryConfig := godo.RetryConfig{
			RetryMax: d.Config.HttpRetryMax,
		}

		if d.Config.HttpRetryWaitMax > 0 {
			retryConfig.RetryWaitMax = godo.PtrTo(float64(d.Config.HttpRetryWaitMax))
		}

		if d.Config.HttpRetryWaitMin > 0 {
			retryConfig.RetryWaitMin = godo.PtrTo(float64(d.Config.HttpRetryWaitMin))
		}
		args = append(args, godo.WithRetryAndBackoffs(retryConfig))
	}

	if d.Config.ApiUrl != "" {
		args = append(args, godo.SetBaseURL(d.Config.ApiUrl))
	}

	return godo.New(oauthClient, args...)
}

// StartSearch starts the search for Digital Ocean clusters
func (d *DigitalOceanStore) StartSearch(channel chan SearchResult) {
	if err := d.InitializeDigitalOceanStore(); err != nil {
		err := fmt.Errorf("failed to initialize store: %w", err)
		channel <- SearchResult{
			Error: err,
		}
		return
	}

	wgResultChannel := sync.WaitGroup{}
	wgResultChannel.Add(len(d.ContextToKubernetesService))

	for doctlContextName, doSvc := range d.ContextToKubernetesService {
		// parallelize.
		go func(resultChannel chan SearchResult, doctlCtxName string, svc do.KubernetesService) {
			// reading from this context is finished, decrease wait counter
			defer wgResultChannel.Done()

			d.Logger.Debugf("Digital Ocean: Start listing clusters for context %q", doctlCtxName)
			clusters, err := svc.List()
			if err != nil {
				channel <- SearchResult{
					Error: fmt.Errorf("error listing DOKS clusters for context %s: %w", doctlCtxName, err),
				}
				return
			}

			for _, cluster := range clusters {
				d.Logger.Debugf("Digital Ocean: found cluster (context: %s, ID: %s, name: %s, region: %s)", doctlCtxName, cluster.ID, cluster.Name, cluster.RegionSlug)
				kubeconfigPath := getDigitalOceanKubeconfigPath(doctlCtxName, cluster.RegionSlug, cluster.Name)

				nodePools := "["
				for _, pool := range cluster.NodePools {
					nodePools = fmt.Sprintf("%s %s", nodePools, pool.Name)
				}
				nodePools = fmt.Sprintf("%s]", nodePools)

				channel <- SearchResult{
					KubeconfigPath: kubeconfigPath,
					Tags: map[string]string{
						tagDOKSClusterID:    cluster.ID,
						tagDoctlContextName: doctlCtxName,
						tagDOKSClusterName:  cluster.Name,
						tagRegion:           cluster.RegionSlug,
						tagVersion:          cluster.VersionSlug,
						tagNodePools:        nodePools,
					},
					Error: nil,
				}
			}

			d.Logger.Debugf("Digital Ocean: Search done for context %q", doctlCtxName)
		}(
			channel,
			doctlContextName,
			doSvc,
		)
	}

	// wait for all goroutines to finish
	wgResultChannel.Wait()

	d.Logger.Debugf("Digital Ocean: Search done for all contexts")
}

func getDigitalOceanKubeconfigPath(context, region, clusterName string) string {
	// required to be unique for each cluster
	return fmt.Sprintf("do_%s--%s--%s", context, region, clusterName)
}

func (s *DigitalOceanStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	doctlContextName, _, _, err := parseDigitalOceanIdentifier(path)
	if err != nil {
		// fallback and hope that the generated context name is unique
		return "do_"
	}
	// the DigitalOcean store encodes the path with semantic information do_<context-name/account-name>.
	// However, the cluster is NOT identified via the information in the search path unlike in other stores.
	// The metadata tags with the cluster ID is used for this pupose instead. The search path is simply for users to be able to differentiate DOKS clusters.
	return fmt.Sprintf("do_%s", doctlContextName)
}

// IsInitialized checks if the store has been initialized with clients already
func (s *DigitalOceanStore) IsInitialized() bool {
	return len(s.ContextToKubernetesService) > 0
}

func (s *DigitalOceanStore) GetID() string {
	id := "default"

	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}

	return fmt.Sprintf("%s.%s", types.StoreKindDigitalOcean, id)
}

func (s *DigitalOceanStore) GetKind() types.StoreKind {
	return types.StoreKindDigitalOcean
}

func (s *DigitalOceanStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *DigitalOceanStore) GetLogger() *logrus.Entry {
	return s.Logger
}

// GetKubeconfigForPath gets the kubeconfig bytes for the given kubeconfig path and tags
// For this store, instead of using the path to identify the kubeconfig in the backing store, the cluster ID in the tags metadata
// is used. Reason: the clusterID is a long non-intuitive string that we don't want to
func (d *DigitalOceanStore) GetKubeconfigForPath(path string, tags map[string]string) ([]byte, error) {
	if !d.IsInitialized() {
		if err := d.InitializeDigitalOceanStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize Digital Ocean store: %w", err)
		}
	}

	var (
		clusterID        string
		region           string
		name             string
		doctlContextName string
		ok               bool
	)

	// the tags are either set from the initial search or when using an index, are stoerd in the index file itself.
	if clusterID, ok = tags[tagDOKSClusterID]; !ok {
		return nil, fmt.Errorf("failed to GetKubeconfigForPath: %s. Required cluster ID not found in the metadata tags: %v", path, tags)
	}

	if doctlContextName, ok = tags[tagDoctlContextName]; !ok {
		return nil, fmt.Errorf("failed to GetKubeconfigForPath: %s. Required doctl context name not found in the metadata tags: %v", path, tags)
	}

	region = tags[tagRegion]
	name = tags[tagDOKSClusterName]

	d.Logger.Debugf("Digital Ocean: GetKubeconfigForPath (context: %s, region: %s, DOKS cluster name: %s, DOKS cluster ID: %s)", doctlContextName, region, name, clusterID)

	kubeconfigBytes, err := d.ContextToKubernetesService[doctlContextName].GetKubeConfig(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain kubeconfig for DOKS cluster (context: %s, region: %s, DOKS cluster name: %s, cluster_id: %s): %w", doctlContextName, region, name, clusterID, err)
	}

	return kubeconfigBytes, nil
}

func (s *DigitalOceanStore) VerifyKubeconfigPaths() error {
	// NOOP
	return nil
}

// ParseIdentifier takes a kubeconfig identifier and
// returns the
// 1) the `doctl` context name
// 1) the region
// 2) the name of the DOKS cluster
func parseDigitalOceanIdentifier(path string) (string, string, string, error) {
	split := strings.Split(path, "--")
	switch len(split) {
	case 3:
		return strings.TrimPrefix(split[0], "do_"), split[1], split[2], nil
	default:
		return "", "", "", fmt.Errorf("unable to parse kubeconfig path: %q", path)
	}
}

// GetSearchPreview enhances the preview with information stored in the metadata tags (no API requests are being performed)
func (d *DigitalOceanStore) GetSearchPreview(_ string, tags map[string]string) (string, error) {
	asciTree := gotree.New(fmt.Sprintf("DOKS: %s", tags[tagDOKSClusterName]))

	if id, ok := tags[tagDOKSClusterID]; ok {
		asciTree.Add(fmt.Sprintf("ID: %s", id))
	}

	if doctlContext, ok := tags[tagDoctlContextName]; ok {
		asciTree.Add(fmt.Sprintf("doctl context: %s", doctlContext))
	}

	if version, ok := tags[tagVersion]; ok {
		asciTree.Add(fmt.Sprintf("Kubernetes Version: %s", version))
	}

	if region, ok := tags[tagRegion]; ok {
		asciTree.Add(fmt.Sprintf("Region: %s", region))
	}

	if pools, ok := tags[tagNodePools]; ok {
		asciTree.Add(fmt.Sprintf("Node Pools: %s", pools))
	}

	return asciTree.Print(), nil
}
