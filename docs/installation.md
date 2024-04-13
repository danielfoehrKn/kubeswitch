# Installation

The kubeswitch installation consists of both a `switcher` binary and a shell script which needs to be sourced.

**NOTE**: to invoke kubeswitch, do not call the `switcher` binary directly from the command line. 
Instead, use the sourced shell function as described in [source the shell function](#required-source-the-shell-function).

## Option 1 - Homebrew
**NOTE**: `fish` users please follow [install via Github releases](#option-2---github-releases) as the shell script only works for `zsh` and `bash` shells.

Install the `switcher` binary with `homebrew`.
```
$ brew install danielfoehrkn/switch/switch
```

Next, follow [required: source the shell function](#required-source-the-shell-function).

### Option 2 - MacPorts
**NOTE**: `fish` users please follow [install via Github releases](#option-2---github-releases) as the shell script only works for `zsh` and `bash` shells.

Mac users can also install both `switch.sh` and `switcher` from [MacPorts](https://www.macports.org)
```
$ sudo port selfupdate
$ sudo port install kubeswitch
```

Next, follow [required: source the shell function](#required-source-the-shell-function).

### Option 2 - Github releases

Download the switcher binary
```sh
OS=linux                        # Pick the right os: linux, darwin (intel only)
VERSION=0.8.0                   # Pick the current version.

curl -L -o /usr/local/bin/switcher https://github.com/danielfoehrKn/kubeswitch/releases/download/${VERSION}/switcher_${OS}_amd64
chmod +x /usr/local/bin/switcher
```

Next, follow [required: source the shell function](#required-source-the-shell-function).

### Option 3 - From source

```
$ go get github.com/danielfoehrkn/kubeswitch
```

From the repository root run `make build-switcher`.
This builds the binaries to `/hack/switch/`.
Copy the build binary for your OS/Architecture to e.g. `/usr/local/bin`.

Next, follow [required: source the shell function](#required-source-the-shell-function).

## Required: Source the shell function

Source the shell function which is used to call the `switcher` binary. 
For `zsh/bash` the name of the shell function is `switch` and for `fish` its `kubeswitch`.
Additionally, installs the command completion script.

### Bash

```sh
echo 'source <(switcher init bash)' >> ~/.bashrc

# optionally use alias `s` instead of `switch`
echo 'alias s=switch' >> ~/.bashrc
echo 'complete -o default -F _switcher s' >> ~/.bashrc
```
### Zsh
```sh
echo 'source <(switcher init zsh)' >> ~/.zshrc

# optionally use alias `s` instead of `switch`
echo 'source <(alias s=switch)' >> ~/.zshrc

# optionally use command completion
echo 'source <(switch completion zsh)' >> ~/.zshrc
```
### Fish
Fish shell have a built-in `switch` function. Hence, differently from `zsh` shells, the kubeswitch function is called `kubeswitch`.
```sh
echo 'switcher init fish | source' >> ~/.config/fish/config.fish

# optionally use alias `s` instead of `kubeswitch` (add to config.fish)
function s --wraps switcher
        kubeswitch $argv;
end
```

## Check that it works

If you installed kubeswitch correctly, you can run the command `switch` (zsh, bash) or `kubeswitch` (fish) or alternatively the alias `s` from the terminal.
In case the terminal can't find the function, you might need to open another terminal or re-source your config file (`.zshrc`,`.bashrc`,...).

That should display the contexts the tool can find with the default configuration.
If you get the error `Error: you need to point kubeswitch to a kubeconfig file` or do not see all
desired kubeconfig contexts that you want to choose from, follow
[kubeconfig stores](kubeconfig_stores.md) for the configuration.
