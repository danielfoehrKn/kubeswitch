package types


type KubeUser struct {
	Name string `yaml:"name"`
	User struct {
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
		Token    string `yaml:"token,omitempty"`
		ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
		ClientKeyData string `yaml:"client-key-data,omitempty"`
	} `yaml:"user"`
}

type KubeCluster struct {
	Name    string `yaml:"name"`
	Cluster struct {
		CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
		Server                   string `yaml:"server"`
		Insecure 				 bool   `yaml:"insecure-skip-tls-verify,omitempty"`
	} `yaml:"cluster"`
}

type KubeConfig struct {
	TypeMeta TypeMeta `yaml:",inline"`
	CurrentContext string `yaml:"current-context"`
	Contexts []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string
		} `yaml:"context"`
	} `yaml:"contexts"`

	Clusters []KubeCluster `yaml:"clusters"`

	Users []KubeUser `yaml:"users"`
}

type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	Kind string `yaml:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
	// +optional
	APIVersion string `yaml:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}
