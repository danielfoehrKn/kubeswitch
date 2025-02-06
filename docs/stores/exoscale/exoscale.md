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
