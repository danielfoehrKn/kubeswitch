# Hooks

By providing a configuration file, the `kubeswitch` tool can call arbitrary hooks (an executable or inline shell command).

```
switch hooks help
Usage:
  --config-path path to the configuration file. (default "~/.kube/switch-config.yaml")
  --state-directory path to the state directory. (default "~/.kube/switch-state")
  --hook-name the name of the hook that should be run.
  --run-hooks-immediately run hooks right away. Do not respect the hooks execution configuration. (default "true").
```

Hooks are executed prior to the fuzzy search via `$ switch` or 
can be called directly via `$ switch hooks --hook-name=<name>`.

The default location for the config file is at `~/.kube/switch-config.yaml` or can be set with `--hook-config-path`.
 
You can find several demo configuration files [here](https://github.com/danielfoehrkn/kubeswitch/tree/master/resources/demo-config-files).

### See configured Hooks

This shows an overview of all configured hooks including their type, interval and when the next execution will be.
```
$ switch hooks ls

+---------------------------+------------+----------+----------------+
| NAME                      | TYPE       | INTERVAL | NEXT EXECUTION |
+---------------------------+------------+----------+----------------+
| sync-local-landscape      | Executable | 24h0m0s  | 3h48m0s        |
| sync-dev-landscape        | Executable | 6h0m0s   | 4h39m0s        |
| sync-ns2-canary-landscape | Executable | 12h0m0s  | 3h57m0s        |
| sync-live-landscape       | Executable | 24h0m0s  | 3h47m0s        |
| sync-ns2-live-landscape   | Executable | 48h0m0s  | 27h47m0s       |
| sync-canary-landscape     | Executable | 48h0m0s  | 27h47m0s       |
+---------------------------+------------+----------+----------------+
| TOTAL                     | 6          |          |                |
+---------------------------+------------+----------+----------------+
```
### Hook calling an executable

Hooks can call an arbitrary executable and pass arguments.

Below is the configuration to use with [Gardener](https://github.com/gardener/gardener) installations.
The Hook [`gardener-landscape-sync`](https://github.com/danielfoehrkn/kubeswitch/tree/master/hooks/gardener-landscape-sync) downloads the 
kubeconfig files for all available Kubernetes clusters to the local filesystem.
The hook is executed every 6 hours.
Not specifying an execution interval means that the hook shall only be run on demand via `switch hooks --hook-name <name>`.

```
kind: SwitchConfig
hooks:
  - name: sync-dev-landscape
    type: Executable
    path: /Users/<your-user>/go/src/github.com/danielfoehrkn/kubeswitch/hack/hooks/hook-gardener-landscape-sync
    arguments:
      - "sync"
      - "--garden-kubeconfig-path"
      - "/Users/<your-user>/.kube/switch/dev/dev-virtual-garden/config"
      - "--export-path"
      - "/Users/<your-user>/.kube/gardener-landscapes"
      - "--landscape-name"
      - "dev"
      - "--clean-directory=true"
    execution:
      interval: 6h
```

### Hook with inline command

A hook can also just execute shell commands directly. 
For example, the below configuration uses an inline command to garbage collection temporary kubeconfig files every 6 hours.

```
kind: SwitchConfig
hooks:
  - name: inline
    type: InlineCommand
    execution:
      interval: 6h
    arguments:
      - "/Users/<your-user>/go/src/github.com/danielfoehrkn/kubeswitch/hack/switch/switcher clean && echo ' Garbage collection complete.'"
```

### Hook State

To remember the last execution time for hooks, a file is written into the state directory.
The default location for the hook state files are at `~/.kube/switch-state` or can be set with `--state-directory`.
 
