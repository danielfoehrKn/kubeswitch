package history

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"

	"github.com/danielfoehrkn/kubectlSwitch/pkg/util"
)

func ListHistory() error {
	history, err  := util.ReadHistory()
	if err != nil {
		return err
	}

	idx, err := fuzzyfinder.Find(
		history,
		func(i int) string {
			return fmt.Sprintf("%d: %s", len(history) - i - 1, history[i])
		})

	if err != nil {
		return err
	}

	// print selection
	fmt.Println(history[idx])
	return nil
}
