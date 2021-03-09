module github.com/danielfoehrkn/kubeswitch

go 1.12

require (
	github.com/disiqueira/gotree v1.0.0
	github.com/gardener/gardener v1.18.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/jedib0t/go-pretty/v6 v6.1.0
	github.com/karrick/godirwalk v1.16.1
	github.com/ktr0731/go-fuzzyfinder v0.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.5
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.6
	k8s.io/apimachinery v0.19.6
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/controller-runtime v0.7.1
)

replace k8s.io/client-go => k8s.io/client-go v0.19.6
