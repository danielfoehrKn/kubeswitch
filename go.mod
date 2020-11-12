module github.com/danielfoehrkn/kubectlSwitch

go 1.12

require (
	github.com/gardener/gardener v1.11.3
	github.com/hashicorp/vault/api v1.0.4
	github.com/karrick/godirwalk v1.16.1
	github.com/ktr0731/go-fuzzyfinder v0.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.6.3
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
