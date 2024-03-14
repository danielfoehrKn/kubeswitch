package k8s

import (
	"io/ioutil"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"gopkg.in/yaml.v2"
)

// Kubeconfig represents a kubernetes kubeconfig file
type Kubeconfig struct {
	raw            []byte
	APIVersion     string                       `yaml:"apiVersion"`
	Kind           string                       `yaml:"kind"`
	CurrentContext string                       `yaml:"current-context"`
	Clusters       []*KubeconfigClusterWithName `yaml:"clusters"`
	Contexts       []*KubeconfigContextWithName `yaml:"contexts"`
	Users          []*KubeconfigUserWithName    `yaml:"users"`
}

// KubeconfigUserWithName represents a named cluster in the kubeconfig file
type KubeconfigClusterWithName struct {
	Name    string            `yaml:"name"`
	Cluster KubeconfigCluster `yaml:"cluster"`
}

// KubeconfigCluster represents a cluster in the kubeconfig file
type KubeconfigCluster struct {
	Server                   string `yaml:"server,omitempty"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
}

// KubeconfigContextWithName represents a named context in the kubeconfig file
type KubeconfigContextWithName struct {
	Name    string            `yaml:"name"`
	Context KubeconfigContext `yaml:"context"`
}

// KubeconfigContext represents a context in the kubeconfig file
type KubeconfigContext struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace,omitempty"`
	User      string `yaml:"user"`
}

// KubeconfigUserWithName represents a named user in the kubeconfig file
type KubeconfigUserWithName struct {
	Name string         `yaml:"name"`
	User KubeconfigUser `yaml:"user"`
}

// KubeconfigUser represents a user in the kubeconfig file
type KubeconfigUser struct {
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
	Password              string `yaml:"password,omitempty"`
	Username              string `yaml:"username,omitempty"`
	Token                 string `yaml:"token,omitempty"`
}

// GetRaw returns the raw bytes of the kubeconfig
func (k *Kubeconfig) GetRaw() []byte {
	return k.raw
}

// GetServer returns the server URL of the cluster in the kubeconfig
func (k *Kubeconfig) GetServer() (string, error) {
	if len(k.Clusters) != 1 {
		return "", errors.New("kubeconfig should have only one cluster")
	}

	return k.Clusters[0].Cluster.Server, nil
}

// GetCertificateAuthorityData returns the server certificate authority data of the cluster in the kubeconfig
func (k *Kubeconfig) GetCertificateAuthorityData() (string, error) {
	if len(k.Clusters) != 1 {
		return "", errors.New("kubeconfig should have only one cluster")
	}

	return k.Clusters[0].Cluster.CertificateAuthorityData, nil
}

// GetToken returns the token for the cluster in the kubeconfig
func (k *Kubeconfig) GetToken() (string, error) {
	if len(k.Users) != 1 {
		return "", errors.New("kubeconfig should have only one user")
	}

	return k.Users[0].User.Token, nil
}

// GetClusterKubeConfig downloads the kubeconfig for the given cluster
func (s *API) GetClusterKubeConfig(req *GetClusterKubeConfigRequest, opts ...scw.RequestOption) (*Kubeconfig, error) {
	kubeconfigFile, err := s.getClusterKubeConfig(&GetClusterKubeConfigRequest{
		Region:    req.Region,
		ClusterID: req.ClusterID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting cluster kubeconfig")
	}

	kubeconfigContent, err := ioutil.ReadAll(kubeconfigFile.Content)
	if err != nil {
		return nil, errors.Wrap(err, "error reading kubeconfig content")
	}

	var kubeconfig Kubeconfig
	err = yaml.Unmarshal(kubeconfigContent, &kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling kubeconfig")
	}

	kubeconfig.raw = kubeconfigContent

	return &kubeconfig, nil
}
