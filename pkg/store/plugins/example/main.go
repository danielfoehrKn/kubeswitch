// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/danielfoehrkn/kubeswitch/pkg/store/plugins"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
)

// Store is the implementation of the store plugin
type Store struct {
	Logger hclog.Logger
}

// GetID returns the ID of the store
func (s *Store) GetID(ctx context.Context) (string, error) {
	return "example", nil
}

// GetContextPrefix returns the context prefix
func (s *Store) GetContextPrefix(ctx context.Context, path string) (string, error) {
	return "example", nil
}

// VerifyKubeconfigPaths verifies the kubeconfig paths
func (s *Store) VerifyKubeconfigPaths(ctx context.Context) error {
	return nil
}

// StartSearch starts the search
func (s *Store) StartSearch(ctx context.Context, channel chan storetypes.SearchResult) {
	channel <- storetypes.SearchResult{
		KubeconfigPath: "fake",
		Error:          nil,
	}
}

// GetKubeconfigForPath gets the kubeconfig for the path
func (s *Store) GetKubeconfigForPath(ctx context.Context, path string, tags map[string]string) ([]byte, error) {
	if path == "fake" {
		// valid fake kubeconfig
		file := `apiVersion: v1
kind: Config
preferences: {}

clusters:
- name: development
- name: test

users:
- name: developer
- name: experimenter

contexts:
- name: dev-frontend
- name: dev-storage
- name: exp-test`
		return []byte(file), nil
	}

	s.Logger.Error("invalid path")
	return nil, nil
}

func main() {
	logger := hclog.Default()
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			"store": &plugins.StorePlugin{Impl: &Store{Logger: logger}},
		},
		Logger: logger,

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
