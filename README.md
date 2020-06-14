# Kubectl switch

`switch` is a tiny standalone tool, designed to conveniently switch between hundreds of `kubeconfig` files without having to remember `kubeconfig` context names.

## Features

- Fuzzy search for `kubeconfig` files in a configurable location on your local filesystem.
- Isolation between terminal sessions
  - Each terminal window can target a different cluster
  - Each terminal window can target the same cluster and set a [different namespace preference](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/#setting-the-namespace-preference) 
  e.g via the tool [kubens](https://github.com/ahmetb/kubectx) 
- `Kubeconfig` is prefixed with the (immediate) parent folder name. Context names in `kubeconfig` files do not always identify the cluster in a way that you can easily find it - especially when they are generated. 
- Live preview (does not include credentials).

![demo GIF](resources/switch-demo.gif)

## Installation

Mac users can just use `homebrew` for installation.

If you are running Linux, you would need to download the `switcher` binary yourself, put it in your path, and then source the `switch` script.
#### Option 1

Install the `switcher` binary.
```
 $ brew install danielfoehrkn/switch/switcher
```

Grab the `switch` bash script [from here](https://github.com/danielfoehrKn/kubectlSwitch/blob/master/hack/switch/switch.sh), place it somewhere on your local filesystem and **source** it.
Where you source the script depends on your terminal (e.g .bashrc or .zsrhc).

`
$ source <my-path>/switch.sh
`

#### Option 2

Install both the `switcher` tool and the `switch` script with `homebrew`. 
```
 $ brew install danielfoehrkn/switch/switch
```

Source the `switch` script from the `homebrew` installation path.

```
$ source /usr/local/Cellar/switch/v0.0.1/switch.sh
```

Updating the version of the `switch` utility via `brew` (e.g changing from version 0.0.1 to 0.0.2) requires you to change the sourced path. 

## Usage 

```
$ switch -h
Usage:
  -kubeconfig-directory directory containing the kubeconfig files. Default is ~/.kube
  -kubeconfig-name only shows kubeconfig files with exactly this name. Defaults to 'config'.
  -executable-path path to the 'switch' executable. If unset tries to use 'switch' from the path.
  -help shows available flags.
  clean removes all the temporary kubeconfig files created in the directory ~/.kube/switch_tmp.
```

This is part of the directory tree of how I order my `kubeconfig` files. 
You can see that they are ordered hierarchically. 
Each landscape (dev, canary and live) in its own directory containing one directory per `kubeconfig`.
Every `kubeconfig` is named `config`.

```
$ tree .kube/switch
├── canary
│   ├── canary-seed-aws-eu-1
│   │   └── config
│   ├── canary-seed-aws-eu-2
│   │   └── config
│   ├── canary-seed-az-eu-3
│   │   └── config
│   ├── canary-virtual-garden
│   │   └── config
│   └── ns2
│       ├── ns2-canary-garden
│       │   └── config
│       ├── ns2-canary-seed-aws
│       │   └── config
│       └── ns2-canary-virtual-garden
│           └── config
├── dev
│   ├── dev-seed-alicloud
│   │   └── config
│   ├── dev-seed-aws
│   │   └── config
│   ├── dev-seed-az
│   │   └── config
├── live
│   ├── live-garden
│   │   └── config
│   ├── live-seed-aws-eu1
│   │   └── config
│   ├── live-seed-aws-eu2
│   │   └── config
```

Using the `switch` utility allows me to easily find the `kubeconfig` I am looking for.
Because the directory name are part of the search result, the target `kubeconfig` can be identified without having to remember context names.
In addition, the selection can be verified by looking at the live preview.
Please take a look at the gif above how that looks like.

```
# switch
```

If you think that could be helpful in managing you `kubeconfig` files, try it out and let me know what you think.

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

- `homebrew` places the `switch` script into `/usr/local/Cellar/switch/v0.0.1/switch.sh`. 
This is undesirable as the file location contains the version. Hence for each version you currently need to change your configuration.
- Make sure that within one directory, there are no identical `kubeconfig` context names. Put them in separate folders. 
Within one `kubeconfig` file, the context name is unique. So the easiest way is to just put each `kubeconfig` file in 
its own directory with a meaningful name.