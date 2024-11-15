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

package switcher

import (
	"fmt"
	"os"
	"strings"

	"github.com/bombsimon/logrusr/v4"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/danielfoehrkn/kubeswitch/pkg"

	"github.com/danielfoehrkn/kubeswitch/pkg/cache"
	"github.com/danielfoehrkn/kubeswitch/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/utils/ptr"

	switchconfig "github.com/danielfoehrkn/kubeswitch/pkg/config"
	"github.com/danielfoehrkn/kubeswitch/pkg/config/validation"
	"github.com/danielfoehrkn/kubeswitch/pkg/store"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
	"github.com/danielfoehrkn/kubeswitch/types"
)

const (
	vaultTokenFileName          = ".vault-token"
	defaultKubeconfigName       = "config"
	defaultKubeconfigPath       = "$HOME/.kube/config"
	linuxEnvKubeconfigSeperator = ":"
)

var (
	// root command
	kubeconfigPath string
	kubeconfigName string
	showPreview    bool
	deleteContext  bool
	unsetContext   bool
	currentContext bool

	// vault store
	storageBackend          string
	vaultAPIAddressFromFlag string

	// hook command
	configPath     string
	stateDirectory string
	hookName       string
	runImmediately bool

	// version command
	version   string
	buildDate string

	showDebugLogs bool
	noIndex       bool

	rootCommand = &cobra.Command{
		Use:     "switcher",
		Short:   "Launch the switch binary",
		Long:    `The kubectx for operators.`,
		Version: version,
		Args: func(cmd *cobra.Command, args []string) error {
			switch {
			case deleteContext:
				if err := cobra.ExactArgs(1)(cmd, args); err != nil {
					return err
				}
			case unsetContext || currentContext:
				if err := cobra.NoArgs(cmd, args); err != nil {
					return err
				}
			}
			return cmd.ParseFlags(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case deleteContext:
				return deleteContextCmd.RunE(cmd, args)
			case unsetContext:
				return unsetContextCmd.RunE(cmd, args)
			case currentContext:
				return currentContextCmd.RunE(cmd, args)
			}

			if len(args) > 0 {
				switch args[0] {
				case "-":
					return previousContextCmd.RunE(cmd, args[1:])
				case ".":
					return lastContextCmd.RunE(cmd, args[1:])
				default:
					return setContextCmd.RunE(cmd, args)
				}
			}

			stores, config, err := initialize()
			if err != nil {
				return err
			}

			// config file setting overwrites the command line default (--showPreview true)
			if showPreview && config.ShowPreview != nil && !*config.ShowPreview {
				showPreview = false
			}

			kubeconfigPath, contextName, err := pkg.Switcher(stores, config, stateDirectory, noIndex, showPreview)
			reportNewContext(kubeconfigPath, contextName)
			return err
		},
		SilenceUsage: true,
	}
)

func init() {
	setFlagsForContextCommands(rootCommand)
	rootCommand.Flags().BoolVarP(&deleteContext, "d", "d", false, "delete desired context. Context name is required")
	rootCommand.Flags().BoolVarP(&unsetContext, "unset", "u", false, "unset current context")
	rootCommand.Flags().BoolVarP(&currentContext, "current", "c", false, "show current context")
}

func NewCommandStartSwitcher() *cobra.Command {
	return rootCommand
}

func setCommonFlags(command *cobra.Command) {
	command.Flags().BoolVar(
		&showDebugLogs,
		"debug",
		false,
		"show debug logs")
	command.Flags().BoolVar(
		&noIndex,
		"no-index",
		false,
		"stores do not read from index files. The index is refreshed.")
	command.Flags().StringVar(
		&kubeconfigPath,
		"kubeconfig-path",
		defaultKubeconfigPath,
		"path to be recursively searched for kubeconfigs. Can be a file or a directory on the local filesystem or a path in Vault.")
	command.Flags().StringVar(
		&stateDirectory,
		"state-directory",
		os.ExpandEnv("$HOME/.kube/switch-state"),
		"path to the local directory used for storing internal state.")
}

