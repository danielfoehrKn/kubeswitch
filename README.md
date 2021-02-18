# Kubeswitch

The `kubectx` for operators.

Kubeswitch (short: `switch`) is a tiny standalone tool, 
designed to conveniently switch between the context of thousands of `Kubeconfig` files.

## Highlights

- **Configurable Kubeconfig store**
  - Local filesystem
  - Hashicorp Vault
- **Search over multiple directories, paths, files and Kubeconfig stores**
  - Unified search over the Kubeconfigs from multiple Kubeconfig stores and configured paths (e.g, directories on the local filesystem and paths in Vault).
- **Efficient search**
  - stores a pre-computed index of kube-context names to speed up future searches
  - recursive directory search
  - hot reload capability (adds Kubeconfigs to the search on the fly - especially useful when initially searching large directories)
- **Improved search experience** when dealing with many Kubeconfigs
  - Live preview of the Kubeconfig file (**sanitized from credentials**).
  - Kubeconfig context names are easily identifiable. The `context` is prefixed with the (immediate) parent folder name to allow to easily find the context you are looking for. 
  - Same fuzzy search capability known from `kubectx`
- **Terminal Window isolation**
  - Each terminal window can target a different cluster (does not override the current-context in a shared Kubeconfig).
  - Each terminal window can target the same cluster and set a [different namespace preference](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/#setting-the-namespace-preference) 
  e.g via the tool [kubens](https://github.com/ahmetb/kubectx).
- **Extensibility** 
  - Integrate custom functionality using [Hooks](./hooks/README.md) (comparable with Git pre-commit hooks).

![demo GIF](resources/pictures/switch-demo-large.gif)

## Non-goals

- Anything else but efficient searching and switching of Kubeconfig contexts. 
  This includes the ability to change Kubernetes namespaces. Use another tool e.g. [kubens](https://github.com/ahmetb/kubectx/blob/master/kubens).

## Installation

### Option 1 - Homebrew

Mac and Linux users can install both the `switcher` tool and the `switch` script with `homebrew`. 
```
$ brew install danielfoehrkn/switch/switch
```

Source the `switch` script from the `homebrew` installation path.
```
$ source $HOMEBREW_CELLAR/switch/v0.2.0/switch.sh
```

Updating the version of the `switch` utility via `brew` (e.g changing from version 0.1.0 to 0.2.0) requires to change the sourced path.

### Option 2 - Manual Installation

#### From Source

```
$ go get github.com/danielfoehrKn/kubeswitch
```

From the repository root run `make build-switcher`.
This builds the binaries to `/hack/switch/`.
Copy the build binary for your OS / Architecture to e.g. `/usr/local/bin`
and source the switch script from `/hack/switch/switch.sh`.

#### Github Releases

Download the switch script and the switcher binary for your OS / Architecture (darwin / linux).
```
# grab pre-compiled switcher binary for your OS/architecture
OS=linux #darwin
wget https://github.com/danielfoehrKn/kubeswitch/releases/download/0.2.0/switcher_${OS}_amd64.tar.gz
tar -zxvf switcher_${OS}_amd64.tar.gz
cp switcher_${OS}_amd64 /usr/local/bin/switcher
rm switcher_${OS}_amd64.tar.gz

# grab switch script
wget https://github.com/danielfoehrKn/kubeswitch/releases/download/0.2.0/switch.tar.gz
tar -zxvf switch.tar.gz
cp switch.sh /usr/local/bin/switch.sh
rm switch.tar.gz
```

Source `switch.sh` in `.bashrc`/`.zsh` via:
```
$ source /usr/local/bin/switch.sh
```
### Command completion

Please [see here](docs/command_completion.md) how to install command completion for bash and zsh shells.

## Usage 

```
$ switch -h

The kubectx for operators.

Usage:
  switch [flags]
  switch [command]

Available Commands:
  clean       Cleans all temporary Kubeconfig files
  help        Help about any command
  hooks       Runs configured hooks

Flags:
      --config-path string         path on the local filesystem to the configuration file. (default "~/.kube/switch-config.yaml")
  -h, --help                       help for switch
      --kubeconfig-name string     only shows Kubeconfig files with this name. Accepts wilcard arguments "*" and "?". Defaults to "config". (default "config")
      --kubeconfig-path string     path to be recursively searched for Kubeconfig files. Can be a file or directory on the local filesystem or a path in Vault. (default "~/.kube/config")
      --show-preview               show preview of the selected Kubeconfig. Possibly makes sense to disable when using vault as the Kubeconfig store to prevent excessive requests against the API. (default true)
      --state-directory string     path to the local directory used for storing internal state. (default "~/.kube/switch-state")
      --store string               the backing store to be searched for Kubeconfig files. Can be either "filesystem" or "vault" (default "filesystem")
      --vault-api-address string   the API address of the Vault store. Overrides the default "vaultAPIAddress" field in the SwitchConfig. This flag is overridden by the environment variable "VAULT_ADDR".

Use "switch [command] --help" for more information about a command.
```

Just type `switch` to search over the context names defined in the default Kubeconfig file `~/.kube/config`.

To recursively **search over multiple directories, files and Kubeconfig stores**, please see the [documentation](docs/kubeconfig_stores.md) 
to set up the necessary configuration file.

## Kubeconfig stores

Two Kubeconfig stores are supported: `filesystem` and `vault`.
The local filesystem is the default store and does not require any additional setup.
However, if you intend to search for all Kubeconfig context/files in the `~/.kube` directory, 
please [first consider this](docs/kubeconfig_stores.md#additional-considerations).

To search over multiple directories and setup Kubeconfig stores (such as Vault), [please see here](docs/kubeconfig_stores.md).

### Search Index

Use a search index to speed up searches.
Per default, the search index is **not** used.

Using the search index is especially useful when 
 - dealing with large amounts of Kubeconfigs and querying the Kubeconfig store is slow (e.g. searching a large directory)
 - when using vault as the Kubeconfig store to save requests against the Vault API 

Enable the search index in the `SwitchConfig` file (per default located in `~/.kube/switch-config.yaml` or configured via flag `--config-path`. The flag has to point to the file, not the directory).
The field `KubeconfigRediscoveryInterval` determines the time after which the tool should 
refresh its index against the configured Kubeconfig store.
The index is stored in a file in the state directory (default: `~/.kube/switch-state/switch.<store>.index`)

```
$ cat ~/.kube/switch-config.yaml

kind: SwitchConfig
KubeconfigRediscoveryInterval: 1h
```

Compared to the example in the [hot reload feature](#hot-reload), see that all the Kubeconfig contexts are available almost instantly.

![demo GIF](resources/pictures/index-demo.gif)

## Directory setup 

The `switch` tool just recursively searches through a specified path in a Kubeconfig store for Kubeconfig files matching a name.
The path layout presented below is purely optional.

When dealing with a large amount of Kubeconfig names, the `context` names are not necessarily unique or convey a meaning (especially when they are generated names).
To circumvent that issue, the fuzzy search includes the parent folder name.
This way, the directory layout can actually convey information useful for the search.

To exemplify this, look at the directory layout below. 
Each Kubernetes landscape (called `dev`, `canary` and `live`) have their own directory containing the Kubeconfigs of the Kubernetes clusters on that landscape.
Every `Kubeconfig` is named `config`.

```
$ tree .kube/switch
.kube/switch
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

![demo GIF](resources/pictures/search-show-parent-folder.png)

### Hot Reload

For large directories with many Kubeconfig files, the Kubeconfigs are added to the search set on the fly.
For smaller directory sizes, the search feels instantaneous.

 ![demo GIF](resources/pictures/hot-reload.gif)

### Extensibilty 

Customization is possible by using `Hooks` (think Git pre-commit hooks). 
Hooks can call an arbitrary executable or execute commands at a certain time (e.g every 6 hours) prior to the search via `switch`.
For more information [take a look here](./hooks/README.md). 

### How it works

The tool sets the `Kubeconfig` environment variable in the current shell session to a temporary copy of the selected `Kubeconfig` file. 
This way different Kubernetes clusters can be targeted in each terminal window.

There are two separate tools involved. The first one is `switch.sh`, a tiny bash script, and then there is the `switcher` binary.
The only thing the `switch` script does, is calling the `switcher` binary, capturing the path to the user selected `Kubeconfig` and then setting 
the `Kubeconfig` environment variable.
In order for the script to set the environment variable in the current shell session, it has to be sourced.

The `switcher`'s job is to displays a fuzzy search based on a recursive directory search for `Kubeconfig` files in the configured directory.
A temporary copy of the selected `Kubeconfig` file is created in `~/.kube/switch_tmp`.
To clean all created temporary files use `switch clean`.

### Difference to [kubectx.](https://github.com/ahmetb/kubectx)

While [kubectx.](https://github.com/ahmetb/kubectx) is designed to switch between contexts in a Kubeconfig file, 
this tool is best for dealing with many individual `Kubeconfig` files.

Another difference is, that multiple terminal windows targeting the same cluster do not interfere with each other.
Each terminal window can target a different cluster and namespace.

### Limitations

- `homebrew` places the `switch` script into `/usr/local/Cellar/switch/v0.0.3/switch.sh`. 
This is undesirable as the file location contains the version. Hence for each version you currently need to change your configuration.
- Make sure that within one directory, there are no identical `Kubeconfig` context names. Put them in separate folders. 
Within one `Kubeconfig` file, the context name is unique. So the easiest way is to just put each `Kubeconfig` file in 
its own directory with a meaningful name.
