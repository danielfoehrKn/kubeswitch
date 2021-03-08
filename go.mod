module github.com/danielfoehrkn/kubeswitch

go 1.12

require (
	github.com/disiqueira/gotree v1.0.0
	github.com/gardener/gardener v1.11.3
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/jedib0t/go-pretty/v6 v6.1.0
	github.com/karrick/godirwalk v1.16.1
	github.com/ktr0731/go-fuzzyfinder v0.2.0
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20191120175047-4206685974f2
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451
	sigs.k8s.io/controller-runtime v0.6.3
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
