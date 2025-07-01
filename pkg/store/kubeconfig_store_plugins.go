// Copyright 2024 The Kubeswitch authors
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

package store

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/danielfoehrkn/kubeswitch/pkg/store/plugins"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
)

func NewPluginStore(store types.KubeconfigStore) (*PluginStore, error) {
	storePlugin := &types.StoreConfigPlugin{}
	if store.Config != nil {
		buf, err := yaml.Marshal(store.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to process plugin store config: %w", err)
		}

		err = yaml.Unmarshal(buf, storePlugin)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal plugin config: %w", err)
		}
	}

	return &PluginStore{
		Logger:          logrus.New().WithField("store", types.StoreKindPlugin),
		KubeconfigStore: store,
		Config:          storePlugin,
	}, nil
}

func convertLogrusLevelToHclogLevel(level logrus.Level) hclog.Level {
	switch level {
	case logrus.TraceLevel:
		return hclog.Trace
	case logrus.DebugLevel:
		return hclog.Debug
	case logrus.InfoLevel:
		return hclog.Info
	case logrus.WarnLevel:
		return hclog.Warn
	default:
		// Default to Error level for any other logrus levels (e.g. Fatal or Panic)
		return hclog.Error
	}
}

// InitializePluginStore initializes the plugin store
func (s *PluginStore) InitializePluginStore() error {

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugins.Handshake,
		Plugins:         plugins.PluginMap,
		Cmd:             exec.Command(s.Config.CmdPath, s.Config.Args...),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  convertLogrusLevelToHclogLevel(s.Logger.Level),
			Name:   "plugin",
		}),
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return err
	}

	plugin, err := rpcClient.Dispense("store")
	if err != nil {
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	c, ok := plugin.(plugins.Store)
	if !ok {
		return fmt.Errorf("plugin does not implement Store interface")
	}

	s.Client = c

	return nil
}

// GetID returns the unique store ID
func (s *PluginStore) GetID() string {
	ctx := context.Background()

	id, err := s.Client.GetID(ctx)
	if err != nil {
		localID := "default"
		if s.KubeconfigStore.ID != nil {
			localID = *s.KubeconfigStore.ID
		}
		return fmt.Sprintf("%s.%s", s.GetKind(), localID)
	}
	return id
}

func (s *PluginStore) GetKind() types.StoreKind {
	return types.StoreKindPlugin
}

func (s *PluginStore) GetContextPrefix(path string) string {
	ctx := context.Background()

	prefix, err := s.Client.GetContextPrefix(ctx, path)
	if err != nil {
		return fmt.Sprintf("%s/%s", s.GetKind(), path)
	}
	return prefix
}

func (s *PluginStore) VerifyKubeconfigPaths() error {
	ctx := context.Background()

	if err := s.InitializePluginStore(); err != nil {
		return err
	}

	return s.Client.VerifyKubeconfigPaths(ctx)
}

func (s *PluginStore) GetStoreConfig() types.KubeconfigStore {
	return s.KubeconfigStore
}

func (s *PluginStore) GetLogger() *logrus.Entry {
	return s.Logger
}

func (s *PluginStore) StartSearch(channel chan storetypes.SearchResult) {
	s.Logger.Debug("Plugin: start search")

	ctx := context.Background()

	if err := s.InitializePluginStore(); err != nil {
		channel <- storetypes.SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}

	s.Client.StartSearch(ctx, channel)
}

func (s *PluginStore) GetKubeconfigForPath(path string, tags map[string]string) ([]byte, error) {
	s.Logger.Debugf("Plugins: get kubeconfig for path %s", path)

	ctx := context.Background()

	if err := s.InitializePluginStore(); err != nil {
		return nil, err
	}

	return s.Client.GetKubeconfigForPath(ctx, path, tags)
}
