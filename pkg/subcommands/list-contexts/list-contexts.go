package list_contexts

import (
	"fmt"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/types"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func ListContexts(stores []store.KubeconfigStore, switchConfig *types.Config, stateDir string) error {
	c, err := pkg.DoSearch(stores, switchConfig, stateDir)
	if err != nil {
		return fmt.Errorf("cannot list contexts: %v", err)
	}

	for discoveredKubeconfig := range *c {
		if discoveredKubeconfig.Error != nil {
			logger.Warnf("cannot list contexts. Error returned from search: %v", discoveredKubeconfig.Error)
			continue
		}

		name := discoveredKubeconfig.Name
		if len(discoveredKubeconfig.Alias) > 0 {
			name = discoveredKubeconfig.Alias
		}

		// write to STDIO
		fmt.Println(name)
	}

	return nil
}
