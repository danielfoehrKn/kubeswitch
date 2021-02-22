package history

import (
	"fmt"

	"github.com/danielfoehrkn/kubectlSwitch/pkg/store"
	"github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/history/util"
	setcontext "github.com/danielfoehrkn/kubectlSwitch/pkg/subcommands/set-context"
	"github.com/danielfoehrkn/kubectlSwitch/types"
	"github.com/ktr0731/go-fuzzyfinder"
)

func ListHistory(stores []store.KubeconfigStore, switchConfig *types.Config, stateDir string) error {
	history, err := util.ReadHistory()
	if err != nil {
		return err
	}

	idx, err := fuzzyfinder.Find(
		history,
		func(i int) string {
			return fmt.Sprintf("%d: %s", len(history)-i-1, history[i])
		})

	if err != nil {
		return err
	}

	return setcontext.SetContext(history[idx], stores, switchConfig, stateDir)
}

func SetPreviousContext(stores []store.KubeconfigStore, switchConfig *types.Config, stateDir string) error {
	history, err := util.ReadHistory()
	if err != nil {
		return err
	}

	if len(history) == 0 {
		return nil
	}

	var position int
	if len(history) == 1 {
		position = 0
	} else {
		position = 1
	}

	return setcontext.SetContext(history[position], stores, switchConfig, stateDir)
}
