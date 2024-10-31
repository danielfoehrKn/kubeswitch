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

package ns

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	historyutil "github.com/danielfoehrkn/kubeswitch/pkg/subcommands/history/util"
	kubeconfigutil "github.com/danielfoehrkn/kubeswitch/pkg/util/kubectx_copied"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultKubeconfigPath       = "$HOME/.kube/config"
	linuxEnvKubeconfigSeperator = ":"
)

var (
	kubeconfigPathFromEnv = os.Getenv("KUBECONFIG")
	// only use namespace cache for contexts switched to by the switch tool
	cache         *NamespaceCache
	logger        = logrus.New()
	hotReloadLock sync.RWMutex

	allNamespaces []string
)

// SwitchToNamespace takes a target namespace and - given that the namespace exists - sets it on the current kubeconfig file
func SwitchToNamespace(targetNamespace, kubeconfigPathFromFlag string, checkExistence bool) error {
	kubeconfigPath, err := getKubeconfigPath(kubeconfigPathFromFlag)
	if err != nil {
		return err
	}

	if checkExistence {
		c, err := getClient(kubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed to retrieve current namespaces: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ns := corev1.Namespace{}
		if err := c.Get(ctx, client.ObjectKey{Name: targetNamespace}, &ns); err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("namespace %q not found", targetNamespace)
			}
			return fmt.Errorf("failed to find namespace %q: %v", targetNamespace, err)
		}
	}

	kubeconfig, err := kubeconfigutil.NewKubeconfigForPath(kubeconfigPath)
	if err != nil {
		return err
	}

	if err := kubeconfig.SetNamespaceForCurrentContext(targetNamespace); err != nil {
		return fmt.Errorf("failed to set namespace %q: %v", targetNamespace, err)
	}

	// this updates the actual kubeconfif file (does not create a new tmp. kubeconfig to set namespace)
	if _, err := kubeconfig.WriteKubeconfigFile(); err != nil {
		return fmt.Errorf("failed to write kubeconfig file: %v", err)
	}

	kubeswitchContext := kubeconfig.GetKubeswitchContext()
	if err := historyutil.AppendToHistory(kubeswitchContext, targetNamespace); err != nil {
		return fmt.Errorf("failed to write namespace history: %v", err)
	}

	return nil
}

// SwitchNamespace retrieves all available namespaces (either via API call or from local cache)
// Then sets the selected namespace on the current kubeconfig file (does not create a new tmp. kubeconfig to set namespace)
func SwitchNamespace(kubeconfigPathFromFlag, stateDir string, noIndex bool) error {
	cachedNamespaces := sets.NewString()

	kubeconfigPath, err := getKubeconfigPath(kubeconfigPathFromFlag)
	if err != nil {
		return err
	}

	kubeconfig, err := kubeconfigutil.NewKubeconfigForPath(kubeconfigPath)
	if err != nil {
		return err
	}

	kubeswitchContext := kubeconfig.GetKubeswitchContext()

	if len(kubeswitchContext) > 0 && !noIndex {
		cache, err = NewNamespaceCache(stateDir, kubeswitchContext)
		if err != nil {
			logger.Warnf("failed to use namespace cache: %v", err)
		}
		allNamespaces = cache.GetContent()
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cachedNamespaces.Insert(allNamespaces...)

		client, err := getClient(kubeconfigPath)
		if err != nil {
			logger.Warnf("failed to retrieve current namespaces: %v", err)
			return
		}

		// list all the namespaces
		list := &corev1.NamespaceList{}
		if err := client.List(ctx, list); err != nil {
			logger.Warnf("failed to retrieve current namespaces: %v", err)
			return
		}

		realNs := sets.NewString()
		for _, namespace := range list.Items {
			realNs.Insert(namespace.Name)
		}

		n := 0
		// filter array in place
		for _, namespaceInCache := range allNamespaces {
			// this overwrites the index in the array which contains a namespace that is in the cache,
			// but not in the cluster
			if realNs.Has(namespaceInCache) {
				allNamespaces[n] = namespaceInCache
				n++
			}
		}
		// update the slice-internal array pointer to point only to the potentially shorter range of values
		allNamespaces = allNamespaces[:n]

		// add namespaces that are not in the cached maespace list
		for _, ns := range realNs.List() {
			if !cachedNamespaces.Has(ns) {
				allNamespaces = append(allNamespaces, ns)
			}
		}
	}()

	idx, err := fuzzyfinder.Find(
		&allNamespaces,
		func(i int) string {
			return allNamespaces[i]
		},
		fuzzyfinder.WithHotReloadLock(hotReloadLock.RLocker()),
	)
	if err != nil {
		return err
	}

	selectedNamespace := allNamespaces[idx]

	logger.Debugf("setting namespace %q to kubeconfig with path %q", selectedNamespace, kubeconfigPath)

	if err := kubeconfig.SetNamespaceForCurrentContext(selectedNamespace); err != nil {
		return fmt.Errorf("failed to set namespace %q: %v", selectedNamespace, err)
	}

	if _, err := kubeconfig.WriteKubeconfigFile(); err != nil {
		return fmt.Errorf("failed to write kubeconfig file: %v", err)
	}

	if len(kubeswitchContext) == 0 {
		return nil
	}

	if err := historyutil.AppendToHistory(kubeswitchContext, selectedNamespace); err != nil {
		return fmt.Errorf("failed to write namespace history: %v", err)
	}

	return cache.Write(allNamespaces)
}

