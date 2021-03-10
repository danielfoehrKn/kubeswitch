# Search Index

The search index is a file in the `state directory` (default: `~/.kube/switch-state/switch.<store>.<id>.index`)
that contains all the discovered kubecontext names over all kubeconfig stores mapped to their kubeconfig file (only the path).
This index is then used instead of querying the kubeconfig store.
This could lead to outdated search results until the index is refreshed again (kubeconfigs might have been deleted or added in the meanwhile).

Using the search index is especially useful when
- dealing with large amounts of Kubeconfigs and querying the Kubeconfig store is slow (e.g. searching a large directory)
- to save request when using a kubeconfig store that queries an API (Vault & Gardener).

Compared to the example in the [hot reload feature](../README.md#hot-reload), see that all the Kubeconfig contexts are available almost instantly.

![demo GIF](../resources/gifs/index-demo.gif)

Per default, the search index is **not** used.

## Enable search index

Enable the search index in the `SwitchConfig` file (per default located in `~/.kube/switch-config.yaml` 
or configured via flag `--config-path`.
The flag has to point to the file, not the directory).

## Enable index for all stores

The field `refreshIndexAfter` determines the time after which the tool should
refresh the index of every configured store.

```
$ cat ~/.kube/switch-config.yaml

kind: SwitchConfig
refreshIndexAfter: 1h
kubeconfigStores: [...many-stores...]
```

## Enable search index for a specific kubeconfig store

The field `refreshIndexAfter` can also be set for only a specific store.

```
$ cat ~/.kube/switch-config.yaml

kind: SwitchConfig
kubeconfigStores:
  - kind: filesystem
    id: unique-1
    refreshIndexAfter: 3h
    paths:
    - "~/.kube/static-kubeconfigs/"
```

### Overwrite global `refreshIndexAfter`

In the example below, the first store overwrites the global `refreshIndexAfter` of _one hour_ and 
sets the `refreshIndexAfter` to _three hours_ instead.
Store two uses the default.

```
$ cat ~/.kube/switch-config.yaml

kind: SwitchConfig
refreshIndexAfter: 1h
kubeconfigStores:
  - kind: filesystem
    id: unique-1
    refreshIndexAfter: 3h
    paths:
    - "~/.kube/static-kubeconfigs/"
  - kind: filesystem
    id: unique-2
    paths:
    - "~/.kube/next-kubeconfigs/"
```
