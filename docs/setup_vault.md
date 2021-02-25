# Setup Vault for `k8ctx`

Vault can be used as the kubeconfig store for `k8ctx`.
Currently, the [key-value secrets engine](https://www.vaultproject.io/docs/secrets/kv) is supported.
In addition, only one Vault instance can be configured. However, you can configure 
multiple search paths within this one Vault instance.

The `vault` should store kubeconfigs base64 encoded under one or multiple root paths.
These paths will be recursively searched for secrets.

First, enable this path e.g. `landscapes` with the [key-value secrets engine](https://www.vaultproject.io/docs/secrets/kv).

```
vault secrets enable -path=landscapes kv
```

Similar to the filename on a local directory, the `key`
should per default be called `config` or match what is specified in `--kubeconfig-name`.
Create the secret containing the base64 kubeconfig.

```
vault kv put /landscapes/dev/test config=<base64-kubeconfig>
```

You should be able to get back the kubeconfig via

```
vault kv get /landscapes/dev/test
===== Data =====
Key       Value
---       -----
config    <base64 kubeconfig>
```

If you deal with large numbers of changing kubeconfigs, 
it is recommended to setup an automation to sync the kubeconfigs to the vault instance.
You can use [Hooks](../hooks/README.md) to achieve that.

## Configure `k8ctx` to use Vault as Kubeconfig store

Please [see here](kubeconfig_stores.md) on how to configure `k8ctx` either via CLI flags or a `K8ctxConfig` file.

## Additional Considerations

One of the reason this vault integration was built originally is because I
used to synchronize a huge number of kubeconfig files to my local filesystem.
However, for security reasons they should instead be stored in a Vault.

Because in my case there is no central Vault with up to date kubeconfigs,
I run a [local `vault` instance](https://www.vaultproject.io/docs/concepts/dev-server) that uses an encrypted RAM disk for storage.
The kubeconfigs are regularly synced to the local vault with a [custom `k8ctx hook`](../hooks/README.md).
