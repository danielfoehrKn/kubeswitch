# Cluster API (capi) store

To use the Cluster API (capi) store a kubeconfig file should be created for the management cluster.

## Configuration

The Cluster API store configuration is defined in the `kubeswitch` configuration file.

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: capi
  config:
    # Optionally specify a kubeconfigPath for a management cluster, 
    # if not specified your current kube context will be searched for any CAPI clusters
    kubeconfigPath: "/home/user/.kube/management.config" 
```
