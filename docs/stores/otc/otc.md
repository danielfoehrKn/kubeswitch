# OpenTelekom Cloud CCE store

Kubeswitch can discover CCE clusters from OTC.

Kubeswitch relies on the [Openstask configuration files](https://docs.openstack.org/python-openstackclient/latest/configuration/index.html), with additions as defined by the underlying [SDK](https://github.com/opentelekomcloud/gophertelekomcloud)

## Clouds configuration

First create a file `~/.config/openstack/clouds.yaml`

```
cat ~/.config/openstack/clouds.yaml

clouds:
  my-cloud-with-permanent-aksk:
    auth:
      auth_url: https://iam.eu-de.otc.t-systems.com:443/v3
      project_name: eu-de_xxx     # also known as tenant name
      ak: YOUR_AK
      sk: YOUR_SK
    interface: public
    identity_api_version: "3"
    auth_type: aksk
  my-cloud-with-temporary-aksk:
    auth:
      auth_url: https://iam.eu-de.otc.t-systems.com/v3
      project_name: eu-de_yyy    # also known as tenant name
      ak: YOUR_TEMPORARY_AK
      sk: YOUR_TEMPORARY_SK
      security_token: YOUR_TEMPORARY_STS
    interface: public
    identity_api_version: "3"
    auth_type: aksk

```

Permanent and temporary AK/SK as well as temporary STS can be generated in OTC web console after you log in.

Other authentication methods supported by the format and SDK (password, federated, token, assume role, ...) may also work, but were not tested.


## Kubectx configuration

```
cat ~/.kube/switch-config.yaml

kind: SwitchConfig
version: "v1alpha1"
kubeconfigStores:
  - kind: otc
    id: otc1
    config:
      cloud: my-cloud-with-permanent-aksk
  
  - kind: otc
    id: otc2
    config:
      cloud: my-cloud-with-temporary-aksk
      selectContext: externalTLSVerify
```

Retrieved kubeconfig for each CCE cluster contains 3 contexts (same ones as when it is downloaded from the web console):
- internal
- external
- externalTLSVerify

By default, all three are available for selection by `kubeswitch`. 

The optinal parameter `selectContext` allows for selecting only one of these contexts for each cluster.