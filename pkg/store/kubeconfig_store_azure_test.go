// Copyright 2026 The Kubeswitch authors
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
	"strings"
	"testing"

	"github.com/danielfoehrkn/kubeswitch/types"
)

func TestNewAzureStoreSubscriptionID(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
	}{
		{name: "missing config", wantError: true},
		{name: "missing subscription", config: map[string]interface{}{"resourceGroups": []string{"test"}}, wantError: true},
		{name: "null subscription", config: map[string]interface{}{"subscriptionID": nil}, wantError: true},
		{name: "empty subscription", config: map[string]interface{}{"subscriptionID": ""}, wantError: true},
		{name: "blank subscription", config: map[string]interface{}{"subscriptionID": " \t"}, wantError: true},
		{name: "configured subscription", config: map[string]interface{}{"subscriptionID": "test-subscription"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewAzureStore(types.KubeconfigStore{Kind: types.StoreKindAzure, Config: tt.config}, "")
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), "subscriptionID") {
					t.Fatalf("expected subscriptionID configuration error, got %v", err)
				}
				if s != nil {
					t.Fatal("invalid configuration returned a store")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if s.Config.SubscriptionID == nil || *s.Config.SubscriptionID != "test-subscription" {
				t.Fatalf("subscriptionID was not preserved: %#v", s.Config.SubscriptionID)
			}
			if s.AksClient != nil {
				t.Fatal("constructor initialized an Azure client")
			}
		})
	}
}
