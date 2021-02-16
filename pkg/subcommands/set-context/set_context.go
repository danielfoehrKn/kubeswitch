package setcontext

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/danielfoehrkn/kubectlSwitch/pkg"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/util"
	"github.com/danielfoehrkn/kubectlSwitch/types"
)

var logger = logrus.New()

func SetContext(desiredContext string, stores []store.KubeconfigStore, switchConfig *types.Config, stateDir string) error {
	c, err := pkg.DoSearch(stores, switchConfig, stateDir)
	if err != nil {
		return err
	}

	var mError *multierror.Error
	for discoveredKubeconfig := range *c {
		if discoveredKubeconfig.Error != nil {
			// remember in case the wanted context name cannot be found
			mError = multierror.Append(mError, discoveredKubeconfig.Error)
			continue
		}

		if discoveredKubeconfig.Store == nil {
			// this should not happen
			logger.Debugf("store returned from search is nil. This should not happen")
			continue
		}
		kubeconfigStore := *discoveredKubeconfig.Store

		for _, name := range discoveredKubeconfig.ContextNames {
			contextWithoutFolderPrefix := util.GetContextWithoutFolderPrefix(name)
			if desiredContext == name || desiredContext == contextWithoutFolderPrefix {
				kubeconfigData, err := kubeconfigStore.GetKubeconfigForPath(discoveredKubeconfig.Path)
				if err != nil {
					return err
				}

				tempKubeconfigPath, err := util.SetCurrentContextOnTemporaryKubeconfigFile(kubeconfigData, contextWithoutFolderPrefix)
				if err != nil {
					return fmt.Errorf("failed to write current context to temporary kubeconfig: %v", err)
				}

				// print kubeconfig path to std.out -> captured by calling bash script to set KUBECONFIG environment Variable
				fmt.Print(tempKubeconfigPath)
				return nil
			}
		}

	}

	if mError != nil{
		return fmt.Errorf("context with name %q not found. Possibly due to errors: %v", desiredContext, mError.Error())
	}

	return fmt.Errorf("context with name %q not found", desiredContext)
}
