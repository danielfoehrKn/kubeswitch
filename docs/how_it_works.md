# How it works

`Kubeswitch` consists of two components:

`switch.sh` - contains the shell function `switch()` which is the entry point to Kubeswitch.
`switcher`  - a go binary which is executed by the switch shell function and handles context selection and manipulation of the selected Kubeconfig file.

For a proper installation
 - you have to source the `switch.sh` script
 - make the switcher binary available in your $PATH (the switch.sh shell script looks for the switcher binary on the $PATH).

**Flow**

1) User executes `$ switch` (bash function available via the sourced **switch.sh** script) 
2) The `switch.sh` script discovers and executes the switcher binary from the $PATH
3) The `switcher` binary searches for kubeconfigs in the [kubeconfig stores](kubeconfig_stores.md) configured in the `SwitchConfig` configuration file.
4) The `switcher` binary displays a fuzzy search for kubeconfig context names
5) The user selects on context name
6) The `switcher` binary creates a temporary copy of the selected kubeconfig file, sets the `current-context` and writes the kubeconfig to `~/.kube/switch_tmp`
7) The `switcher` binary writes the filepath to the kubeconfig to STDOUT
8) The `switch.sh` script captures this filepath and executes `export KUBECONFIG=</path/to/tmp/kubeconfig/file>` 

Each terminal window operates on its own copy of the kubeconfig file (terminal window isolation).
