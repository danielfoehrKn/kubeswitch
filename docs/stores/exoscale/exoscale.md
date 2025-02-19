# Exoscale store

You need to create an IAM role with the following allowed operations:

- list-zones
- list-sks-clusters
- generate-sks-cluster-kubeconfig

## Configuration

The Exoscale store configuration is defined in the `kubeswitch` configuration file. An example configuration is shown below:

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: exoscale
  config:
    exoscaleAPIKey: EXOAPIKEY
    exoscaleSecretKey: THEAPISECRET
  cache:
    kind: filesystem
    config:
      path: ~/.kube/cache
```

The Exoscale store can be used without a filesystem cache but the Exoscale API will create a new Kubeconfig file every time you open switcher.
Therefore, it is recommended to use a filesystem cache.

You can also use multiple organizations. For that you can define `ID` freely, the ID will be shown as prefix in the list if `ShowPrefix` is true (default).

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: exoscale
  id: exoscale-production
  showPrefix: true
  config:
    exoscaleAPIKey: EXOAPIKEY
    exoscaleSecretKey: THEAPISECRET
  cache:
    kind: filesystem
    config:
      path: ~/.kube/cache
- kind: exoscale
  id: exoscale-dev
  config:
    exoscaleAPIKey: EXOAPIKEY
    exoscaleSecretKey: THEAPISECRET
  cache:
    kind: filesystem
    config:
      path: ~/.kube/cache
```
