# The case of k8ctx

Managing a handful of Kubeconfig files and contexts is straight forward and well-supported by existing tools.
You can use `kubectl config view --flatten`, define aliases or use `kubectx`.

Meanwhile, Kubernetes adoption has gone through the roof and large scale Kubernetes landscapes 
with hundreds to thousands of clusters are not uncommon.
On top of that, projects like [Gardener](https://gardener.cloud/), [SIG Cluster-API](https://github.com/kubernetes-sigs/cluster-api) or [Rancher](https://github.com/rancher/rancher) make it easy to spin up 
and maintain clusters at scale.
Many have moved on from treating Kubernetes clusters
[as pets to treating them as cattle](https://devops.stackexchange.com/questions/653/what-is-the-definition-of-cattle-not-pets).
Hence, there is a need for a tool that is build for this reality.

The idea of `k8ctx` is to enable seamless interaction with Kubeconfig files
for operators of large scale Kubernetes landscapes.
No matter if they are stored on disk, in an Enterprise Vault, are spread in different directories 
or need to be synchronized from a remote system.

`k8ctx` offers
- Convenience features (terminal window isolation, context history, [context aliasing](https://github.com/danielfoehrKn/k8ctx#alias), [improved search experience](https://github.com/danielfoehrKn/k8ctx#improved-search-experience), sanitized Kubeconfig preview);
- advanced search capabilities (search index, hot reload, unified search over all kubeconfig stores);
- as well as custom integration points with external systems (see [hooks](https://github.com/danielfoehrKn/k8ctx/tree/master/hooks/README.md)).

To not break existing setups, `k8ctx` is a drop-in replacement for _kubectx_.
You can just set an alias.

```
  alias kubectx='k8ctx'
  alias kctx='k8ctx'
```

These are all the currently available commands of the CLI:
```
$ k8ctx -h

Usage:
  k8ctx [flags]
  k8ctx [command]

Available Commands:
  <context-name>  Switch to context name provided as first argument
  history, h      Switch to any previous context from the history (short: h)
  hooks           Runs configured hooks
  alias           Create an alias for a context. Use <ALIAS>=<CONTEXT_NAME> (<ALIAS>=. to rename current-context to <ALIAS>). 
                  To list all use "alias ls" and to remove an alias use "alias rm <ALIAS>"
  list-contexts   List all available contexts without fuzzy search
  clean           Cleans all temporary kubeconfig files
  -               Switch to the previous context from the history
  -d <NAME>       Delete context <NAME> ('.' for current-context) from the local kubeconfig file.
  -c, --current   Show the current context name
  -u, --unset     Unset the current context from the local kubeconfig file
```

Future plans are to act as an authentication helper for Kubeconfig files 
to inject the credentials from the backing store
and to support more storage backends on top of Vault and the local filesystem.

This should not be a long ramble, so I invite you to check out [k8ctx on Github](https://github.com/danielfoehrKn/k8ctx) 
with more information or head straight to the [installation section](https://github.com/danielfoehrKn/k8ctx#installation).
Of course contributions are more than welcome.
Cheers,
Daniel