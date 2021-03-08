// Copyright 2021 Daniel Foehr
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
	"github.com/danielfoehrkn/kubeswitch/types"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
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
	Client          *vaultapi.Client
	KubeconfigName  string
	KubeconfigStore types.KubeconfigStore
	vaultPaths      []string
}

type GardenerStore struct {
	Logger            *logrus.Entry
	Client            client.Client
	KubeconfigStore   types.KubeconfigStore
	Config            *types.StoreConfigGardener
	LandscapeIdentity string
	LandscapeName     string
	StateDirectory    string
	// if a search against the Gardener API has been executed, this is filled with
	// all the Shoot secrets.
	// This way we can save some requests against the API when getting the kubeconfig later
	ShootNameToKubeconfigSecret map[string]corev1.Secret
}
