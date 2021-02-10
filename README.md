# Kubeswitch

The `kubectx` for operators.

Kubeswitch (short: `switch`) is a tiny standalone tool, 
designed to conveniently switch between the context of thousands of `kubeconfig` files.

## Highlights

- **Efficient search**
  - stores a pre-computed index of kube-context names to speed up future searches
  - recursive directory search
  - hot reload capability (adds kubeconfigs to the search on the fly - especially useful when initially searching large directories)
- **Configurable Kubeconfig store**
  - Local filesystem
  - Hashicorp Vault 
- **Improved search experience** when dealing with many kubeconfigs
  - Live preview of the kubeconfig file (**sanitized from credentials**).
  - Kubeconfig context names are easily identifiable. The `context` is prefixed with the (immediate) parent folder name to allow to easily find the context you are looking for. 
  - Same fuzzy search capability known from `kubectx`
- **Terminal Window isolation**
  - Each terminal window can target a different cluster (does not override the current-context in a shared kubeconfig).
  - Each terminal window can target the same cluster and set a [different namespace preference](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/#setting-the-namespace-preference) 
  e.g via the tool [kubens](https://github.com/ahmetb/kubectx).
- **Extensibility** 
  - Integrate custom functionality using [Hooks](./hooks/README.md) (comparable with Git pre-commit hooks).

![demo GIF](resources/switch-demo-large.gif)

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
$ source $HOMEBREW_CELLAR/switch/v0.1.0/switch.sh
```

Updating the version of the `switch` utility via `brew` (e.g changing from version 0.1.0 to 0.1.1) requires to change the sourced path.

### Option 2 - Manual Installation

#### Mac

Download the switch script and the switcher binary for your OS/architecture (darwin / linux).
```
# grab pre-compiled switcher binary for your OS/architecture
OS=linux #darwin
wget https://github.com/danielfoehrKn/kubeswitch/releases/download/0.1.0/switcher_${OS}_amd64.tar.gz
tar -zxvf switcher_${OS}_amd64.tar.gz
cp switcher_${OS}_amd64 /usr/local/bin/switcher
rm switcher_${OS}_amd64.tar.gz

# grab switch script
wget https://github.com/danielfoehrKn/kubeswitch/releases/download/0.1.0/switch.tar.gz
tar -zxvf switch.tar.gz
cp switch.sh /usr/local/bin/switch.sh
rm switch.tar.gz
```

Source `switch.sh` e.g. in the `.bashrc`/`.zsh`) via:
```
$ source /usr/local/bin/switch.sh
```
#### Linux

If you are running Linux, you can download the `switcher` binary from the [releases](https://github.com/danielfoehrKn/kubeswitch/releases)
, put it in your path, and then source the `switch` script from [here](https://github.com/danielfoehrKn/kubeswitch/blob/master/hack/switch/switch.sh).

## Usage 

```
$ switch -h

Simple tool for switching between kubeconfig contexts. The kubectx build for people with a lot of kubeconfigs.

Usage:
  switch [flags]
  switch [command]

Available Commands:
  clean       Cleans all temporary kubeconfig files
  help        Help about any command
  hooks       Runs configured hooks

Flags:
      --config-path string    path to the configuration file. (default "~/.kube/switch-config.yaml")
  -h, --help                       help for switch
      --kubeconfig-name string     only shows kubeconfig files with this name. Accepts wilcard arguments "*" and "?". Defaults to "config". (default "config")
      --kubeconfig-path string     path to be recursively searched for kubeconfig files. Can be a directory on the local filesystem or a path in Vault. (default "~/.kube")
      --show-preview               show preview of the selected kubeconfig. Possibly makes sense to disable when using vault as the kubeconfig store to prevent excessive requests against the API. (default true)
      --state-directory string     path to the local directory used for storing internal state. (default "~/.kube/switch-state")
      --store string               the backing store to be searched for kubeconfig files. Can be either "filesystem" or "vault" (default "filesystem")
      --vault-api-address string   the API address of the Vault store.

Use "switch [command] --help" for more information about a command.
```
Just type `switch` to search for kubeconfig files with name `config` in the local `~/.kube` directory. 
As shown above, there are multiple flags available to adjust this behaviour.

## Configuration

I recommend setting up an alias for your `switch` commands.

For instance, to always using the directory `~/.kube/switch` to search for kubeconfig files,
use the following alias:

```
alias switch='switch --kubeconfig-path ~/.kube/switch'
```

### Kubeconfig stores

There are two Kubeconfig stores support: `filesystem` and `vault`.

The local filesystem is the default store and does not require any additional settings.
However, to speed up the fuzzy search on the local filesystem, 
I would recommend putting all the kubeconfig files into a single directory containing only kubeconfig files.
Then create a `switch` alias via `--kubeconfig-path` pointing to this directory.
This is because the default `~/.kube` directory contains a bunch of other files 
that have to be filtered out and thus slowing down the search.

To use vault as the kubeconfig store, [please see here](docs/use_vault_store.md).

### Search Index

Use a search index to speed up searches.
Per default, the search index is **not** used.

Using the search index is especially useful when 
 - dealing with large amounts of kubeconfigs and querying the kubeconfig store is slow (e.g. searching a large directory)
 - when using vault as the kubeconfig store to save requests against the Vault API 

Enable the search index in the `SwitchConfig` file (per default located in `~/.kube/switch-config.yaml` or configured via flag `--config-path`. The flag has to point to the file, not the directory).
The field `kubeconfigRediscoveryInterval` determines the time after which the tool should 
refresh its index against the configured kubeconfig store.
The index is stored in a file in the state directory (default: `~/.kube/switch-state/switch.<store>.index`)

```
$ cat ~/.kube/switch-config.yaml

kind: SwitchConfig
kubeconfigRediscoveryInterval: 1h
```

Compared to the example in the [hot reload feature](#hot-reload), see that all the kubeconfig contexts are available almost instantly.

![demo GIF](resources/index-demo.gif)

## Directory setup 

The `switch` tool just recursively searches through a specified path in a kubeconfig store for kubeconfig files matching a name.
The path layout presented below is purely optional.

When dealing with a large amount of kubeconfig names, the `context` names are not necessarily unique or convey a meaning (especially when they are generated names).
To circumvent that issue, the fuzzy search includes the parent folder name.
This way, the directory layout can actually convey information useful for the search.

To exemplify this, look at the directory layout below. 
Each Kubernetes landscape (called `dev`, `canary` and `live`) have their own directory containing the kubeconfigs of the Kubernetes clusters on that landscape.
Every `kubeconfig` is named `config`.

```
$ tree .kube/switch
.kube/switch
├── canary
│   └── config
├── dev
│   ├── config
│   └── config-tmp
└── live
    └── confi
```
This is how the search looks like for this directory.
The parent directory name is part of the search.

![demo GIF](resources/search-show-parent-folder.png)

### Hot Reload

For large directories with many kubeconfig files, the kubeconfigs are added to the search set on the fly.
For smaller directory sizes, the search feels instantaneous.

 ![demo GIF](resources/hot-reload.gif)

### Extensibilty 

Customization is possible by using `Hooks` (think Git pre-commit hooks). 
Hooks can call an arbitrary executable or execute commands at a certain time (e.g every 6 hours) prior to the search via `switch`.
For more information [take a look here](./hooks/README.md). 

### How it works

The tool sets the `KUBECONFIG` environment variable in the current shell session to a temporary copy of the selected `kubeconfig` file. 
This way different Kubernetes clusters can be targeted in each terminal window.

There are two separate tools involved. The first one is `switch.sh`, a tiny bash script, and then there is the `switcher` binary.
The only thing the `switch` script does, is calling the `switcher` binary, capturing the path to the user selected `kubeconfig` and then setting 
the `KUBECONFIG` environment variable.
In order for the script to set the environment variable in the current shell session, it has to be sourced.

The `switcher`'s job is to displays a fuzzy search based on a recursive directory search for `kubeconfig` files in the configured directory.
A temporary copy of the selected `kubeconfig` file is created in `~/.kube/switch_tmp`.
To clean all created temporary files use `switch clean`.

### Difference to [kubectx.](https://github.com/ahmetb/kubectx)

While [kubectx.](https://github.com/ahmetb/kubectx) is designed to switch between contexts in a kubeconfig file, 
this tool is best for dealing with many individual `kubeconfig` files.

Another difference is, that multiple terminal windows targeting the same cluster do not interfere with each other.
Each terminal window can target a different cluster and namespace.

### Limitations

- `homebrew` places the `switch` script into `/usr/local/Cellar/switch/v0.0.3/switch.sh`. 
This is undesirable as the file location contains the version. Hence for each version you currently need to change your configuration.
- Make sure that within one directory, there are no identical `kubeconfig` context names. Put them in separate folders. 
Within one `kubeconfig` file, the context name is unique. So the easiest way is to just put each `kubeconfig` file in 
its own directory with a meaningful name.
