# Search Index

The search index is nothing else than a file in the switch `state directory`
that contains all the discovered kubecontext names over all kubeconfig stores mapped to their kubeconfig file (only the path).
The index is stored in a file in the state directory (default: `~/.kube/switch-state/switch.<store>.index`)
This index is then used instead of querying the kubeconfig store.

Using the search index is especially useful when
- dealing with large amounts of Kubeconfigs and querying the Kubeconfig store is slow (e.g. searching a large directory)
- when using vault as the Kubeconfig store to save requests against the Vault API

Per default, the search index is **not** used.
Enable the search index in the `SwitchConfig` file (per default located in `~/.kube/switch-config.yaml` or configured via flag `--config-path`. The flag has to point to the file, not the directory).
The field `KubeconfigRediscoveryInterval` determines the time after which the tool should
refresh its index against the configured Kubeconfig store.


```
$ cat ~/.kube/switch-config.yaml

kind: SwitchConfig
KubeconfigRediscoveryInterval: 1h
```

Compared to the example in the [hot reload feature](../README.md#hot-reload), see that all the Kubeconfig contexts are available almost instantly.

![demo GIF](../resources/pictures/index-demo.gif)
