package clean

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/danielfoehrkn/kubectlSwitch/pkg/util"
)

func Clean() error {
	tempDir := os.ExpandEnv(util.TemporaryKubeconfigDir)
	files, _ := ioutil.ReadDir(tempDir)
	err := os.RemoveAll(tempDir)
	if err != nil {
		return err
	}
	fmt.Printf("Cleaned %d files.", len(files))
	return nil
}
