# Use the Vault kubeconfig store

Vault can be used as the kubeconfig store for `switch`.
Currently, the [key-value secrets engine](https://www.vaultproject.io/docs/secrets/kv) is supported.

## Setup vault

The `vault` should store all base64 encoded kubeconfigs under one root path.
This path will be recursively searched for secrets.


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

## Configure `switch` to use the `vault`

Using `vault` requires the following
- set either the environment variable `VAULT_ADDR` or the switch command line flag `--vault-api-address` to the API endpoint of your vault instance.
- make sure the file `~/.vault-token` is set (automatically created via the `vault` CLI) and contains the token for your vault server. Alternatively set the environment variable `VAULT_TOKEN`.
- set the switch command line flag `--kubeconfig-path` to the root path of the vault secrets engine. E.g if the kubeconfigs are stored in vault under `landscapes/dev/cluster-1` and `landscapes/canary/cluster-1` then set the flag value to `landscapes`
- set the switch command line flag `--store vault`

Example usage:

`
switch --kubeconfig-path landscapes --store vault  --vault-api-address http://127.0.0.1:8200
`

I recommend creating an alias for this command.

## Additional Considerations

One of the reason I build this vault integration was that I
used to synchronize a huge number of kubeconfig files to my local filesystem.
However, for security reasons they should instead be stored in a Vault.

Because in my case there is no central Vault with up to date kubeconfigs,
I run a [local `vault` instance](https://www.vaultproject.io/docs/concepts/dev-server) that uses an encrypted RAM disk for storage.
The kubeconfigs are regularly synced to the local vault with a [custom `switch hook`](../hooks/README.md).
