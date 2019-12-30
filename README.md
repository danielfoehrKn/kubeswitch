## Kubectl Switch

Enables switching between kubeconfigs with fuzzy search similar to [kubectx.](https://github.com/ahmetb/kubectx)

**Features**

- Display fzf style fuzzy search dialog 
- Show live preview of selected kubeconfig
- One target cluster per terminal window/tab possible ([kubectx](https://github.com/ahmetb/kubectx) only supports a single context globally)

**How it works**

I currently use this tool together with a bashscript that searches a set directory and hands over all found kubeconfigs to this tool.
The _switch tool_ then displays a selection dialog and sets the current-context in the selected kubeconfig file.
The path to the selected kubeconfig is written to stdout to be read by the bashscript.
As a result the KUBECONFIG environment variable is set.

You can find the bashscript [here.](https://github.com/danielfoehrKn/bashscripts/blob/master/functions/switch.sh#L4)

**Limitations/ Planned**

- Getting rid of the bashscript to have a standalone tooling.
Requires implementing the _find utility_ with golang.
- Configurable search directory (currently needs to be configured in bashscript)