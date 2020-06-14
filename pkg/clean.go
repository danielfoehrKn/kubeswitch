package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
)

func Clean() error {
	tempDir := os.ExpandEnv(temporaryKubeconfigDir)
	files,_ := ioutil.ReadDir(tempDir)
	err := os.RemoveAll(tempDir)
	if err != nil {
		return err
	}
	fmt.Printf("Cleaned %d files.", len(files))
	return nil
}
