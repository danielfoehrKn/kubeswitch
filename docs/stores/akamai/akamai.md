# Akamai store

To use the Akamai store a token should be created [on linode's website](https://cloud.linode.com/profile/tokens)

In order to create this token you also need to specify the scope.
The required permissions for this plugin to work are the following:
- `read/write` for Kubernetes

## Configuration

The Akamai store configuration is defined in the `kubeswitch` configuration file.
An example configuration is shown below:

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: akamai
  config:
    linode_token: "your-linode-token"
```

`linode_token` can be ignored if set with the environment variable `LINODE_TOKEN`.
