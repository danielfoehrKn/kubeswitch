# Command completion 

Currently, command line completion is not pre-installed in any installation method.
You need to do it manually.

Install the completion script by running:

### Bash

```sh
echo 'source <(switch completion bash)' >> ~/.bashrc
```
### Zsh
```sh
echo 'source <(switch completion zsh)' >> ~/.bashrc
```
### Fish
```sh
echo 'kubeswitch completion fish | source' >> ~/.config/fish/config.fish
```