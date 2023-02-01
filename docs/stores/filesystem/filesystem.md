# Filesystem store

Search for kubeconfig files on the local filesystem.

The filesystem store is the default store for `kubeswitch`.
That means the CLI flag `--store`  defaults to `filesystem`.
Therefore, you can search through a specific directory on the local filesystem for 
kubeconfig files by only specifying the flag  --kubeconfig-path.

## Configuration

The filesystem store can be configured using [CLI flags](#cli-flags) as well as the [with the `SwitchConfig` file](#set-up-the-configuration-file-switchconfig).

## CLI flags

### Search a directory on the local filesystem

To search the directory `~/.kube/switch` on the local filesystem  use the flag `--kubeconfig-path`.
You do not need to specify `--store filesystem` as the flag `--store`  defaults to `filesystem`.
This flag only supports **one path**, to supply multiple directories, use the [`SwitchConfig`file](#set-up-the-configuration-file-switchconfig).

```
switch --kubeconfig-path ~/.kube/my-path
```

### Search a file on the local filesystem

The flag `--kubeconfig-path` also accepts a file as an argument.

```
switch --kubeconfig-path ~/.kube/my-path/my-kubeconfig-file.yaml
```

## Set up the configuration file `SwitchConfig`

### Search multiple directories on the local filesystem

`kubeswitch` can search over **multiple** directories and combine the search results.
The `path` field accepts both directories and filepaths.

```
$ cat ~/.kube/switch-config.yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: filesystem
  paths:
  - ~/.kube/my-other-kubeconfigs/
  - ~/.kube/my-next-kubeconfigs/
  - ~/.kube/config
```

Configuring more than one store of kind `filesystem` is possible. 
This makes sense if you use an `index` and want to define a different refresh interval per filepath.
Please take a look [here](../../kubeconfig_stores.md#combined-search-over-multiple-stores) for more information.
