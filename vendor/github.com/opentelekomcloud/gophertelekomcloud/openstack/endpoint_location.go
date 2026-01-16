package openstack

import (
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	tokens3 "github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/tokens"
)

/*
V3EndpointURL discovers the endpoint URL for a specific service from a Catalog
acquired during the v3 identity service.

The specified EndpointOpts are used to identify a unique, unambiguous endpoint
to return. It's an error both when multiple endpoints match the provided
criteria and when none do. The minimum that can be specified is a Type, but you
will also often need to specify a Name and/or a Region depending on what's
available on your OpenStack deployment.
*/
func V3EndpointURL(catalog *tokens3.ServiceCatalog, opts golangsdk.EndpointOpts) (string, error) {
	// Extract Endpoints from the catalog entries that match the requested Type, Interface,
	// Name if provided, and Region if provided.
	var endpoints = make([]tokens3.Endpoint, 0, 1)
	for _, entry := range catalog.Entries {
		if (entry.Type == opts.Type) && (opts.Name == "" || entry.Name == opts.Name) {
			for _, endpoint := range entry.Endpoints {
				if opts.Availability != golangsdk.AvailabilityAdmin &&
					opts.Availability != golangsdk.AvailabilityPublic &&
					opts.Availability != golangsdk.AvailabilityInternal {
					err := &ErrInvalidAvailabilityProvided{}
					err.Argument = "Availability"
					err.Value = opts.Availability
					return "", err
				}
				if opts.Availability == golangsdk.Availability(endpoint.Interface) &&
					(opts.Region == "" || endpoint.Region == opts.Region) {
					endpoints = append(endpoints, endpoint)
				}
			}
		}
	}

	// If multiple endpoints were found, use the first result
	// and disregard the other endpoints.
	//
	// This behavior matches the Python library. See GH-1764.
	if len(endpoints) > 1 {
		endpoints = endpoints[0:1]
	}

	// Extract the URL from the matching Endpoint.
	for _, endpoint := range endpoints {
		return golangsdk.NormalizeURL(endpoint.URL), nil
	}

	// Report an error if there were no matching endpoints.
	err := &golangsdk.ErrEndpointNotFound{}
	return "", err
}
