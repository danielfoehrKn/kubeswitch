# Rancher store

To use the Rancher store an API token is required. The token can be created in the Rancher UI.
Searching over multiple Rancher instances is supported, but may require `showPrefix` to be set to `true` in the `SwitchConfig` file to avoid name collisions.

## Configuration

The Rancher store configuration is defined in the `kubeswitch` configuration file. An example configuration is shown below:

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: rancher
  id: rancher
  config:
    rancherAPIAddress: https://rancher.yourdomain.com/v3
    rancherToken: token-12abc:bmjlzslas......x4hv5ptc29wt4sfk
  cache:
    kind: filesystem
    config:
      path: ~/.kube/cache
```

The Rancher store can be used without a filesystem cache but the Rancher API will create a new Kubeconfig file (and token) every time you switch to one of the Rancher contexts.
Therefore, it is recommended to use a filesystem cache.
