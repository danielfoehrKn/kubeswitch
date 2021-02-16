package store

import (
	"github.com/danielfoehrkn/kubectlSwitch/types"
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
	VeryKubeconfigPaths() error
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