// ListNamespaces retrieves all available namespaces (either via API call or from local cache)
func ListNamespaces(kubeconfigPathFromFlag, stateDir string, noIndex bool) ([]string, error) {
	cachedNamespaces := sets.NewString()

	kubeconfigPath, err := getKubeconfigPath(kubeconfigPathFromFlag)
	if err != nil {
		return nil, err
	}
	kubeconfig, err := kubeconfigutil.NewKubeconfigForPath(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	kubeswitchContext := kubeconfig.GetKubeswitchContext()
	if len(kubeswitchContext) == 0 {
		// If no context return
		return nil, nil
	}

	if len(kubeswitchContext) > 0 && !noIndex {
		cache, err = NewNamespaceCache(stateDir, kubeswitchContext)
		if err != nil {
			logger.Warnf("failed to use namespace cache: %v", err)
		}
		allNamespaces = cache.GetContent()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cachedNamespaces.Insert(allNamespaces...)

	// Build the Kubernetes client configuration
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// Create the Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// list all the namespaces
	list, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	realNs := sets.NewString()
	for _, namespace := range list.Items {
		realNs.Insert(namespace.Name)
	}

	n := 0
	// filter array in place
	for _, namespaceInCache := range allNamespaces {
		// this overwrites the index in the array which contains a namespace that is in the cache,
		// but not in the cluster
		if realNs.Has(namespaceInCache) {
			allNamespaces[n] = namespaceInCache
			n++
		}
	}
	// update the slice-internal array pointer to point only to the potentially shorter range of values
	allNamespaces = allNamespaces[:n]

	// add namespaces that are not in the cached maespace list
	for _, ns := range realNs.List() {
		if !cachedNamespaces.Has(ns) {
			allNamespaces = append(allNamespaces, ns)
		}
	}

	err = cache.Write(allNamespaces)
	if err != nil {
		return nil, err
	}
	return allNamespaces, nil
}

func getKubeconfigPath(kubeconfigPathFromFlag string) (string, error) {
	kubeconfigPath := kubeconfigPathFromFlag

	// kubeconfig path from flag is preferred over env (just not if it is only the default)
	if (len(kubeconfigPath) == 0 || kubeconfigPath == os.ExpandEnv(defaultKubeconfigPath)) && len(kubeconfigPathFromEnv) > 0 {
		if len(strings.Split(kubeconfigPathFromEnv, linuxEnvKubeconfigSeperator)) > 1 {
			return "", fmt.Errorf("providing multiple kubeconfig files via environment variable KUBECONFIG is not supported for namespace switching")
		}

		kubeconfigPath = os.ExpandEnv(kubeconfigPathFromEnv)
	}

	if _, err := os.Stat(kubeconfigPath); err != nil {
		return "", fmt.Errorf("unable to list namespaces. The kubeconfig file %q does not exist", kubeconfigPath)
	}
	return kubeconfigPath, nil
}

func getClient(kubeconfigPath string) (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create rest config: %v", err)
	}

	// increase QPS and Burst to avoid rate limiting, these values are the same as kubectl uses
	restConfig.QPS = 50.0
	restConfig.Burst = 300

	client, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes client: %v", err)
	}
	return client, nil
}
