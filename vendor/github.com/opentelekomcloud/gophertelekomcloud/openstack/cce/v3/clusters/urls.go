package clusters

import "github.com/opentelekomcloud/gophertelekomcloud"

const (
	rootPath     = "clusters"
	certPath     = "clustercert"
	masterIpPath = "mastereip"
)

func rootURL(client *golangsdk.ServiceClient) string {
	return client.ServiceURL(rootPath)
}

func resourceURL(c *golangsdk.ServiceClient, id string) string {
	return c.ServiceURL(rootPath, id)
}

func certificateURL(c *golangsdk.ServiceClient, id string) string {
	return c.ServiceURL(rootPath, id, certPath)
}

func masterIpURL(c *golangsdk.ServiceClient, id string) string {
	return c.ServiceURL(rootPath, id, masterIpPath)
}
