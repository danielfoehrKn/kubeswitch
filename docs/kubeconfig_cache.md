# Configure Kubeconfig stores

kubeswitch supports multiple remote config stores. Every time kubeswitch is called, the kubeconfig file is downloaded again.

With the help of the cache, the kubeconfig file can be temporarily stored in the file system. The next time the file is requested, the file is made available from the local cache.

The search_index is not affected. The [search index](search_index.md) must be configured in order to also prevent these requests to remote.

To enable the cache you need to use the `SwitchConfig` file.

## General Cache configuration

For each kubeconfig store it is possible to add a cache configuration section.
At the moment only the kind `filesystem` is supported

```
$ cat ~/.kube/switch-config.yaml
kind: SwitchConfig
version: v1alpha1
refreshIndexAfter: 10h
kubeconfigStores:
- kind: vault
  id: example
  [...]
  cache:
    kind: filesystem
    config:
      path: ~/kubetest/cache
```

Each downloaded kubeconfig file will be stored in the path for the cache.
It is safe to point multiple caches to the same directory as the filename for the cache includes the kind and id of the corresponding store.
Example: 08df4a6d672ebac1a7d0657e7800f264.vault.example.cache

Note: The file is not encrypted. The directory should be protected.


### Clean up cache

The files are cached forever. The switch clean command will delete all files of every configured cache.

```
$ switch clean

Cleaned 3 files from temporary kubeconfig directory.
Cleaned 15 files of vault.example cache
```


