# Configure Kubeconfig stores

`k8ctx` can recursively search over multiple directories, files and Kubeconfig stores.
If you neither provide a flag or a `K8ctxConfig` file, it will default to the file `~/.kube/config`.

This is configurable using [CLI flags](#configure-via-cli-flags)
or the [`K8ctxConfig` file](#configure-via-k8ctxconfig-file).
To search over **multiple** directories and kubeconfig stores,
the `K8ctxConfig` file has to be configured properly.

To use Vault as a Kubeconfig store, please first check [this document](setup_vault.md).

## Configure via CLI flags

### Search a directory on the local filesystem

To search the directory `~/.kube/k8ctx` on the local filesystem  use the flag `--kubeconfig-path`.
You do not need to specify  `--store filesystem` as this is the default.
This flag only supports **one path**, to supply multiple directories, use the [`K8ctxConfig`file](#configure-via-k8ctxconfig-file).

```
k8ctx --kubeconfig-path ~/.kube/my-path
```

### Search a file on the local filesystem

The flag `--kubeconfig-path` also accepts a file as an argument.

```
k8ctx --kubeconfig-path ~/.kube/my-path/my-kubeconfig-file.yaml
```

### Search a path in Vault

Using `vault` requires the following
- set either the environment variable `VAULT_ADDR` or the k8ctx command line flag `--vault-api-address` to the API endpoint of your vault instance.
- make sure the file `~/.vault-token` is set (automatically created via the `vault` CLI) and contains the token for your vault server.
  Alternatively set the environment variable `VAULT_TOKEN`.
- set the k8ctx command line flag `--kubeconfig-path` to the desired path of the vault secrets engine.
  e.g., if the kubeconfigs are stored in vault under `landscapes/dev/cluster-1` and `landscapes/canary/cluster-1`
  then set the flag value to `landscapes`
- set the k8ctx command line flag `--store vault`

Example usage:

```
$ k8ctx --kubeconfig-path landscapes --store vault  --vault-api-address http://127.0.0.1:8200
```

## Configure via `K8ctxConfig` file

The `K8ctxConfig` file is expected to be in the default location
on the local filesystem at `~/.kube/k8ctx-config.yaml` or set via flag `--config-path`.
Example config files can be found [here](../resources/demo-config-files).

### Search multiple directories on the local filesystem

`k8ctx` can search over **multiple** directories and combine the search results.
The `path` field accepts both directories and filepaths.

```
$ cat ~/.kube/k8ctx-config.yaml
kind: K8ctxConfig
kubeconfigPaths:
  - path: "~/.kube/my-other-kubeconfigs/"
    store: filesystem
  - path: "~/.kube/my-next-kubeconfigs/"
    store: filesystem
  - path: "~/.kube/config" # filepath
    store: filesystem
```

### Configure Vault

Using `vault` also requires setting the API endpoint of the Vault instance.
Either set via
- the environment variable `VAULT_ADDR` (overrides other settings)
- the k8ctx command line flag `--vault-api-address` [see here](#search-a-path-in-vault).
- the `K8ctxConfig` file

```
kind: K8ctxConfig
vaultAPIAddress: "http://127.0.0.1:8200"
```

### Search over Kubeconfigs in Vault

`k8ctx` can search over **multiple** paths in Vault and combine the search results.

```
$ cat ~/.kube/k8ctx-config.yaml
kind: K8ctxConfig
vaultAPIAddress: "http://127.0.0.1:8200"
kubeconfigPaths:
  - path: "landscapes/dev"
    store: vault
  - path: "landscapes/canary"
    store: vault
```

Before executing `k8ctx` to search in Vault, make sure the file `~/.vault-token` is set (automatically created via the `vault` CLI)
and contains the token for your vault server.
Alternatively set the environment variable `VAULT_TOKEN`.

### Combined search on the local filesystem and Vault

Just provide paths for both the local filesystem as well as Vault.

```
$ cat ~/.kube/k8ctx-config.yaml
kind: K8ctxConfig
vaultAPIAddress: "http://127.0.0.1:8200"
kubeconfigPaths:
  - path: "landscapes/dev"
    store: vault
  - path: "landscapes/canary"
    store: vault
  - path: "~/.kube/local-kubeconfigs/"
    store: filesystem
```

## Using both CLI and `K8ctxConfig` file

- The flag `--vault-api-address` takes presendence over the config field `vaultAPIAddress`.
- Specifying `--kubeconfig-path` and `--store` plus `kubeconfigPaths` in the config file
  causes a search over all of those paths combined.

# Additional considerations

To speed up the fuzzy search on the local filesystem,
I would recommend putting all the Kubeconfig files into a single directory containing only Kubeconfig files.
This is because the default `~/.kube` directory contains a bunch of other files
that have to be filtered out and thus slowing down the search.

To do this, create a `k8ctx` alias via `--Kubeconfig-path` pointing
to this directory or setup the kubeconfig path in the `K8ctxConfig`.

