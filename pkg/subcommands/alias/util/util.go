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

package util

import kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"

// GetContextForAlias returns the alias for the given context or an empty string given a map (context -> alias)
func GetContextForAlias(context string, mapping map[string]string) string {
	if value, ok := mapping[context]; ok {
		return value
	}
	if value, ok := mapping[kubeconfigutil.GetContextWithoutFolderPrefix(context)]; ok {
		return value
	}
	return ""
}
