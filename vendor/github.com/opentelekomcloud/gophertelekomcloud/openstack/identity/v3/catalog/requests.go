package catalog

import (
	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/pagination"
)

// List enumerates the services available to a specific user.
func List(client *golangsdk.ServiceClient) pagination.Pager {
	url := listURL(client)
	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return CatalogPage{pagination.LinkedPageBase{PageResult: r}}
	})
}
