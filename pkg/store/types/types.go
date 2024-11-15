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

package types

import (
	"github.com/danielfoehrkn/kubeswitch/types"

	"github.com/sirupsen/logrus"
)

// SearchResult is a full kubeconfig path discovered from the kubeconfig store
// given the contained kubeconfig path, the store knows how to retrieve and return the
// actual kubeconfig
type SearchResult struct {
	// KubeconfigPath is the kubeconfig path in the backing store which most of the time encodes enough information to
	// retrieve the kubeconfig associated with it.
	KubeconfigPath string
	// Tags contains the additional metadata that the store wants to associate with a context name.
	// This metadata is later handed over in the getKubeconfigForPath() function when retrieving the kubeconfig bytes for the path and might contain
	// information necessary to retrieve the kubeconfig from the backing store (such a unique ID for the cluster required for the API)
	Tags map[string]string
	// Error is an error which occured when trying to discover kubeconfig paths in the backing store
	Error error
}

type KubeconfigStore interface {
	// GetID returns the unique store ID
	// should be
	// - "<store kind>.default" if the kubeconfigStore.ID is not set
	// - "<store kind>.<id>" if the kubeconfigStore.ID is set
	GetID() string

	// GetKind returns the store kind (e.g., filesystem)
	GetKind() types.StoreKind

	// GetContextPrefix returns the prefix for the kubeconfig context names displayed in the search result
	// includes the path to the kubeconfig in the backing store because some stores compute the prefix based on that
	GetContextPrefix(path string) string

	// VerifyKubeconfigPaths verifies that the configured search paths are valid
	// can also include additional preprocessing
	VerifyKubeconfigPaths() error

	// StartSearch starts the search over the configured search paths
	// and populates the results via the given channel
	StartSearch(channel chan SearchResult)

	// GetKubeconfigForPath returns the byte representation of the kubeconfig
	// the kubeconfig has to fetch the kubeconfig from its backing store (e.g., uses the HTTP API)
	// Optional tags might help identify the cluster in the backing store, but typically such information is already encoded in the kubeconfig path (implementation specific)
	GetKubeconfigForPath(path string, tags map[string]string) ([]byte, error)

	// GetLogger returns the logger of the store
	GetLogger() *logrus.Entry

	// GetStoreConfig returns the store's configuration from the switch config file
	GetStoreConfig() types.KubeconfigStore
}

// Previewer can be optionally implemented by stores to show custom preview content
// before the kubeconfig
type Previewer interface {
	GetSearchPreview(path string, optionalTags map[string]string) (string, error)
}
