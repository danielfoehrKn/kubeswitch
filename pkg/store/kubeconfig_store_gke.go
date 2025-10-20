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

	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
	"google.golang.org/api/cloudresourcemanager/v1"
)

var (
	scheme           = runtime.NewScheme()
	gcloudBinaryPath = ""
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

	binaryPath, err := getGcloudBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("gcloud must be installaed when useing the GKE store: %v", err)
	}
	gcloudBinaryPath = binaryPath

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
		// this can happen when there are no application-default credentials available on the local disk
		// try to re-authenticate using local gcloud installation
		if len(gcloudBinaryPath) == 0 {
			return fmt.Errorf("failed to create Google Kubernetes Engine client. Please check that the `gcloud` CLI is installed and `gcloud auth application-default login` has run: %w", err)
		}

		// gcloud auth application-default login
		_, err_exec := exec.Command(gcloudBinaryPath, "auth", "application-default", "login").Output()
		if err_exec != nil {
			return fmt.Errorf("failed to acquire missing credentials via gcloud: %v: Failed to create client: %v", err_exec, err)
		}

		s.Logger.Infof("Sucessfully obtained application default credentials.")

		// try again with obtained credentials
		client, err = container.NewService(ctx)
		if err != nil {
			return fmt.Errorf("failed to create Google Kubernetes Engine client: %w", err)
		}
	}

	if s.Config.GCPAccount != nil {
		isActive, err := isAccountActive(*s.Config.GCPAccount)
		if err != nil {
			return fmt.Errorf("failed to check if Google Cloud account %q is active: %w", *s.Config.GCPAccount, err)
		}

		if !isActive {
			return fmt.Errorf("google cloud account %q is not active. Please use `gcloud config set account %s` to activate the account", *s.Config.GCPAccount, *s.Config.GCPAccount)
		}
	}

	s.GkeClient = client

	// Discover projects in this account
	allowedProjectIDs := sets.NewString(s.Config.ProjectIDs...)

	cloudResourceManagerService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create cloud resource manager client: %w", err)
	}

	req := cloudResourceManagerService.Projects.List()
	if err := req.Pages(ctx, func(page *cloudresourcemanager.ListProjectsResponse) error {
		for _, project := range page.Projects {
			if allowedProjectIDs.Len() > 0 && !allowedProjectIDs.Has(project.ProjectId) {
				continue
			}
			// remember project name -> project ID
			s.ProjectNameToID[project.Name] = project.ProjectId
		}
		return nil
	}); err != nil {
		// this might happen when the JWT token (id token) from Googles OIDC provider has expired
		// so the actual request against the API returns 401
		// Try to re-authenticate using gcloud!
		if len(gcloudBinaryPath) == 0 {
			return fmt.Errorf("failed to list Google cloud projects. This indicates either connectivity issues or invalid credentials. Make sure you are connected to the internet and that the `gcloud` CLI is installed for authentication. (Try running: `gcloud auth application-default login`): %w", err)
		}

		// gcloud auth application-default login
		_, errExec := exec.Command(gcloudBinaryPath, "auth", "application-default", "login").Output()
		if errExec != nil {
			return fmt.Errorf("failed to list Google Cloud projects probably due to permission issues. Also failed to acquire application-default credentials via gcloud OIDC authentication flow: %v: %v", errExec, err)
		}

		s.Logger.Infof("Sucessfully obtained application default credentials.")
	}

	if len(s.ProjectNameToID) == 0 {
		return fmt.Errorf("no projects found in Google Cloud. Unable to discover GKE clusters")
	}

	return nil
}

