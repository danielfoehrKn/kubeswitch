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

package cache

import (
	"fmt"
	"sync"

	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
)

var (
	cachesMu sync.RWMutex
	caches   = make(map[string]CacheFactory)
)

type CacheFactory func(store storetypes.KubeconfigStore, cfg *types.Cache) (storetypes.KubeconfigStore, error)

func Register(kind string, creator CacheFactory) {
	cachesMu.Lock()
	defer cachesMu.Unlock()
	caches[kind] = creator
}

func New(kind string, store storetypes.KubeconfigStore, cfg *types.Cache) (storetypes.KubeconfigStore, error) {
	cachesMu.RLock()
	create, ok := caches[kind]
	cachesMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no cache factory registered for kind %s", kind)
	}
	cache, err := create(store, cfg)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

type Flushable interface {
	Flush() (int, error)
}
