# Command completion 

Currently, command line completion is not pre-installed in any installation method.
You need to do it manually.

## Bash

Source this [this script ](../scripts/_switch.bash) from your `~/.bashrc`
or put it into [your completions directory](https://serverfault.com/questions/506612/standard-place-for-user-defined-bash-completion-d-scripts).

## Zsh

There is currently only a bash completion script.
But you can use it in zsh also.

Add below lines to your `~/.zshrc` file (before you source the bash completion script).

```
autoload bashcompinit
bashcompinit
```

Then source the bash completion script.