func (s *GKEStore) StartSearch(channel chan storetypes.SearchResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.InitializeGKEStore(); err != nil {
		err := fmt.Errorf("failed to initialize store: %w", err)
		channel <- storetypes.SearchResult{
			Error: err,
		}
		return
	}

	for projectName, projectId := range s.ProjectNameToID {
		resp, err := s.GkeClient.Projects.Locations.Clusters.List(fmt.Sprintf("projects/%s/locations/-", projectId)).Context(ctx).Do()
		if err != nil {
			channel <- storetypes.SearchResult{
				Error: fmt.Errorf("failed to list GKE clusters for project with ID %q: %w", projectId, err),
			}
			continue
		}

		// for every GKE cluster in the project
		for _, f := range resp.Clusters {
			// kubeconfig path used to uniquely identify this cluster
			// gke_<project-name>--<zone>--<gke-cluster-name>

			kubeconfigPath := fmt.Sprintf("gke_%s--%s--%s", projectName, f.Location, f.Name)

			// cache for when getting the kubeconfig for the unique path later
			s.DiscoveredClusters[kubeconfigPath] = f

			channel <- storetypes.SearchResult{
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
	// <project-name>--<location>--<cluster-name>
	// just use this semantic information as a prefix & remove the double dashes
	return strings.ReplaceAll(path, "--", "-")
}

// IsInitialized checks if the store has been initialized already
func (s *GKEStore) IsInitialized() bool {
	return s.GkeClient != nil && s.Config != nil
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

func (s *GKEStore) GetKubeconfigForPath(path string, _ map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !s.IsInitialized() {
		if err := s.InitializeGKEStore(); err != nil {
			return nil, fmt.Errorf("failed to initialize GKE store: %w", err)
		}
	}
	projectName, location, clusterName, err := parseIdentifier(path)
	if err != nil {
		return nil, err
	}

	projectID := s.ProjectNameToID[strings.TrimPrefix(projectName, "gke_")]

	cluster := s.DiscoveredClusters[path]

	// cluster has not been discovered from the GCP API yet
	// this is the case when a search index is used
	if cluster == nil {
		// The name (project, location, cluster) of the cluster to retrieve.
		// Specified in the format 'projects/*/locations/*/clusters/*'.
		name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, location, clusterName)
		resp, err := s.GkeClient.Projects.Locations.Clusters.Get(name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to get GKE cluster with name %q for project with ID %q: %w", clusterName, projectID, err)
		}
		cluster = resp
	}

	// context name does not include the location or the account as this information is already included in the path (different to gcloud)
	contextName := fmt.Sprintf("gke_%s", cluster.Name)

	if cluster.MasterAuth == nil {
		return nil, fmt.Errorf("no authentication information found for GKE cluster with name %q in project with ID %q", clusterName, projectID)
	}

	// need to provide a CA certificate in the kubeconfig (if not using insecure configuration)
	if len(cluster.MasterAuth.ClusterCaCertificate) == 0 {
		return nil, fmt.Errorf("cluster CA certificate not found for cluster=%s in project with ID %q", contextName, projectID)
	}

	var args []string

	// supply authentication information based on the configured auth option
	if s.Config.GKEAuthentication == nil || *s.Config.GKEAuthentication.AuthenticationType == types.GcloudAuthentication {
		args = nil
	} else if s.Config.GKEAuthentication != nil && *s.Config.GKEAuthentication.AuthenticationType == types.ServiceAccountAuthentication {
		// using service accounts, the kubeconfig does not contain any client-credentials
		// Instead, the switch.sh script has to set the env variable GOOGLE_APPLICATION_CREDENTIALS=path/to/gsa-key.json
		// on the shell session.
		// This way, the gcp auth provider called by kubectl can discover the credentials via the env variable.
		// see: https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication#environments-without-gcloud
		args = []string{
			"--use_application_default_credentials",
		}
	}

	var endpoint string
	var certificate string

	certificate = cluster.MasterAuth.ClusterCaCertificate

	if s.Config.PreferredEndpoint == nil {
		endpoint = cluster.Endpoint
	} else {
		switch *s.Config.PreferredEndpoint {
		case types.GkeDnsEndpoint:
			endpoint = cluster.ControlPlaneEndpointsConfig.DnsEndpointConfig.Endpoint
			// DNS Endpoint certificate is not signed this Certificate Authority
			certificate = ""
		case types.GkePrivateEndpoint:
			endpoint = cluster.ControlPlaneEndpointsConfig.IpEndpointsConfig.PrivateEndpoint
		case types.GkePublicEndpoint:
			endpoint = cluster.ControlPlaneEndpointsConfig.IpEndpointsConfig.PublicEndpoint
		default:
			endpoint = cluster.Endpoint
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
				CertificateAuthorityData: certificate,
				Server:                   fmt.Sprintf("https://%s", endpoint),
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
						APIVersion:         "client.authentication.k8s.io/v1beta1",
						Args:               args,
						Command:            "gke-gcloud-auth-plugin",
						InstallHint:        "Install gke-gcloud-auth-plugin for use with kubectl by following\nhttps://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke",
						ProvideClusterInfo: true,
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
// 1) the GCP project name
// 2) the location (zone or region if regional cluster) of the GKE cluster
// 3) the name of the GKE cluster
func parseIdentifier(path string) (string, string, string, error) {
	split := strings.Split(path, "--")
	switch len(split) {
	case 3:
		return split[0], split[1], split[2], nil
	default:
		return "", "", "", fmt.Errorf("unable to parse kubeconfig path: %q", path)
	}
}

// isAccountActive checks if the given GCP account is active
func isAccountActive(targetAccount string) (bool, error) {
	// gcloud auth application-default login
	result, err := exec.Command(gcloudBinaryPath, "auth", "list", "--format", " json").Output()
	if err != nil {
		return false, fmt.Errorf("failed to shell out to gcloud: %w", err)
	}

	accounts := []gcloudAccount{}
	err = yaml.Unmarshal(result, &accounts)
	if err != nil {
		return false, err
	}

	if len(accounts) == 0 {
		return false, fmt.Errorf("no accounts configured for GCP. This can be verified by executing `gcloud auth list`")
	}

	accountFound := false
	for _, account := range accounts {
		if account.Account != targetAccount {
			continue
		}
		accountFound = true
		if account.Status == "ACTIVE" {
			return true, nil
		}
	}

	if !accountFound {
		return false, fmt.Errorf("GCP account %q not found. This can be verified by executing `gcloud auth list`", targetAccount)
	}

	return false, nil
}

type gcloudAccount struct {
	Account string `json:"account"`
	Status  string `json:"status"`
}
