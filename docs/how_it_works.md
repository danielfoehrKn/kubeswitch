# How it works

`Kubeswitch` sets the Kubeconfig environment variable in the current shell session to a temporary copy of the selected `Kubeconfig` file.
This way different Kubernetes clusters can be targeted in each terminal window.

There are two separate tools involved. The first one is `switch.sh`, a tiny bash script, and then there is the `switcher` binary.
The only thing the `switch` script does, is calling the `switcher` binary, capturing the path to the user selected `Kubeconfig` and then setting
the `Kubeconfig` environment variable.
In order for the script to set the environment variable in the current shell session, it has to be sourced.

The `switcher`'s job is to display a fuzzy search based on a recursive directory search for `Kubeconfig` files in the configured directory.
A temporary copy of the selected `Kubeconfig` file is created in `~/.kube/switch_tmp`.
To clean all created temporary files use `switch clean`.