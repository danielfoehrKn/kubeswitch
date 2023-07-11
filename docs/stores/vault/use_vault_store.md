# Use the Vault store

To use Vault as a Kubeconfig store, please first [setup Vault](setup_vault.md).
Currently, searching over multiple Vault instances is not supported.

## Configuration

Setting the environment variable `VAULT_ADDR` overwrites the configuration from 
the CLI or the switch config.

Make sure the file `~/.vault-token` is set (automatically created via the `vault` CLI) and 
contains the token for your vault server.
Alternatively set the environment variable `VAULT_TOKEN`.
Currently, you cannot specify a different token for each store, so even though you can define multiple
vault stores in the `SwitchConfig` file, they will all use the same credentials.

## CLI

Using `vault` via the CLI requires the following
- optionally set the command line flag `--vault-api-address` to the API endpoint of your vault instance.
- set the switch command line flag `--kubeconfig-path` to the desired path of the vault secrets engine.
  e.g., if the kubeconfigs are stored in vault under `landscapes/dev/cluster-1` and `landscapes/canary/cluster-1`
  then set the flag value to `landscapes`
- set the switch command line flag `--store vault`

Example usage:

  ```
  $ switch --kubeconfig-path landscapes --store vault  --vault-api-address http://127.0.0.1:8200
  ```

### Configure Vault in the SwitchConfig file

Below there is an example configuration for Vault in the `SwitchConfig` file.

```
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: vault
  paths:
  - my-vault-path
  - my-next-vault-path
  config:
    vaultAPIAddress: http://127.0.0.1:8200
```

### Configure Vault KV Secrets engine v2 in SwitchConfig file

If Vault is setup with a KV secrets engine v2, below is an example configuration for using Vault in the `SwitchConfig` file.

```
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: vault
  showPrefix: false
  paths:
  - "shared/kubernetes"
  config:
    vaultAPIAddress: "https://address.to.vault"
    vaultKeyKubeconfig: "kubeconfig"
    vaultEngineVersion: "v2"
```

`vaultKeyKubeconfig` specifies which key in the secret the kubeconfig is saved under. Defaults to `config`.

`vaultEngineVersion` specifies which Vault secrets engine to use. Defaults to `v1`.

Combining `vault` with `cache` means that the fetched kubeconfig's from Vault are cached locally, and thus limiting the number of requests to Vault significant:

```
kind: SwitchConfig
version: v1alpha1
refreshIndexAfter: 12h
kubeconfigStores:
- kind: vault
  showPrefix: false
  paths:
  - "shared/kubernetes"
  config:
    vaultAPIAddress: "https://address.to.vault"
    vaultKeyKubeconfig: "kubeconfig"
    vaultEngineVersion: "v2"
  cache:
    kind: filesystem
    config:
      path: ~/.kube/cache/switch
```

By using `refreshIndexAfter` you can force a refresh of the index. In this case every 12th hour.

Note: Make sure that the folder mentioned under `cache.path` is present, otherwise it will not work.