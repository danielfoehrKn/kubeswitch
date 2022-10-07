# Installation

## `switch` command

### Option 1a - Homebrew

Mac and Linux users can install both the `switch.sh` script and the `switcher` binary with `homebrew`.
```
$ brew install danielfoehrkn/switch/switch
```

Source the `switch.sh` script from the `homebrew` installation path by adding this to `.bashrc`/`.zsh`:
```
INSTALLATION_PATH=$(brew --prefix switch) && source $INSTALLATION_PATH/switch.sh
```

### Option 1b - MacPorts

Mac users can also install both `switch.sh` and `switcher` from [MacPorts](https://www.macports.org)
```
$ sudo port selfupdate
$ sudo port install kubeswitch
```

Source the `switch.sh` script from the MacPorts root (/opt/local).
```
$ source /opt/local/libexec/kubeswitch/switch.sh
```

### Option 2 - Manual Installation

#### From source

```
$ go get github.com/danielfoehrkn/kubeswitch
```

From the repository root run `make build-switcher`.
This builds the binaries to `/hack/switch/`.
Copy the build binary for your OS / Architecture to e.g. `/usr/local/bin`
and source the switch script from `/hack/switch/switch.sh`.

#### From Github releases

Download the switch script and the switcher binary.

```sh
OS=linux                        # Pick the right os: linux, darwin (intel only)
VERSION=0.7.0                   # Pick the current version.

curl -L -o /usr/local/bin/switcher https://github.com/danielfoehrKn/kubeswitch/releases/download/${VERSION}/switcher_${OS}_amd64
chmod +x /usr/local/bin/switcher

curl -L -o  /usr/local/bin/switch.sh https://github.com/danielfoehrKn/kubeswitch/releases/download/${VERSION}/switch.sh
chmod +x /usr/local/bin/switch.sh
```

Source `switch.sh` in `.bashrc`/`.zsh` via:

```sh
source /usr/local/bin/switch.sh
```

## Set up finding kubeconfig files and contexts

If you installed kubeswitch correctly, you can run the command `switch`
and should see the contexts it can find with its default configuration.
The command is only available once you open a new terminal, in case you load
`switch.sh` through `.bashrc`/`.zsh`. If you get the error
`Error: you need to point kubeswitch to a kubeconfig file` or do not see all
desired kubeconfig contexts that you want to choose from, follow
[kubeconfig stores](kubeconfig_stores.md) for the configuration.

## Command completion

Please [see here](command_completion.md) how to install command completion for bash and zsh shells.
This completes both the `kubeswitch` commands as well as the context names.

## Clean up temporary kubeconfig files

To not alter the current shell session, `kubeswitch` does not spawn a new sub-shell.
You need to configure a cleanup handler if you care to remove temporary kubeconfig files from `$HOME/.kube/.switch_tmp` when the shell session
ends (close the terminal window, or `exit` is called).
For `zsh`, please source [this script](/scripts/cleanup_handler_zsh.sh) from your `.zshrc` file.
