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

package validation_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"

	"github.com/danielfoehrkn/kubeswitch/pkg/config/validation"
	"github.com/danielfoehrkn/kubeswitch/types"
)

var _ = Describe("ValidateConfig", func() {
	var config types.Config

	BeforeEach(func() {
		config = types.Config{
			Version: "v1alpha1",
		}
	})

	It("should successfully validate an empty config", func() {
		errorList := validation.ValidateConfig(&config)
		Expect(errorList).To(BeEmpty())
	})

	It("should successfully validate config", func() {
		kubeconfigName := "gago"
		rediscoveryInterval := time.Minute * 60
		config := &types.Config{
			KubeconfigName:    &kubeconfigName,
			Version:           "v1alpha1",
			RefreshIndexAfter: &rediscoveryInterval,
			KubeconfigStores: []types.KubeconfigStore{
				{
					Kind:           types.StoreKindVault,
					KubeconfigName: &kubeconfigName,
					Paths:          []string{"path/abc", "path/abc/xyz"},
					Config:         nil,
				},
			},
		}
		errorList := validation.ValidateConfig(config)
		Expect(errorList).To(BeEmpty())
	})

	It("should throw error - invalid kubeconfig kind", func() {
		config := &types.Config{
			Version: "v1alpha1",
			KubeconfigStores: []types.KubeconfigStore{
				{
					Kind:  "invalid-kind",
					Paths: []string{"path/abc", "path/abc/xyz"},
				},
			},
		}
		errorList := validation.ValidateConfig(config)
		Expect(errorList).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("kubeconfigStores[0].kind"),
			})),
		))
	})

	It("should throw error - invalid config version", func() {
		config := &types.Config{
			Version: "my-version",
		}
		errorList := validation.ValidateConfig(config)
		Expect(errorList).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("version"),
			})),
		))
	})

	It("should throw error - no paths are configured for the kubeconfig store", func() {
		config := &types.Config{
			Version: "v1alpha1",
			KubeconfigStores: []types.KubeconfigStore{
				{
					Kind:  types.StoreKindVault,
					Paths: []string{},
				},
			},
		}
		errorList := validation.ValidateConfig(config)
		Expect(errorList).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("kubeconfigStores[0].paths"),
			})),
		))
	})

	It("should throw error - requires unique IDs when using multiple kubeconfig stores with the same kind and using an index", func() {
		minute := time.Minute
		config := &types.Config{
			Version:           "v1alpha1",
			RefreshIndexAfter: &minute,
			KubeconfigStores: []types.KubeconfigStore{
				{
					Kind:  types.StoreKindVault,
					Paths: []string{"ab"},
				},
				{
					Kind:  types.StoreKindVault,
					Paths: []string{"ab"},
				},
			},
		}

		errorList := validation.ValidateConfig(config)
		Expect(errorList).ToNot(BeEmpty())
		Expect(errorList).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("kubeconfigStores[1].id"),
			})),
		))
	})

	It("should throw error - requires unique IDs (rediscovery interval set on stores instead of globally)", func() {
		minute := time.Minute
		config := &types.Config{
			Version: "v1alpha1",
			KubeconfigStores: []types.KubeconfigStore{
				{
					Kind:              types.StoreKindVault,
					RefreshIndexAfter: &minute,
					Paths:             []string{"ab"},
				},
				{
					Kind:              types.StoreKindVault,
					RefreshIndexAfter: &minute,
					Paths:             []string{"ab"},
				},
			},
		}

		errorList := validation.ValidateConfig(config)
		Expect(errorList).ToNot(BeEmpty())
		Expect(errorList).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("kubeconfigStores[1].id"),
			})),
		))
	})

	It("should validate successfully via multiple kubeconfig stores with the same kind", func() {
		minute := time.Minute
		config := &types.Config{
			Version: "v1alpha1",
			KubeconfigStores: []types.KubeconfigStore{
				{
					Kind:              types.StoreKindVault,
					RefreshIndexAfter: &minute,
					Paths:             []string{"ab"},
					ID:                pointer.String("id-one"),
				},
				{
					Kind:              types.StoreKindVault,
					RefreshIndexAfter: &minute,
					Paths:             []string{"ab"},
					ID:                pointer.String("id-two"),
				},
			},
		}

		errorList := validation.ValidateConfig(config)
		Expect(errorList).To(BeEmpty())
	})

	Context("Gardener store", func() {
		It("should successfully validate the Gardener store config", func() {
			config := &types.Config{
				Version: "v1alpha1",
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind:  types.StoreKindGardener,
						Paths: []string{"garden", "garden-0xx1-x--y"},
						Config: types.StoreConfigGardener{
							GardenerAPIKubeconfigPath: "my-path-to-gardener-kubeconfig",
						},
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).To(BeEmpty())
		})

		It("should successfully validate the Gardener store config - multiple stores without ID but with landscape name", func() {
			duration := time.Minute
			config := &types.Config{
				Version:           "v1alpha1",
				RefreshIndexAfter: &duration,
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind: types.StoreKindGardener,
						Config: types.StoreConfigGardener{
							GardenerAPIKubeconfigPath: "my-path-to-gardener-kubeconfig",
							LandscapeName:             pointer.String("dev"),
						},
					},
					{
						Kind: types.StoreKindGardener,
						Config: types.StoreConfigGardener{
							GardenerAPIKubeconfigPath: "my-path-to-gardener-kubeconfig",
							LandscapeName:             pointer.String("canary"),
						},
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).To(BeEmpty())
		})

		It("should throw error - providing paths that are not gardener or gardener-", func() {
			config := &types.Config{
				Version: "v1alpha1",
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind:  types.StoreKindGardener,
						Paths: []string{"abc"},
						Config: types.StoreConfigGardener{
							GardenerAPIKubeconfigPath: "my-path-to-gardener-kubeconfig",
						},
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("kubeconfigStores[0].paths[0]"),
				})),
			))
		})

		It("should throw error - provided path / in addition to other paths", func() {
			config := &types.Config{
				Version: "v1alpha1",
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind:  types.StoreKindGardener,
						Paths: []string{"/", "garden"},
						Config: types.StoreConfigGardener{
							GardenerAPIKubeconfigPath: "my-path-to-gardener-kubeconfig",
						},
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("kubeconfigStores[0].paths[0]"),
				})),
			))
		})

		It("should throw error - the Gardener store needs configuration", func() {
			config := &types.Config{
				Version: "v1alpha1",
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind: types.StoreKindGardener,
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("kubeconfigStores[0].config"),
				})),
			))
		})

		It("should throw error - the Gardener store config cannot be parsed", func() {
			config := &types.Config{
				Version: "v1alpha1",
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind:   types.StoreKindGardener,
						Config: []string{"wrong-config"},
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("kubeconfigStores[0].config"),
				})),
			))
		})

		It("should throw error - the GardenerAPIKubeconfigPath must be set", func() {
			config := &types.Config{
				Version: "v1alpha1",
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind:   types.StoreKindGardener,
						Config: types.StoreConfigGardener{},
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("kubeconfigStores[0].config.gardenerAPIKubeconfigPath"),
				})),
			))
		})

		It("should throw error - the landscape name must not be empty (but can be nil)", func() {
			config := &types.Config{
				Version: "v1alpha1",
				KubeconfigStores: []types.KubeconfigStore{
					{
						Kind: types.StoreKindGardener,
						Config: types.StoreConfigGardener{
							GardenerAPIKubeconfigPath: "xy",
							LandscapeName:             pointer.String(""),
						},
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("kubeconfigStores[0].config.landscapeName"),
				})),
			))
		})
	})

	Context("Hooks", func() {
		It("should successfully validate hooks", func() {
			config := &types.Config{
				Version: "v1alpha1",
				Hooks: []types.Hook{
					{
						Name: "my-hooks",
						Type: types.HookTypeExecutable,
						Path: pointer.String("my-path"),
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).To(BeEmpty())
		})

		It("should throw error - invalid hook type", func() {
			config := &types.Config{
				Version: "v1alpha1",
				Hooks: []types.Hook{
					{
						Type: "unknown-type",
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("hooks[0].type"),
				})),
			))
		})

		It("should throw error - path to binary is required when specifying hook type executable ", func() {
			config := &types.Config{
				Version: "v1alpha1",
				Hooks: []types.Hook{
					{
						Type: types.HookTypeExecutable,
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("hooks[0].path"),
				})),
			))
		})

		It("should throw error - arguments are required when specifying hook type inline", func() {
			config := &types.Config{
				Version: "v1alpha1",
				Hooks: []types.Hook{
					{
						Type: types.HookTypeInlineCommand,
					},
				},
			}

			errorList := validation.ValidateConfig(config)
			Expect(errorList).ToNot(BeEmpty())
			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("hooks[0].arguments"),
				})),
			))
		})
	})
})
