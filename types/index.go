// Copyright 2021 The Kubeswitch authors
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

package types

import "time"

// Index defines how an index for a kubeconfig store is written
type Index struct {
	// Kind defines which store kind this Index belongs to
	Kind StoreKind `yaml:"kind"`
	// ContextToPathMapping contains the actual index, mapping the context name
	// to the kubeconfig path in the backing store
	ContextToPathMapping map[string]string `yaml:"contextToPathMapping"`
}

// IndexState defines how the state of an index for a kubeconfig store is written
type IndexState struct {
	// Kind defines which store kind this Index belongs to
	Kind StoreKind `yaml:"kind"`
	// LastUpdateTime is the last time the index has been updated
	LastUpdateTime time.Time `yaml:"lastExecutionTime"`
}
