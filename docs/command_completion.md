# Command completion 

**Note**: this is typically not needed, as when installing the shell function manually via [source the shell function](#source-the-shell-function), the completion script is already included.

Install the completion script by running:

### Bash

```sh
echo 'source <(switch completion bash)' >> ~/.bashrc
```
### Zsh
```sh
echo 'source <(switch completion zsh)' >> ~/.zshrc
```
### Fish
```sh
echo 'kubeswitch completion fish | source' >> ~/.config/fish/config.fish
```
