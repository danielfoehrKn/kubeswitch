# Rancher store

To use Rancher as a Kubeconfig store, please create an API token from the rancher UI first.
Searching over multiple Rancher instances is supported but in this case `showPrefix` must be set to true to prevent conflicts.

# Configure Rancher in the SwitchConfig file

Below there is an example configuration for Rancher in the `SwitchConfig` file.

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: rancher
  id: rancher
  config:
    rancherAddress: https://rancher.yourdomain.com/v3
    rancherToken: token-12abc:bmjlzslas......x4hv5ptc29wt4sfk
  cache:
    kind: filesystem
    config:
      path: ~/.kube/cache
```

The rancher store can work without filesystem cache but rancher api will create by default a new kubeconfig file everytime the context is selected without invalidating the old one. Therefore it is recommended to activate the filesystem [cache](../../kubeconfig_cache.md). 