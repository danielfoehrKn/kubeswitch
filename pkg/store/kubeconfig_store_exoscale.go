package store

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
)

func NewExoscaleStore(store types.KubeconfigStore) (*ExoscaleStore, error) {
	exoscaleStoreConfig := &types.StoreConfigExoscale{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to process Exoscale store config: %w", err)
		}

		err = yaml.Unmarshal(buf, exoscaleStoreConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal Exoscale config: %w", err)
		}
	}

	logger := logrus.New().WithField("store", types.StoreKindExoscale)

	exoscaleAPIKey := exoscaleStoreConfig.ExoscaleAPIKey
	if len(exoscaleAPIKey) == 0 {
		return nil, fmt.Errorf("when using the Exoscale kubeconfig store, the API key for Exoscale has to be provided via a SwitchConfig file")
	}

	exoscaleSecretKey := exoscaleStoreConfig.ExoscaleSecretKey
	if len(exoscaleSecretKey) == 0 {
		return nil, fmt.Errorf("when using the Exoscale kubeconfig store, the secret key for Exoscale has to be provided via a SwitchConfig file")
	}

	creds := credentials.NewStaticCredentials(exoscaleAPIKey, exoscaleSecretKey)
	client, err := v3.NewClient(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Exoscale client due to error: %w", err)
	}

	return &ExoscaleStore{
		Logger:             logger,
		KubeconfigStore:    store,
		Client:             client,
		DiscoveredClusters: make(map[v3.UUID]ExoscaleKube),
	}, nil
}

// ExoscaleKube represents one discovered SKS cluster, including zone info.
type ExoscaleKube struct {
	ID           v3.UUID
	Name         string
	ZoneName     v3.ZoneName // e.g. "ch-gva-2", "de-fra-1"
	ZoneEndpoint v3.Endpoint // e.g. "https://api-ch-gva-2.exoscale.com/v2"
}

func (s *ExoscaleStore) GetID() string {
	id := "default"
	if s.KubeconfigStore.ID != nil {
		id = *s.KubeconfigStore.ID
	}
	return fmt.Sprintf("%s.%s", types.StoreKindExoscale, id)
}

func (s *ExoscaleStore) GetContextPrefix(path string) string {
	if s.GetStoreConfig().ShowPrefix != nil && !*s.GetStoreConfig().ShowPrefix {
		return ""
	}

	if s.GetStoreConfig().ID != nil {
		return *s.GetStoreConfig().ID
	}

	return string(types.StoreKindExoscale)
}

func (s *ExoscaleStore) GetKind() types.StoreKind {
	return types.StoreKindExoscale
}

func (s *ExoscaleStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *ExoscaleStore) GetLogger() *logrus.Entry {
	return s.Logger
}

// StartSearch queries *all* Exoscale zones, discovers SKS clusters
// in each zone, and publishes the cluster names prefixed with <zoneName>/.
func (s *ExoscaleStore) StartSearch(channel chan storetypes.SearchResult) {
	s.Logger.Debug("Exoscale: start search")

	ctx := context.Background()

	// 1. List all zones
	zonesResp, err := s.Client.ListZones(ctx)
	if err != nil {
		channel <- storetypes.SearchResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("failed to list zones: %w", err),
		}
		return
	}
	if len(zonesResp.Zones) == 0 {
		s.Logger.Debug("No Exoscale zones found")
		return
	}

	// 2. For each zone, list SKS clusters
	for _, zone := range zonesResp.Zones {
		zoneClient := s.Client.WithEndpoint(zone.APIEndpoint)

		clusters, err := zoneClient.ListSKSClusters(ctx)
		if err != nil {
			// If a single zone fails, report it but continue with the others
			s.Logger.WithError(err).Warnf("Failed to list SKS clusters for zone %s", zone.Name)
			continue
		}

		if len(clusters.SKSClusters) == 0 {
			s.Logger.Debugf("No SKS clusters found in zone %s", zone.Name)
			continue
		}

		for _, cluster := range clusters.SKSClusters {
			// Record the cluster in memory
			s.DiscoveredClusters[cluster.ID] = ExoscaleKube{
				ID:           cluster.ID,
				Name:         cluster.Name,
				ZoneName:     zone.Name,
				ZoneEndpoint: zone.APIEndpoint,
			}

			s.Logger.Debugf("Discovered SKS cluster name: %s and id: %s in zone %s", cluster.Name, cluster.ID, zone.Name)

			// The path we present back to kubeswitch:
			// e.g. "ch-gva-2/my-cluster"
			kubeconfigPath := fmt.Sprintf("%s/%s", zone.Name, cluster.Name)

			// Send the discovered path
			channel <- storetypes.SearchResult{
				KubeconfigPath: kubeconfigPath,
				Error:          nil,
			}
		}
	}
}

// GetKubeconfigForPath expects path like "zoneName/clusterName",
// finds that cluster, and returns the decoded YAML kubeconfig.
func (s *ExoscaleStore) GetKubeconfigForPath(path string, _ map[string]string) ([]byte, error) {
	// Split path into zoneName/clusterName
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cluster path %q (expected 'zoneName/clusterName')", path)
	}
	zoneName := parts[0]
	clusterName := parts[1]

	// Find the stored cluster that matches both zone and cluster name
	var match *ExoscaleKube
	for _, c := range s.DiscoveredClusters {
		if string(c.ZoneName) == zoneName && c.Name == clusterName {
			match = &c
			break
		}
	}
	if match == nil {
		return nil, fmt.Errorf("no cluster found for %q", path)
	}

	// Prepare client targeting the cluster's zone
	zoneClient := s.Client.WithEndpoint(match.ZoneEndpoint)

	req := v3.SKSKubeconfigRequest{
		Groups: []string{"system:masters"},
		User:   "default",
		Ttl:    2592000, // 30 days
	}
	ctx := context.Background()

	resp, err := zoneClient.GenerateSKSClusterKubeconfig(ctx, match.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kubeconfig for cluster %q: %w", path, err)
	}

	// The returned kubeconfig is base64-encoded YAML
	rawKubeconfig, err := base64.StdEncoding.DecodeString(resp.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 kubeconfig for cluster %q: %w", path, err)
	}

	return rawKubeconfig, nil
}

func (r *ExoscaleStore) VerifyKubeconfigPaths() error {
	return nil
}
