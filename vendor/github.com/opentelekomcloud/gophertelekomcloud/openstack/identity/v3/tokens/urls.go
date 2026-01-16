package tokens

import "github.com/opentelekomcloud/gophertelekomcloud"

func tokenURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("auth", "tokens")
}
