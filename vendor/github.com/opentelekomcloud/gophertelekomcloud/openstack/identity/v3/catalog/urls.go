package catalog

import "github.com/opentelekomcloud/gophertelekomcloud"

func listURL(client *golangsdk.ServiceClient) string {
	return client.ServiceURL("auth", "catalog")
}
