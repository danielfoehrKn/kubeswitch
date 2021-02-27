// Copyright 2021 Daniel Foehr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clean

import (
	"fmt"
	"io/ioutil"
	"os"

	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
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
