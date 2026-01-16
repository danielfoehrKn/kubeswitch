package openstack

import (
	"regexp"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
)

// A regular expression used to verify whether or not contains a project id in an endpoint url
var endpointProjectIdMatcher = regexp.MustCompile(`http[s]?://.+/(?:[V|v]\d+|[V|v]\d+\.\d+)/([a-z|A-Z|0-9]{32})(?:/|$)`)

// ContainsProjectId detects whether or not the endpoint url contains a projectID
func ContainsProjectId(endpointUrl string) bool {
	return endpointProjectIdMatcher.Match([]byte(endpointUrl))
}

func StdRequestOpts() *golangsdk.RequestOpts {
	return &golangsdk.RequestOpts{
		MoreHeaders: map[string]string{"Content-Type": "application/json", "X-Language": "en-us"},
	}
}
