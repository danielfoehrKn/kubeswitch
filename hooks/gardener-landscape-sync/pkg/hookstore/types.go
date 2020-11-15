package hookstore

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/danielfoehrkn/kubectlSwitch/types"
)

const (
	KubeconfigStoreFilesystem = "filesystem"
	KubeconfigStoreVault      = "vault"
)

type KubeconfigStore interface {
	GetKind() types.StoreKind
	CreateLandscapeDirectory(dir string) error
	WriteKubeconfigFile(directory, kubeconfigName string, kubeconfigSecret corev1.Secret) error
	CleanExistingKubeconfigs(log *logrus.Entry, dir string) error
}

type FileStore struct {
}

type VaultStore struct {
	Client *vaultapi.Client
}
