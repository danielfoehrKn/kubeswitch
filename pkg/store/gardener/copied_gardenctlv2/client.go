/*
SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

package gardenclient

import (
	"context"
	"errors"
	"fmt"

	openstackinstall "github.com/gardener/gardener-extension-provider-openstack/pkg/apis/openstack/install"
	openstackv1alpha1 "github.com/gardener/gardener-extension-provider-openstack/pkg/apis/openstack/v1alpha1"
	gardencore "github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	seedmanagementv1alpha1 "github.com/gardener/gardener/pkg/apis/seedmanagement/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var decoder runtime.Decoder

func init() {
	extensionsScheme := runtime.NewScheme()
	utilruntime.Must(openstackinstall.AddToScheme(extensionsScheme))
	decoder = serializer.NewCodecFactory(extensionsScheme).UniversalDecoder()
}

//go:generate mockgen -destination=./mocks/mock_client.go -package=mocks github.com/gardener/gardenctl-v2/internal/gardenclient Client

// Client returns a new client with functions to get Gardener and Kubernetes resources
type Client interface {
	// GetProject returns a Gardener project resource by name
	GetProject(ctx context.Context, name string) (*gardencorev1beta1.Project, error)
	// GetProjectByNamespace returns a Gardener project resource by namespace
	GetProjectByNamespace(ctx context.Context, namespace string) (*gardencorev1beta1.Project, error)
	// ListProjects returns all Gardener project resources
	ListProjects(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.ProjectList, error)

	// GetSeed returns a Gardener seed resource by name
	GetSeed(ctx context.Context, name string) (*gardencorev1beta1.Seed, error)
	// ListSeeds returns all Gardener seed resources
	ListSeeds(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.SeedList, error)
	// GetSeedClientConfig returns the client config for a seed
	GetSeedClientConfig(ctx context.Context, name string) (clientcmd.ClientConfig, error)

	// GetShoot returns a Gardener shoot resource in a namespace by name
	GetShoot(ctx context.Context, namespace, name string) (*gardencorev1beta1.Shoot, error)
	// FindShoot tries to get exactly one shoot with the given list options.
	// If no shoot or more than one shoot is found it returns an error.
	FindShoot(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.Shoot, error)
	// ListShoots returns all Gardener shoot resources, filtered by a list option
	ListShoots(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.ShootList, error)
	// GetShootClientConfig returns the client config for a shoot
	GetShootClientConfig(ctx context.Context, namespace, name string, memoizedShoot gardencorev1beta1.Shoot, memoizedCASecret corev1.Secret) (clientcmd.ClientConfig, error)

	// GetSecretBinding returns a Gardener secretbinding resource
	GetSecretBinding(ctx context.Context, namespace, name string) (*gardencorev1beta1.SecretBinding, error)

	// GetCloudProfile returns a Gardener cloudprofile resource
	GetCloudProfile(ctx context.Context, name string) (*gardencorev1beta1.CloudProfile, error)

	// GetNamespace returns a Kubernetes namespace resource
	GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	// GetSecret returns a Kubernetes secret resource
	GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error)
	// GetConfigMap returns a Kubernetes configmap resource
	GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error)
	// GetShootOfManagedSeed returns shoot of seed using ManagedSeed resource, nil if not a managed seed
	GetShootOfManagedSeed(ctx context.Context, name string) (*seedmanagementv1alpha1.Shoot, error)

	// RuntimeClient returns the underlying kubernetes runtime client
	// TODO: Remove this when we switched all APIs to the new gardenclient
	RuntimeClient() client.Client
}

type clientImpl struct {
	c client.Client

	// name is a unique identifier of this Garden client
	name string
}

// NewGardenClient returns a new gardenclient
func NewGardenClient(client client.Client, name string) Client {
	return &clientImpl{
		c:    client,
		name: name,
	}
}

func (g *clientImpl) GetProject(ctx context.Context, name string) (*gardencorev1beta1.Project, error) {
	project := &gardencorev1beta1.Project{}
	key := types.NamespacedName{Name: name}

	if err := g.c.Get(ctx, key, project); err != nil {
		return nil, fmt.Errorf("failed to get project %v: %w", key, err)
	}

	return project, nil
}

func (g *clientImpl) GetProjectByNamespace(ctx context.Context, namespace string) (*gardencorev1beta1.Project, error) {
	fieldSelector := client.MatchingFields{gardencore.ProjectNamespace: namespace}
	limit := client.Limit(1)

	projectList, err := g.ListProjects(ctx, fieldSelector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project by namespace: %w", err)
	}

	if len(projectList.Items) == 0 {
		return nil, errors.New("failed to fetch project by namespace")
	}

	return &projectList.Items[0], nil
}

func (g *clientImpl) ListProjects(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.ProjectList, error) {
	projectList := &gardencorev1beta1.ProjectList{}
	if err := g.c.List(ctx, projectList, opts...); err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projectList, nil
}

func (g *clientImpl) GetSeed(ctx context.Context, name string) (*gardencorev1beta1.Seed, error) {
	seed := &gardencorev1beta1.Seed{}
	key := types.NamespacedName{Name: name}

	if err := g.c.Get(ctx, key, seed); err != nil {
		return nil, fmt.Errorf("failed to get seed %v: %w", key, err)
	}

	return seed, nil
}

func (g *clientImpl) ListSeeds(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.SeedList, error) {
	seedList := &gardencorev1beta1.SeedList{}
	if err := g.c.List(ctx, seedList, opts...); err != nil {
		return nil, fmt.Errorf("failed to list seeds: %w", err)
	}

	return seedList, nil
}

func (g *clientImpl) GetShoot(ctx context.Context, namespace, name string) (*gardencorev1beta1.Shoot, error) {
	shoot := &gardencorev1beta1.Shoot{}
	key := types.NamespacedName{Name: name, Namespace: namespace}

	if err := g.c.Get(ctx, key, shoot); err != nil {
		return nil, fmt.Errorf("failed to get shoot %v: %w", key, err)
	}

	return shoot, nil
}

func (g *clientImpl) FindShoot(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.Shoot, error) {
	opts = append(opts, client.Limit(2))

	shootList, err := g.ListShoots(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list shoot clusters: %w", err)
	}

	if len(shootList.Items) == 0 {
		return nil, fmt.Errorf("no shoot found matching the given list options %q", opts)
	}

	var remainingItemCount int64 = 0
	if shootList.RemainingItemCount != nil {
		remainingItemCount = *shootList.RemainingItemCount
	}

	if len(shootList.Items) > 1 || remainingItemCount > 0 {
		return nil, fmt.Errorf("multiple shoots found matching the given list options %q, please target a project or seed to make your choice unambiguous", opts)
	}

	return &shootList.Items[0], nil
}

func (g *clientImpl) resolveListOptions(ctx context.Context, opts ...client.ListOption) error {
	for _, o := range opts {
		if o, ok := o.(resolver); ok {
			if err := o.resolve(ctx, g); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *clientImpl) ListShoots(ctx context.Context, opts ...client.ListOption) (*gardencorev1beta1.ShootList, error) {
	shootList := &gardencorev1beta1.ShootList{}

	if err := g.resolveListOptions(ctx, opts...); err != nil {
		return nil, err
	}

	if err := g.c.List(ctx, shootList, opts...); err != nil {
		return nil, fmt.Errorf("failed to list shoots with list options %q: %w", opts, err)
	}

	return shootList, nil
}

// GetNamespace returns a Kubernetes namespace resource
func (g *clientImpl) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{}
	key := types.NamespacedName{Name: name}

	if err := g.c.Get(ctx, key, namespace); err != nil {
		return nil, fmt.Errorf("failed to get namespace %v: %w", key, err)
	}

	return namespace, nil
}

// GetSecretBinding returns a Gardener secretbinding resource
func (g *clientImpl) GetSecretBinding(ctx context.Context, namespace, name string) (*gardencorev1beta1.SecretBinding, error) {
	secretBinding := &gardencorev1beta1.SecretBinding{}
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := g.c.Get(ctx, key, secretBinding); err != nil {
		return nil, fmt.Errorf("failed to get secretbinding %v: %w", key, err)
	}

	return secretBinding, nil
}

// GetSecret returns a Kubernetes secret resource
func (g *clientImpl) GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := g.c.Get(ctx, key, secret); err != nil {
		return nil, fmt.Errorf("failed to get secret %v: %w", key, err)
	}

	return secret, nil
}

// GetConfigMap returns a Gardener configmap resource
func (g *clientImpl) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: name, Namespace: namespace}

	if err := g.c.Get(ctx, key, cm); err != nil {
		return nil, fmt.Errorf("failed to get configmap %v: %w", key, err)
	}

	return cm, nil
}

func (g *clientImpl) GetShootOfManagedSeed(ctx context.Context, name string) (*seedmanagementv1alpha1.Shoot, error) {
	managedSeed := &seedmanagementv1alpha1.ManagedSeed{}
	key := types.NamespacedName{Namespace: "garden", Name: name} // Currently, managed seeds are restricted to the garden namespace

	if err := g.c.Get(ctx, key, managedSeed); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get managed seed %v: %w", key, err)
	}

	klog.V(1).Infof("using referred shoot %q for seed %q", managedSeed.Spec.Shoot.Name, name)

	return managedSeed.Spec.Shoot, nil
}

func (g *clientImpl) GetSeedClientConfig(ctx context.Context, name string) (clientcmd.ClientConfig, error) {
	if shoot, err := g.GetShootOfManagedSeed(ctx, name); err != nil {
		return nil, err
	} else if shoot != nil {
		return g.GetShootClientConfig(ctx, "garden", shoot.Name, gardencorev1beta1.Shoot{}, corev1.Secret{})
	}

	key := types.NamespacedName{Name: name}

	secret, err := g.GetSecret(ctx, "garden", name+".login")
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		// fallback to deprecated .oidc secret
		var oidcErr error

		secret, oidcErr = g.GetSecret(ctx, "garden", name+".oidc")
		if oidcErr != nil {
			return nil, fmt.Errorf("failed to get kubeconfig for seed %v: %w", key, err) // use original not-found error as cause and ignore error of fallback
		}
	}

	value, ok := secret.Data["kubeconfig"]
	if !ok {
		return nil, fmt.Errorf("invalid kubeconfig secret for seed %v", key)
	}

	config, err := clientcmd.NewClientConfigFromBytes(value)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize kubeconfig for seed %v: %w", key, err)
	}

	return config, nil
}

func (g *clientImpl) GetCloudProfile(ctx context.Context, name string) (*gardencorev1beta1.CloudProfile, error) {
	cloudProfile := &gardencorev1beta1.CloudProfile{}
	key := types.NamespacedName{Name: name}

	if err := g.c.Get(ctx, key, cloudProfile); err != nil {
		return nil, fmt.Errorf("failed to get cloudprofile %v: %w", key, err)
	}

	return cloudProfile, nil
}

// RuntimeClient returns the underlying Kubernetes runtime client
func (g *clientImpl) RuntimeClient() client.Client {
	return g.c
}

type CloudProfile gardencorev1beta1.CloudProfile

func (cp CloudProfile) GetOpenstackProviderConfig() (*openstackv1alpha1.CloudProfileConfig, error) {
	const apiVersion = "core.gardener.cloud/v1alpha1.CloudProfile"

	providerConfig := cp.Spec.ProviderConfig
	if providerConfig == nil {
		return nil, fmt.Errorf("providerConfig of %s %s is empty", apiVersion, cp.Name)
	}

	var cloudProfileConfig *openstackv1alpha1.CloudProfileConfig

	switch {
	case providerConfig.Object != nil:
		var ok bool
		if cloudProfileConfig, ok = providerConfig.Object.(*openstackv1alpha1.CloudProfileConfig); !ok {
			return nil, fmt.Errorf("cannot cast providerConfig of %s %s", apiVersion, cp.Name)
		}

	case providerConfig.Raw != nil:
		cloudProfileConfig = &openstackv1alpha1.CloudProfileConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: openstackv1alpha1.SchemeGroupVersion.String(),
				Kind:       "CloudProfileConfig",
			},
		}
		if _, _, err := decoder.Decode(providerConfig.Raw, nil, cloudProfileConfig); err != nil {
			return nil, fmt.Errorf("cannot decode providerConfig of %s %s", apiVersion, cp.Name)
		}
	default:
		return nil, fmt.Errorf("providerConfig of %s %s contains neither raw data nor a decoded object", apiVersion, cp.Name)
	}

	return cloudProfileConfig, nil
}

// ShootFilter restricts the list operation to the given where condition.
type ShootFilter fields.Set

type resolver interface {
	resolve(context.Context, Client) error
}

type listOptionResolver interface {
	client.ListOption
	resolver
}

var _ listOptionResolver = &ShootFilter{}

func (w ShootFilter) ApplyToList(opts *client.ListOptions) {
	m := fields.Set{}

	for key, value := range w {
		switch key {
		case "metadata.namespace":
			opts.Namespace = value
		default:
			m[key] = value
		}
	}

	if len(m) > 0 {
		opts.FieldSelector = m.AsSelector()
	}
}

func (w ShootFilter) resolve(ctx context.Context, g Client) error {
	if name, ok := w["project"]; ok {
		delete(w, "project")

		project, err := g.GetProject(ctx, name)
		if err != nil {
			return err
		}

		if project.Spec.Namespace == nil || *project.Spec.Namespace == "" {
			return fmt.Errorf("project %q has not yet been assigned to a namespace", name)
		}

		w["metadata.namespace"] = *project.Spec.Namespace
	}

	return nil
}
