package clean

import (
	"fmt"
	"io/ioutil"
	"os"

	kubeconfigutil "github.com/danielfoehrkn/kubectlSwitch/pkg/util/kubectx_copied"
)

func Clean() error {
	tempDir := os.ExpandEnv(kubeconfigutil.TemporaryKubeconfigDir)
	files, _ := ioutil.ReadDir(tempDir)
	err := os.RemoveAll(tempDir)
	if err != nil {
		return err
	}
	fmt.Printf("Cleaned %d files.", len(files))
	return nil
}
