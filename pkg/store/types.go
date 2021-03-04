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
)

type SearchResult struct {
	KubeconfigPath string
	Error          error
}

type KubeconfigStore interface {
	GetLogger() *logrus.Entry
	GetKind() types.StoreKind
	VerifyKubeconfigPaths() error
	StartSearch(channel chan SearchResult)
	GetKubeconfigForPath(path string) ([]byte, error)
}

type FilesystemStore struct {
	Logger                *logrus.Entry
	KubeconfigPaths       []types.KubeconfigPath
	KubeconfigName        string
	kubeconfigDirectories []string
	kubeconfigFilepaths   []string
}

type VaultStore struct {
	Logger          *logrus.Entry
	Client          *vaultapi.Client
	KubeconfigName  string
	KubeconfigPaths []types.KubeconfigPath
	vaultPaths      []string
}
