# STACKIT Kubernetes Engine (SKE) store

The SKE store discovers and provides access to all [STACKIT Kubernetes Engine](https://www.stackit.de/en/product/kubernetes-engine/) clusters in a given STACKIT project.

## Prerequisites

An authentication method must be configured. The SKE store supports the following options (evaluated in order):

1. **STACKIT CLI** – Browser-based personal account login. Requires the [`stackit` CLI](https://github.com/stackitcloud/stackit-cli) to be installed and a one-time `stackit auth login`. Token refresh is handled automatically. Enable with `useStackitCLIAuth: true`.
2. **Service Account Token** – Set directly in the store config or via the `STACKIT_SERVICE_ACCOUNT_TOKEN` environment variable.
3. **Service Account Key File** – A JSON key file created via the [STACKIT portal](https://portal.stackit.cloud/).


## Configuration

The SKE store configuration is defined in the `kubeswitch` configuration file. An example is shown below:

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: ske
  # Optional: cache the cluster list for 1 hour; omit to always fetch from the API
  refreshIndexAfter: 1h
  config:
    # Required: the STACKIT project ID to search for SKE clusters
    projectID: "your-stackit-project-id"

    # Optional: human-readable project label used in context paths.
    # Context paths become: ske/<projectName>/<clusterName>
    # Falls back to projectID if not set.
    projectName: "my-project"

    # Optional: region (default: "eu01")
    region: "eu01"

    # --- Authentication (choose one) ---

    # Option 1: delegate to the STACKIT CLI (browser-based personal account login)
    # Requires: stackit auth login
    useStackitCLIAuth: true

    # Option 2: service account token
    # serviceAccountToken: "your-service-account-token"
    # (or set STACKIT_SERVICE_ACCOUNT_TOKEN env variable)

    # Option 3: path to a STACKIT service account key JSON file
    # serviceAccountKeyPath: "/path/to/sa-key.json"
```

### Personal account login (STACKIT CLI)

The easiest way to authenticate as a personal account is to install the STACKIT CLI and log in once:

```bash
stackit auth login   # opens browser for OAuth login
```

Then set `useStackitCLIAuth: true` in your store config. The CLI manages token refresh transparently — no manual credential management is needed.



### Multiple projects

To discover clusters across multiple STACKIT projects, configure one store per project:

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: ske
  id: project-a
  refreshIndexAfter: 1h
  config:
    projectID: "project-a-id"
    projectName: "Project A"
    useStackitCLIAuth: true
- kind: ske
  id: project-b
  refreshIndexAfter: 1h
  config:
    projectID: "project-b-id"
    projectName: "Project B"
    useStackitCLIAuth: true
```

Context paths follow the pattern `ske/<projectName>/<clusterName>` (or `<storeID>/<projectName>/<clusterName>` when an explicit store `id` is set).

### Caching

Setting `refreshIndexAfter` causes kubeswitch to write a local index file (e.g. `~/.kube/switch-state/switch.ske.default.index`) after the first search. Subsequent runs read from this index instead of hitting the STACKIT API, until the duration expires.
