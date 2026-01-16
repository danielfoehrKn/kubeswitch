# GopherTelekomCloud: a OpenTelekomCloud SDK for Golang

[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/gophertelekomcloud?branch=devel)](https://goreportcard.com/report/github.com/opentelekomcloud/gophertelekomcloud)
[![Zuul Gated](https://zuul-ci.org/gated.svg)](https://zuul.eco.tsi-dev.otc-service.com/t/eco/buildsets?project=opentelekomcloud%2Fgophertelekomcloud&pipeline=gate)
[![LICENSE](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/opentelekomcloud/gophertelekomcloud/blob/master/LICENSE)

GopherTelekomCloud is a OpenTelekomCloud clouds Go SDK. GopherTelekomCloud is based
on [Gophercloud](https://github.com/gophercloud/gophercloud)
which is an OpenStack Go SDK and has a great design. GopherTelekomCloud has added and removed some features to support
OpenTelekomCloud.

## Useful links

* [Reference documentation](http://godoc.org/github.com/opentelekomcloud/gophertelekomcloud)
* [Effective Go](https://golang.org/doc/effective_go.html)

## How to install

Installation with modern Go and `go mod` is really simple:

Just run `go mod download` to install all dependencies.

## Getting started

### Credentials

Because you'll be hitting an API, you will need to retrieve your OpenTelekomCloud credentials and store them using
standard Openstack approaches:
either [`clouds.yaml`](https://docs.openstack.org/python-openstackclient/latest/configuration/index.html)
file (recommended) or environment variables.

You will need to retrieve the following:

* domain name
* username
* password
* project name/id (for most of the services)
* a valid IAM identity URL

### Authentication

Once you have access to your credentials, you can begin plugging them into Golangsdk. The next step is authentication,
and this is handled by a base
"Provider" struct. To get one, you can either pass in your credentials explicitly, or tell Golangsdk to use environment
variables:

#### Option 1: Pass in the values yourself

```go
opts := golangsdk.AuthOptions{
IdentityEndpoint: "https://openstack.example.com:5000/v2.0",
Username:         "{username}",
Password:         "{password}",
}
client, err := openstack.AuthenticatedClient(opts)
```

#### Option 2: Use a utility function to retrieve cloud configuration from env variables and configuration files

```go
env := openstack.NewEnv("OS_") // use OS_ prefixed env variables
client, err := env.AuthenticatedClient()
```

The `ProviderClient` is the top-level client that all of your OpenTelekomCloud services derive from. The provider
contains all of the authentication details that allow your Go code to access the API - such as the base URL and token
ID.

### Provision a rds instance

Once we have a base Provider, we inject it as a dependency into each OpenTelekomCloud service. In order to work with the
rds API, we need a rds service client; which can be created like so:

```go
client, err := openstack.NewRdsServiceV1(provider, golangsdk.EndpointOpts{
	Region: utils.GetRegion(ao),
})
```

We then use this `client` for any rds API operation we want. In our case, we want to provision a rds instance - so we
invoke the `Create` method and pass in the name and the flavor ID (database specification) we're interested in:

```go
import "github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v1/instances"

instance, err := instances.Create(client, instances.CreateOpts{
	Name:      "My new rds instance!",
	FlavorRef: "flavor_id",
}).Extract()
```

The above code sample creates a new rds instance with the parameters, and embodies the new resource in the `instance`
variable (a[`instances.Instance`](http://godoc.org/github.com/opentelekomcloud/gophertelekomcloud) struct).

## Advanced Usage

Have a look at the [FAQ](./FAQ.md) for some tips on customizing the way Golangsdk works.

## Backwards-Compatibility Guarantees

None. Vendor it and write tests covering the parts you use.

## Contributing

See the [contributing guide](./.github/CONTRIBUTING.md).

## Help and feedback

If you're struggling with something or have spotted a potential bug, feel free to submit an issue to
our [bug tracker](https://github.com/opentelekomcloud/gophertelekomcloud/issues).
