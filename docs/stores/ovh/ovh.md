# OVH store

To use the OVH store a token should be created on OVH's website. In order to create a token you should visit https://www.ovh.com/auth/api/createToken. You will get an `application key`, `application secret` and a `consumer key`.

In order to create this token you also need to specify the scope of the application. The required permissions for this plugin to work are the following:

- `GET /cloud/project`
- `GET /cloud/project/*/kube`
- `GET /cloud/project/*/kube/*`
- `POST /cloud/project/*/kube/*/kubeconfig`

Searching over multiple OVH instances is supported, but may require `showPrefix` to be set to `true` in the `SwitchConfig` file to avoid name collisions.

## Configuration

The OVH store configuration is defined in the `kubeswitch` configuration file. An example configuration is shown below:

```yaml
kind: SwitchConfig
version: v1alpha1
kubeconfigStores:
- kind: ovh
  config:
    application_key: <application key>
    application_secret: <application secret>
    consumer_key: <consumer_key>
  cache:
    kind: filesystem
    config:
      path: ~/.kube/cache
```

The OVH store can be used without a filesystem cache but the OVH API will create a new Kubeconfig file (and token) every time you switch to one of the OVH contexts.
Therefore, it is recommended to use a filesystem cache.
