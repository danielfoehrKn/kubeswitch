package pagination

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/internal/extract"
)

// LinkedPageBase may be embedded to implement a page that provides navigational "Next" and "Previous" links within its result.
type LinkedPageBase struct {
	PageResult

	// LinkPath lists the keys that should be traversed within a response to arrive at the "next" pointer.
	// If any link along the path is missing, an empty URL will be returned.
	// If any link results in an unexpected value type, an error will be returned.
	// When left as "nil", []string{"links", "next"} will be used as a default.
	LinkPath []string
}

// NextPageURL extracts the pagination structure from a JSON response and returns the "next" link, if one is present.
// It assumes that the links are available in a "links" element of the top-level response object.
// If this is not the case, override NextPageURL on your result type.
func (current LinkedPageBase) NextPageURL() (string, error) {
	var path []string
	var key string

	if current.LinkPath == nil {
		path = []string{"links", "next"}
	} else {
		path = current.LinkPath
	}

	submap := make(map[string]any)

	err := extract.Into(bytes.NewReader(current.Body), &submap)
	if err != nil {
		err := golangsdk.ErrUnexpectedType{}
		err.Expected = "map[string]any"
		err.Actual = fmt.Sprintf("%v", reflect.TypeOf(current.Body))
		return "", err
	}

	for {
		key, path = path[0], path[1:]

		value, ok := submap[key]
		if !ok {
			return "", nil
		}

		if len(path) > 0 {
			submap, ok = value.(map[string]any)
			if !ok {
				err := golangsdk.ErrUnexpectedType{}
				err.Expected = "map[string]any"
				err.Actual = fmt.Sprintf("%v", reflect.TypeOf(value))
				return "", err
			}
		} else {
			if value == nil {
				// Actual null element.
				return "", nil
			}

			url, ok := value.(string)
			if !ok {
				err := golangsdk.ErrUnexpectedType{}
				err.Expected = "string"
				err.Actual = fmt.Sprintf("%v", reflect.TypeOf(value))
				return "", err
			}

			return url, nil
		}
	}
}

// IsEmpty satisfies the IsEmpty method of the Page interface
func (current LinkedPageBase) IsEmpty() (bool, error) {
	body, err := current.GetBodyAsSlice()
	if err != nil {
		return false, fmt.Errorf("error converting page body to slice: %w", err)
	}

	return len(body) == 0, nil
}

// GetBody returns the linked page's body. This method is needed to satisfy the
// Page interface.
func (current LinkedPageBase) GetBody() []byte {
	return current.Body
}

// WrapNextPageURL function use makerID to warp next page url,it returns the full url for request.
func (current LinkedPageBase) WrapNextPageURL(markerID string) (string, error) {
	limit := current.URL.Query().Get("limit")

	if limit == "" {
		return "", nil
	}

	q := current.URL.Query()

	q.Set("marker", markerID)
	current.URL.RawQuery = q.Encode()
	return current.URL.String(), nil
}
