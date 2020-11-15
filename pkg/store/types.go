package store

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"

	"github.com/danielfoehrkn/kubectlSwitch/types"
)

type PathDiscoveryResult struct {
	KubeconfigPath string
	Error          error
}

type KubeconfigStore interface {
	GetKind() types.StoreKind
	CheckRootPath() error
	DiscoverPaths(log *logrus.Entry, channel chan PathDiscoveryResult)
	GetKubeconfigForPath(log *logrus.Entry, path string) ([]byte, error)
}

type FilesystemStore struct {
	KubeconfigDirectory string
	KubeconfigName      string
}

type VaultStore struct {
	KubeconfigName              string
	Client                      *vaultapi.Client
	VaultSecretEnginePathPrefix string
}
