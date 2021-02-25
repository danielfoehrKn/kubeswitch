# K8ctx

![Latest GitHub release](https://img.shields.io/github/v/release/danielfoehrkn/k8ctx.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielfoehrKn/k8ctx)](https://goreportcard.com/badge/github.com/danielfoehrKn/k8ctx)
[![Build](https://github.com/danielfoehrKn/k8ctx/workflows/Build/badge.svg)](https://github.com/danielfoehrKn/k8ctx/actions?query=workflow%3A"Build")
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


The `kubectx` for operators.
`k8ctx` takes kube context switching to the next level,
catering to operators of large scale Kubernetes installations.
Designed as a [drop-in replacement](#difference-to-kubectx) for [kubectx](https://github.com/ahmetb/kubectx).

## Highlights

- **Configurable Kubeconfig store**
  - Local filesystem
  - Hashicorp Vault
- **Unified search over multiple directories, paths, files and Kubeconfig stores**
  - Search over the Kubeconfigs from multiple Kubeconfig stores and configured paths (e.g, directories on the local filesystem and paths in Vault).
- **Easy navigation**
  - Define alias names for contexts without changing the underlying Kubeconfig
  - Switch to any previously used context from the history
- **Efficient search**
  - Stores a pre-computed index of kube-context names to speed up future searches
  - Recursive directory search
  - Hot reload capability (adds Kubeconfigs to the search on the fly - especially useful when initially searching large directories)
- **Improved search experience** when dealing with many Kubeconfigs
  - Live preview of the Kubeconfig file (**sanitized from credentials**)
  - Kubeconfig context names are easily identifiable. The `context` is prefixed with the (immediate) parent folder name to allow to easily find the context you are looking for. 
  - Same fuzzy search capability known from `kubectx`
- **Terminal Window isolation**
  - Each terminal window can target a different cluster (does not override the current-context in a shared Kubeconfig)
  - Each terminal window can target the same cluster and set a [different namespace preference](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/#setting-the-namespace-preference)
    e.g., using the tool [kubens](https://github.com/ahmetb/kubectx)
- **Extensibility** 
  - Integrate custom functionality using [Hooks](./hooks/README.md) (comparable with Git pre-commit hooks).
  - Build your own integration e.g., sync Kubeconfig files of clusters from Git or remote systems like [Gardener](https://gardener.cloud/).

![demo GIF](resources/gifs/k8ctx-demo.gif)

## Non-goals

- Anything else but efficient searching and switching of Kubeconfig contexts. 
  This includes the ability to change Kubernetes namespaces. 
  Use another tool e.g. [kubens](https://github.com/ahmetb/kubectx/blob/master/kubens).

## Installation

### Option 1 - Homebrew

Mac and Linux users can install both the `k8ctx` script and the `switcher` binary with `homebrew`. 
```
$ brew install danielfoehrkn/k8ctx/k8ctx
```

Source the `k8ctx` script from the `homebrew` installation path.
```
$ source $HOMEBREW_CELLAR/k8ctx/v0.2.0/k8ctx.sh
```

Updating the version of the `k8ctx` utility via `brew` (e.g changing from version 0.1.0 to 0.2.0) requires to change the sourced path.

### Option 2 - Manual Installation

#### From Source

```
$ go get github.com/danielfoehrKn/k8ctx
```

From the repository root run `make build-switcher`.
This builds the binaries to `/hack/k8ctx/`.
Copy the build binary for your OS / Architecture to e.g. `/usr/local/bin`
and source the k8ctx script from `/hack/k8ctx/k8ctx.sh`.

#### Github Releases

Download the k8ctx script and the switcher binary for your OS / Architecture (darwin / linux).
```
# grab pre-compiled switcher binary for your OS/architecture
OS=linux #darwin
wget https://github.com/danielfoehrKn/k8ctx/releases/download/0.2.0/switcher_${OS}_amd64.tar.gz
tar -zxvf switcher_${OS}_amd64.tar.gz
cp switcher_${OS}_amd64 /usr/local/bin/switcher
rm switcher_${OS}_amd64.tar.gz

# grab k8ctx script
wget https://github.com/danielfoehrKn/k8ctx/releases/download/0.2.0/k8ctx.tar.gz
tar -zxvf k8ctx.tar.gz
cp k8ctx.sh /usr/local/bin/k8ctx.sh
rm k8ctx.tar.gz
```

Source `k8ctx.sh` in `.bashrc`/`.zsh` via:
```
$ source /usr/local/bin/k8ctx.sh
```
### Command completion

Please [see here](docs/command_completion.md) how to install command completion for bash and zsh shells.
This completes both the `k8ctx` commands as well as the context names.

## Usage 

```
$ k8ctx -h

Usage:
  k8ctx [flags]
  k8ctx [command]

Available Commands:
  <context-name>  Switch to context name provided as first argument
  history, h      Switch to any previous context from the history (short: h)
  hooks           Runs configured hooks
  alias           Create an alias for a context. Use <ALIAS>=<CONTEXT_NAME> (<ALIAS>=. to rename current-context to <ALIAS>). To list all use "alias ls" and to remove an alias use "alias rm <ALIAS>"
  list-contexts   List all available contexts without fuzzy search
  clean           Cleans all temporary kubeconfig files
  -               Switch to the previous context from the history
  -d <NAME>       Delete context <NAME> ('.' for current-context) from the local kubeconfig file.
  -c, --current   Show the current context name
  -u, --unset     Unset the current context from the local kubeconfig file

Flags:
      --config-path string         path on the local filesystem to the configuration file. (default "~/.kube/k8ctx-config.yaml")
      --kubeconfig-name string     only shows Kubeconfig files with this name. Accepts wilcard arguments "*" and "?". Defaults to "config". (default "config")
      --kubeconfig-path string     path to be recursively searched for Kubeconfig files. Can be a file or directory on the local filesystem or a path in Vault. (default "~/.kube/config")
      --show-preview               show preview of the selected Kubeconfig. Possibly makes sense to disable when using vault as the Kubeconfig store to prevent excessive requests against the API. (default true)
      --state-directory string     path to the local directory used for storing internal state. (default "~/.kube/k8ctx-state")
      --store string               the backing store to be searched for Kubeconfig files. Can be either "filesystem" or "vault" (default "filesystem")
      --vault-api-address string   the API address of the Vault store. Overrides the default "vaultAPIAddress" field in the K8ctxConfig. This flag is overridden by the environment variable "VAULT_ADDR".
      -h, --help                   help about any command
```

Just type `k8ctx` to search over the context names defined in the default Kubeconfig file `~/.kube/config`
or from the environment variable `KUBECONFIG`.

To recursively **search over multiple directories, files and Kubeconfig stores**, please see the [documentation](docs/kubeconfig_stores.md) 
to set up the necessary configuration file.

## Kubeconfig stores

Two Kubeconfig stores are supported: `filesystem` and `vault`.
The local filesystem is the default store and does not require any additional setup.
However, if you intend to search for all Kubeconfig context/files in the `~/.kube` directory, 
please [first consider this](docs/kubeconfig_stores.md#additional-considerations).

To search over multiple directories and setup Kubeconfig stores (such as Vault), [please see here](docs/kubeconfig_stores.md).

## Transition from Kubectx

Offers a smooth transition as `k8ctx` is a 
drop-in replacement for _kubectx_.
You can set an alias and keep using your existing setup.
```
  alias kubectx='k8ctx'
  alias kctx='k8ctx'
```

However, that does not mean that `k8ctx` behaves exactly like `kubectx`. 
Please [see here](#difference-to-kubectx) to read about some main differences to kubectx.

## Alias

An alias for any context name can be defined. 
An alias **does not modify** or rename the context in the kubeconfig file (unlike `kubectx`), 
instead it is just injected for the search.

Define an alias.

```
$ k8ctx alias mediathekview=gke_mediathekviewmobile-real_europe-west1-c_mediathekviewmobile
```

It is also possible to use `k8ctx alias <alias>=.` to create an alias for the current context.

See the created alias
```
$ k8ctx alias ls
+---------------+-------------------------------------------------------------------------------+
| ALIAS         | CONTEXT                                                                       |
+---------------+-------------------------------------------------------------------------------+
| mediathekview | mediathekview/gke_mediathekviewmobile-real_europe-west1-c_mediathekviewmobile |
+---------------+-------------------------------------------------------------------------------+
| TOTAL         | 1                                                                             |
+---------------+-------------------------------------------------------------------------------+ 
```

Remove the alias

```
$ k8ctx alias rm mediathekview
```

### Search Index

See [here](docs/search_index.md) how to use a search index to speed up search operations.
Using the search index is especially useful when
- dealing with large amounts of Kubeconfigs and querying the Kubeconfig store is slow (e.g. searching a large directory)
- when using a remote systems (such as Vault) as the Kubeconfig store to increase search speed, reduce latency and save API requests

### Hot Reload

For large directories with many Kubeconfig files, the Kubeconfigs are added to the search set on the fly.
For smaller directory sizes, the search feels instantaneous.

![demo GIF](resources/gifs/hot-reload.gif)

## Improved Search Experience 

The `k8ctx` tool just recursively searches through a specified path in a Kubeconfig store for Kubeconfig files matching a name.
The path layout presented below is purely optional.

When dealing with a large amount of Kubeconfig names, the `context` names are not necessarily unique or convey a meaning (especially when they are generated names).
To circumvent that issue, the fuzzy search includes the parent folder name.
This way, the directory layout can actually convey information useful for the search.

To exemplify this, look at the directory layout below. 
Each Kubernetes landscape (called `dev`, `canary` and `live`) have their own directory containing the Kubeconfigs of the Kubernetes clusters on that landscape.
Every `Kubeconfig` is named `config`.

```
$ tree .kube/my-path
.kube/my-path
├── canary
│   └── config
├── dev
│   ├── config
│   └── config-tmp
└── live
    └── config
```
This is how the search looks like for this directory.
The parent directory name is part of the search.

![demo GIF](resources/gifs/search-show-parent-folder.png)

Limitation: Please make sure that within one directory, there are no kubeconfig files that have the same context names.

### Extensibilty 

Customization is possible by using `Hooks` (think Git pre-commit hooks). 
Hooks can call an arbitrary executable or execute commands at a certain time (e.g every 6 hours) prior to the search via `k8ctx`.
For more information [take a look here](./hooks/README.md).

### Difference to kubectx

`kubectx` is great when dealing with few Kubeconfig files - however lacks support when
operating large Kubernetes installations where clusters spin up on demand,
have cryptic context names or are stored in various kubeconfig stores (e.g., Vault).

`k8ctx` is build for a world where Kubernetes clusters are [treated as cattle, not pets](https://devops.stackexchange.com/questions/653/what-is-the-definition-of-cattle-not-pets).
This has implications on how Kubeconfig files are managed. 
`k8ctx` is fundamentally designed for the modern Kubernetes operator of large dynamic Kubernetes 
installations with possibly thousands of Kubeconfig files in [various locations](docs/kubeconfig_stores.md).

Has build-in
 - convenience features (terminal window isolation, context history, [context aliasing](#alias), [improved search experience](#improved-search-experience), sanitized kubeconfig preview);
 - advanced search capabilities (search index, hot reload);
 - as well as integration points with external systems ([hooks](hooks/README.md)).


In addition, `k8ctx` is a drop-in replacement for _kubectx_.
You can set an alias and keep using your existing setup.
```
  alias kubectx='k8ctx'
  alias kctx='k8ctx'
```

However, that does not mean that `k8ctx` behaves exactly like `kubectx`.

**Alias Names**

`kubectx` supports renaming context names using `kubectx <NEW_NAME>=<NAME>`.
Use `k8ctx <NEW_NAME>=<NAME>` to create an alias.
An alias **does not modify** or rename the context in the kubeconfig file. 
It is just a local configuration that can be removed again via `k8ctx alias rm <NAME>`.
Directly modifying the Kubeconfig is problematic: 
 - Common tooling might be used across the team which needs to rely on predictable cluster naming conventions
 - Modifying the file is not always possible e.g., when the Kubeconfig is actually stored in a Vault
 - No easy way to revert the alias or see aliases that are currently in use

**Terminal Window isolation**

`kubectx` directly modifies the kubeconfig file to set the current context.
This has the disadvantage that every other terminal using the same 
Kubeconfig file (e.g, via environment variable _KUBECONFIG_) will also be affected and change the context.

A guideline of `k8ctx` is to not modify the underlying Kubeconfig file.
Hence, a temporary copy of the original Kubeconfig file is created and used to modify the context.
This way, each terminal window works on its own copy of the Kubeconfig file and cannot interfere with each other.

### Future Plans

- Cleanup temporary kubeconfig files after the terminal session ended (instead of using `k8ctx clean`)
- Act as a credential helper for kubeconfig files to inject the credentials from the backing store
- Support more storage backends (e.g local password managers)
