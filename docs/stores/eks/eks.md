# EKS store

Kubeswitch can discover EKS clusters from AWS.

At the moment, Kubeswitch works only with [named profiles](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html), make sure it is working before setting up the store.
A good test is to follow the "(Create kubeconfig automatically)[https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html#create-kubeconfig-automatically]" documentation as Kubeswitch will output the same configuration (with some exceptions pending implementation).

## Setup

Please make sure the `aws` cli is intalled and on your `PATH` in a version providing the [`aws eks get-token`](https://docs.aws.amazon.com/cli/latest/reference/eks/get-token.html) command (requires version 1.16.156 or later of the AWS CLI).
This is because the generated kubeconfigs use the AWS CLI as a credential helper.

Next, create the EKS store configuration in the `kubeswitch` configuration file.

Search over all EKS clusters for a config profile and region:
```
cat ~/.kube/switch-config.yaml

kind: SwitchConfig
version: "v1alpha1"
kubeconfigStores:
  - kind: eks
    id: <my-id-only-required-if-there-are-multiple-eks-stores>
    config:
      profile: user1
      region: us-east-1
  - kind: eks
    id: <my-id-only-required-if-there-are-multiple-eks-stores>
    config:
      profile: user1
      # no region, will lookup AWS_DEFAULT_REGION
  - kind: eks
    id: <my-id-only-required-if-there-are-multiple-eks-stores>
    # no config, will lookup AWS_PROFILE and AWS_REGION (or AWS_DEFAULT_REGION)
```

## Multiple profiles

Using multiple profiles and/or regions is possible by defining multiple store configurations in the `switch-config` file (one for each profile and/or region).

## Search for EKS Clusters

Kubeconfig context names are fuzzy-searchable using the following semantics. 
The `eks_` prefix helps to narrow down the search to only EKS clusters.

In General:
- `eks_<profile>--<region>--<cluster-name>/<cluster-name>`

Example:
- `eks_prod1--eu-west-1--kubeswitch_test/kubeswitch_test`

In this example:
- Configuration Profile: prod1
- Region: eu-west-1
- EKS Cluster name: kubeswitch_test

However, remember that you can always define an `alias` for each context to define a name that you can better remember or query .

This is how looks like using the `switch` search:
- In addition to the sanitized kubeconfig preview, additional EKS cluster information is shown such as the `Kubernetes version`
