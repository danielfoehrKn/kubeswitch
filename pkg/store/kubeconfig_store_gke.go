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
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/container/v1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/danielfoehrkn/kubeswitch/types"
	"google.golang.org/api/cloudresourcemanager/v1"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(apiv1.AddToScheme(scheme))
}

// NewGKEStore creates a new GKE store
func NewGKEStore(store types.KubeconfigStore, stateDir string) (*GKEStore, error) {
	gkeStoreConfig := &types.StoreConfigGKE{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(buf, gkeStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal gke config: %w", err)
		}
	}

	// TODO: if using gcloud authentication: Check that gcloud is installed e.g via which gcloud

	// TODO: If using gcloud with config specifying the gcp account
	// validate by invoking gcloud auth list --format json that the correct account is ACTIVE

	return &GKEStore{
		Logger:             logrus.New().WithField("store", types.StoreKindGKE),
		KubeconfigStore:    store,
		Config:             gkeStoreConfig,
		StateDirectory:     stateDir,
		ProjectNameToID:    map[string]string{},
		DiscoveredClusters: map[string]*container.Cluster{},
	}, nil
}

// InitializeGKEStore initializes the store by listing all available projects for the Google Account
// Decoupled from the NewGKEStore() to be called when starting the search to reduce
// time when the CLI can start showing the fuzzy search
func (s *GKEStore) InitializeGKEStore() error {
	ctx := context.Background()

	// TODO: how to get the account info to validate the account
	// gcloud config config-helper
	// configuration:
	//  active_configuration: default
	//  properties:
	//    compute:
	//      region: europe-west1
	//      zone: europe-west1-b
	//    core:
	//      account: daniel.fit95gke@gmail.com
	//      disable_usage_reporting: 'True'
	//      project: sap-se-gcp-scp-k8s-dev

	// TODO: IF use service accounts:
	// Here set the GOOGLE_APPLICATION_CREDENTIALS env variable
	// otherwise default credentials will not be found

	// TODO: detect gcloud installation path
	// or default to "/usr/local/bin/gcloud" if not otherwise specified in configuration file

	// TODO when using gcloud: check if file exists $HOME/.config/gcloud/application_default_credentials.json
	// also check if access token is expired/ If yes: execute gcloud auth application-default login automatically
	// prerequisite: need to know the binary path of gcloud

	// Create GKE client
	// Google Application Default Credentials are used for authentication.
	// When using gcloud  'gcloud auth application-default login' so that
	// the library can find a valid access token provided via gcloud's oauth flow at the default location
	// cat $HOME/.config/gcloud/application_default_credentials.json

	// Later, also support API keys provided with the store configuration
	// please see: https://pkg.go.dev/google.golang.org/api/container/v1
	// and: https://cloud.google.com/docs/authentication/production#automatically
	client, err := container.NewService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GKE client: %w", err)
	}
	s.GkeClient = client

	// Discover projects in this account
	projects := sets.NewString(s.Config.ProjectIDs...)

	cloudResourceManagerService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create cloud resource manager client: %w", err)
	}

	req := cloudResourceManagerService.Projects.List()
	if err := req.Pages(ctx, func(page *cloudresourcemanager.ListProjectsResponse) error {
		for _, project := range page.Projects {
			if projects.Len() > 0 && !projects.Has(project.ProjectId) {
				continue
			}
			// remember project name -> project ID
			s.ProjectNameToID[project.Name] = project.ProjectId
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *GKEStore) StartSearch(channel chan SearchResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.InitializeGKEStore(); err != nil {
		err := fmt.Errorf("failed to initialize store. Make sure you provided valid credentials or run `gcloud auth application-default login` when using authentication via gcloud: %v", err)
		channel <- SearchResult{
			Error: err,
		}
		return
	}

	for projectName, projectId := range s.ProjectNameToID {
		resp, err := s.GkeClient.Projects.Zones.Clusters.List(projectId, "-").Context(ctx).Do()
		if err != nil {
			channel <- SearchResult{
				Error: fmt.Errorf("failed to list GKE clusters for project with ID %q: %w", projectId, err),
			}
			return
		}

		// for every GKE cluster in the project
		for _, f := range resp.Clusters {
			var landscapeName string
			if len(s.LandscapeName) > 0 {
				landscapeName = fmt.Sprintf("%s--", s.LandscapeName)
			}

			// kubeconfig path used to uniquely identify this cluster
			kubeconfigPath := fmt.Sprintf("%s%s--%s", landscapeName, projectName, f.Name)

			// cache for when getting the kubeconfig for the unique path later
			s.DiscoveredClusters[kubeconfigPath] = f

			channel <- SearchResult{
				KubeconfigPath: kubeconfigPath,
				Error:          nil,
			}

		}
	}
}

func (s *GKEStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	// the GKE store encodes the path with semantic information
	// <optionalLandscapeName>-<project-name>--<cluster-name>
	// just use this semantic information as a prefix & remove the double dashes
	return strings.ReplaceAll(path, "--", "-")
}

// IsInitialized checks if the store has been initialized already
func (s *GKEStore) IsInitialized() bool {
	return s.GkeClient != nil && s.Config != nil && len(s.Config.ProjectIDs) > 0
}

func (s *GKEStore) GetID() string {
	id := "default"

	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}

	return fmt.Sprintf("%s.%s", types.StoreKindGKE, id)
}

func (s *GKEStore) GetKind() types.StoreKind {
	return types.StoreKindGKE
}

func (s *GKEStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *GKEStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *GKEStore) GetKubeconfigForPath(path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !s.IsInitialized() {
		if err := s.InitializeGKEStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize GKE store: %w", err)
		}
	}
	_, projectName, clusterName, err := parseIdentifier(path)
	if err != nil {
		return nil, err
	}

	projectID := s.ProjectNameToID[projectName]

	cluster := s.DiscoveredClusters[path]

	// cluster has not been discovered from the GCP API yet
	// this is the case when a search index is used
	if cluster == nil {
		// get the cluster from the GCP API
		// TODO: does not work. I need the exact zone here :(
		// how do I remember the zone. Shall I put it in the path? :D
		resp, err := s.GkeClient.Projects.Zones.Clusters.Get(projectID, "-", clusterName).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to get GKE cluster with name %q for project with ID %q: %w", clusterName, projectID, err)
		}
		cluster = resp
	}

	// build context name from cluster information
	contextName := fmt.Sprintf("gke_%s_%s_%s", projectID, cluster.Zone, cluster.Name)

	if cluster.MasterAuth == nil {
		return nil, fmt.Errorf("no authentication information found for GKE cluster with name %q in project with ID %q", clusterName, projectID)
	}

	// need to provide a CA certificate in the kubeconfig (if not using insecure configuration)
	if len(cluster.MasterAuth.ClusterCaCertificate) == 0 {
		return nil, fmt.Errorf("cluster CA certificate not found for cluster=%s in project with ID %q", contextName, projectID)
	}

	authPluginConfig := make(map[string]string)

	// supply authentication information based on the configured auth option
	if s.Config.GKEAuthentication == nil || *s.Config.GKEAuthentication.AuthenticationType == types.GcloudAuthentication {
		gcloudBinaryPath, err := getGcloudBinaryPath()
		if err != nil {
			return nil, err
		}

		// construct an AuthInfo that contains the same information if I would have uses `gcloud container clusters get-credentials`
		authPluginConfig = map[string]string{
			// "access-token": token.AccessToken,
			// "expiry": token.Expiry.Format(time.RFC3339), // make sure has proper format
			"cmd-path": gcloudBinaryPath,                     // TODO: if does not work, I need to detect the gcloud install directory
			"cmd-args": "config config-helper --format=json", // get the credentials
			// "expiry-key": token.Expiry.Format(time.RFC3339),
			"expiry-key": "{.credential.token_expiry}",
			// "token-key": token.AccessToken,
			"token-key": "{.credential.access_token}",
		}

	} else if s.Config.GKEAuthentication != nil && *s.Config.GKEAuthentication.AuthenticationType == types.ServiceAccountAuthentication {
		// using service accounts, the kubeconfig does not contain any client-credentials
		// Instead, the the switch.sh script has to set the env variable GOOGLE_APPLICATION_CREDENTIALS=path/to/gsa-key.json
		// on the shell session.
		// This way, the gcp auth provider called by kubectl can discover the credentials via the env variable.
		// see: https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication#environments-without-gcloud
		authPluginConfig = map[string]string{
			"scopes": "https://www.googleapis.com/auth/cloud-platform",
		}
	}

	kubeconfig := &types.KubeConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Config",
		},
		Clusters: []types.KubeCluster{{
			Name: contextName,
			Cluster: types.Cluster{
				CertificateAuthorityData: cluster.MasterAuth.ClusterCaCertificate,
				Server:                   fmt.Sprintf("https://%s", cluster.Endpoint),
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

					AuthProvider: types.AuthProvider{
						Name:   "gcp",
						Config: authPluginConfig,
					},
				},
			},
		},
	}

	bytes, err := yaml.Marshal(kubeconfig)

	return bytes, err
}

// getGcloudBinaryPath tries to lookup the gcloud binary path
func getGcloudBinaryPath() (string, error) {
	path, err := exec.LookPath("gcloud")
	if err != nil {
		return "", fmt.Errorf("unable to find gcloud on the system. Is it installed?: %v", err)
	}
	return path, nil
}

func (s *GKEStore) VerifyKubeconfigPaths() error {
	// NOOP
	return nil
}

// ParseIdentifier takes a kubeconfig identifier and
// returns the
// 1) the optional landscape name
// 2) the GCP project name
// 3) name of the GKE cluster
func parseIdentifier(path string) (*string, string, string, error) {
	split := strings.Split(path, "--")
	switch len(split) {
	case 2:
		return nil, split[0], split[1], nil
	case 3:
		return &split[0], split[1], split[2], nil
	default:
		return nil, "", "", fmt.Errorf("unable to parse kubeconfig path: %q", path)
	}
}
