package list_contexts

import (
	"fmt"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/types"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func ListContexts(stores []store.KubeconfigStore, switchConfig *types.Config, stateDir string) {
	c, err := pkg.DoSearch(stores, switchConfig, stateDir)
	if err != nil {
		logger.Warnf("cannot list contexts: %v", err)
	}

	for discoveredKubeconfig := range *c {
		if discoveredKubeconfig.Error != nil {
			logger.Warnf("cannot list contexts. Error returned from search: %v", discoveredKubeconfig.Error)
			continue
		}

		// write to STDIO
		for _, name := range discoveredKubeconfig.ContextNames {
			fmt.Println(name)
		}
	}
}
