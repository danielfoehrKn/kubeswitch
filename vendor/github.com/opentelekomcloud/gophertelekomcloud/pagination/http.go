package pagination

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/internal/extract"
)

// PageResult stores the HTTP response that returned the current page of results.
type PageResult struct {
	golangsdk.Result
	url.URL
}

func (r PageResult) GetBody() []byte {
	return r.Body
}

// GetBodyAsSlice tries to convert page body to a slice, returning nil on fail
func (r PageResult) GetBodyAsSlice() ([]any, error) {
	result := make([]any, 0)

	if err := extract.Into(bytes.NewReader(r.Body), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetBodyAsMap tries to convert page body to a map, returning nil on fail
func (r PageResult) GetBodyAsMap() (map[string]any, error) {
	result := make(map[string]any, 0)

	if err := extract.Into(bytes.NewReader(r.Body), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// PageResultFrom parses an HTTP response as JSON and returns a PageResult containing the
// results, interpreting it as JSON if the content type indicates.
func PageResultFrom(resp *http.Response) (PageResult, error) {
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return PageResult{}, err
	}

	return PageResult{
		Result: golangsdk.Result{
			Body:   rawBody,
			Header: resp.Header,
		},
		URL: *resp.Request.URL,
	}, nil
}

// Request performs an HTTP request and extracts the http.Response from the result.
func Request(client *golangsdk.ServiceClient, headers map[string]string, url string) (*http.Response, error) {
	return client.Get(url, nil, &golangsdk.RequestOpts{
		MoreHeaders: headers,
		OkCodes:     []int{200, 204, 300},
	})
}

// NewPageResult stores the HTTP response that returned the current page of results.
type NewPageResult struct {
	// Body is the payload of the HTTP response from the server.
	Body []byte

	// Header contains the HTTP header structure from the original response.
	Header http.Header

	URL url.URL
}

func (r NewPageResult) NewGetBody() []byte {
	return r.Body
}

// NewGetBodyAsMap tries to convert page body to a map, returning nil on fail
func (r NewPageResult) NewGetBodyAsMap() (map[string]any, error) {
	result := make(map[string]any, 0)

	if err := extract.Into(bytes.NewReader(r.Body), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// NewGetBodyAsSlice tries to convert page body to a slice, returning nil on fail
func (r NewPageResult) NewGetBodyAsSlice() ([]any, error) {
	result := make([]any, 0)

	if err := extract.Into(bytes.NewReader(r.Body), &result); err != nil {
		return nil, err
	}

	return result, nil
}