func initialize() ([]storetypes.KubeconfigStore, *types.Config, error) {
	if showDebugLogs {
		logrus.SetLevel(logrus.DebugLevel)
	}

	config, err := switchconfig.LoadConfigFromFile(util.ExpandEnv(configPath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read switch config file: %v", err)
	}

	if config != nil {
		if errList := validation.ValidateConfig(config); len(errList) > 0 {
			return nil, nil, fmt.Errorf("the switch configuration file contains errors: %s", errList.ToAggregate().Error())
		}
	} else {
		config = &types.Config{}
	}

	if kubeconfigName == defaultKubeconfigName {
		if config.KubeconfigName != nil && *config.KubeconfigName != "" {
			kubeconfigName = *config.KubeconfigName
		}
	}

	storeFromFlags := getStoreFromFlagAndEnv(config)
	if storeFromFlags != nil {
		config.KubeconfigStores = append(config.KubeconfigStores, *storeFromFlags)
	}

	if len(config.KubeconfigStores) == 0 {
		return nil, nil, fmt.Errorf("you need to point kubeswitch to a kubeconfig file. This can be done by setting the environment variable KUBECONFIG, setting the flag --kubeconfig-path, having a default kubeconfig file at ~/.kube/config or providing a switch configuration file")
	}

	var (
		stores                          []storetypes.KubeconfigStore
		digitalOceanStoreAddedViaConfig bool
	)
	for _, kubeconfigStoreFromConfig := range config.KubeconfigStores {
		var s storetypes.KubeconfigStore

		if kubeconfigStoreFromConfig.KubeconfigName != nil && *kubeconfigStoreFromConfig.KubeconfigName != "" {
			kubeconfigName = *kubeconfigStoreFromConfig.KubeconfigName
		}

		switch kubeconfigStoreFromConfig.Kind {
		case types.StoreKindFilesystem:
			filesystemStore, err := store.NewFilesystemStore(kubeconfigName, kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = filesystemStore

		case types.StoreKindVault:
			vaultStore, err := store.NewVaultStore(vaultAPIAddressFromFlag,
				vaultTokenFileName,
				kubeconfigName,
				kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = vaultStore

		case types.StoreKindGardener:
			gardenerStore, err := store.NewGardenerStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to create Gardener store: %w", err)
			}
			s = gardenerStore

		case types.StoreKindGKE:
			gkeStore, err := store.NewGKEStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to create GKE store: %w", err)
			}
			s = gkeStore

		case types.StoreKindAzure:
			azureStore, err := store.NewAzureStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, fmt.Errorf("unable to create Azure store: %w", err)
			}
			s = azureStore
		case types.StoreKindEKS:
			eksStore, err := store.NewEKSStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = eksStore
		case types.StoreKindRancher:
			rancherStore, err := store.NewRancherStore(kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = rancherStore
		case types.StoreKindOVH:
			ovhStore, err := store.NewOVHStore(kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = ovhStore
		case types.StoreKindScaleway:
			scalewayStore, err := store.NewScalewayStore(kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = scalewayStore
		case types.StoreKindDigitalOcean:
			doStore, err := store.NewDigitalOceanStore(kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = doStore
			digitalOceanStoreAddedViaConfig = true
		case types.StoreKindAkamai:
			akamaiStore, err := store.NewAkamaiStore(kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = akamaiStore
		case types.StoreKindCapi:
			capiStore, err := store.NewCapiStore(kubeconfigStoreFromConfig, stateDirectory)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = capiStore
		case types.StoreKindPlugin:
			pluginStore, err := store.NewPluginStore(kubeconfigStoreFromConfig)
			if err != nil {
				if kubeconfigStoreFromConfig.Required != nil && !*kubeconfigStoreFromConfig.Required {
					continue
				}
				return nil, nil, err
			}
			s = pluginStore
		default:
			return nil, nil, fmt.Errorf("unknown store %q", kubeconfigStoreFromConfig.Kind)
		}

		if showDebugLogs {
			s.GetLogger().Logger.SetLevel(logrus.DebugLevel)
		}

		// Add cache to the store
		// defaults to in-memory cache -> prevents duplicate reads of the same kubeconfig
		if cacheCfg := kubeconfigStoreFromConfig.Cache; cacheCfg == nil {
			s, err = cache.New("memory", s, nil)
		} else {
			s, err = cache.New(cacheCfg.Kind, s, cacheCfg)
		}
		if err != nil {
			return nil, nil, err
		}
		stores = append(stores, s)
	}

	// the Digital Ocean store is enabled by default for a seamless experience for `doctl` users (automatically discovers the `doctl` config file with stored credentials)
	// this is optional, so don't care about errors
	if !digitalOceanStoreAddedViaConfig {
		doStore, _ := store.NewDigitalOceanStore(types.KubeconfigStore{
			ID:   ptr.To("doDefaultStore"),
			Kind: types.StoreKindDigitalOcean,
			// for users with outdated `doctl` configs, don't show errors if they have no explicitly enabled the DO backing store
			Required:   ptr.To(false),
			ShowPrefix: ptr.To(true),
		})
		if doStore != nil {
			// we found a valid `doctl` config, hence add Digital Ocean as a backing store with default configuration
			s, err := cache.New("memory", doStore, nil)
			if err != nil {
				return nil, nil, err
			}
			stores = append(stores, s)
		}
	}

	// set 'logr' log implementation for the controller-runtime (otherwise controller-runtime code cannot log)
	log := logrusr.New(logrus.New())
	logf.SetLogger(log)

	return stores, config, nil
}

// getStoreFromFlagAndEnv translates the kubeconfig flag --kubeconfig-path & environment variable KUBECONFIG into a
// dedicated store in addition to the stores configured in the switch-config.yaml.
// This way, it is "just another store" -> does not need special handling
func getStoreFromFlagAndEnv(config *types.Config) *types.KubeconfigStore {
	var paths []string

	pathFromFlag := getKubeconfigPathFromFlag()
	if len(pathFromFlag) > 0 {
		logrus.Debugf("Using kubeconfig path from flag %s", pathFromFlag)
		paths = append(paths, pathFromFlag)
	}

	kubeconfigPathFromEnv := os.Getenv("KUBECONFIG")

	pathsFromEnv := strings.Split(kubeconfigPathFromEnv, linuxEnvKubeconfigSeperator)

	for _, path := range pathsFromEnv {
		if !isDuplicatePath(config.KubeconfigStores, path) && !strings.HasSuffix(path, ".tmp") && path != "" {
			// the KUBECONFIG env sets a unique, non kubeswitch set, env variable to a kubeconfig.
			paths = append(paths, util.ExpandEnv(path))
			logrus.Debugf("Adding kubeconfig path from KUBECONFIG env %s", kubeconfigPathFromEnv)
		}
	}

	if len(paths) == 0 {
		return nil
	}

	return &types.KubeconfigStore{
		ID:             ptr.To("env-and-flag"),
		Kind:           types.StoreKind(storageBackend),
		KubeconfigName: ptr.To(kubeconfigName),
		Paths:          paths,
		ShowPrefix:     ptr.To(false),
	}
}

// getKubeconfigPathFromFlag gets the kubeconfig path configured in the flag --kubeconfig-path
// does not add the path in case the configured path does not exist
// this is to not require a kubeconfig file in the default location
func getKubeconfigPathFromFlag() string {
	if len(kubeconfigPath) == 0 {
		return ""
	}

	kubeconfigPath = strings.ReplaceAll(kubeconfigPath, "~", "$HOME")
	if kubeconfigPath == defaultKubeconfigPath {
		// do not return, if the kubeconfig under the default kubeconfig path does not exist
		if _, err := os.Stat(os.ExpandEnv(defaultKubeconfigPath)); err != nil {
			return ""
		}
		// the kubeconfig under the default path exists -> return it.
		return os.ExpandEnv(defaultKubeconfigPath)
	}

	// the flag sets a non-default kubeconfig path
	return os.ExpandEnv(kubeconfigPath)
}

// isDuplicatePath searches through all kubeconfig stores in the switch-config.yaml and checks if the
// given path is already configured in any of these stores
// returns true if it is already configureed
func isDuplicatePath(kubeconfigStores []types.KubeconfigStore, newPath string) bool {
	// O(n square) operation
	// but fortunately it is highly unlikely that there are many stores and paths configured
	for _, store := range kubeconfigStores {
		for _, path := range store.Paths {
			if path == newPath {
				return true
			}
		}
	}
	return false
}
