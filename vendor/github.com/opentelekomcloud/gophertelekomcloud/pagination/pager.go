package pagination

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/opentelekomcloud/gophertelekomcloud"
)

// Page must be satisfied by the result type of any resource collection.
// It allows clients to interact with the resource uniformly, regardless of whether or not or how it's paginated.
// Generally, rather than implementing this interface directly, implementors should embed one of the concrete PageBase structs,
// instead.
// Depending on the pagination strategy of a particular resource, there may be an additional subinterface that the result type
// will need to implement.
type Page interface {
	// NextPageURL generates the URL for the page of data that follows this collection.
	// Return "" if no such page exists.
	NextPageURL() (string, error)

	// IsEmpty returns true if this Page has no items in it.
	IsEmpty() (bool, error)

	// GetBody returns the Page Body. This is used in the `AllPages` method.
	GetBody() []byte
	// GetBodyAsSlice tries to convert page body to a slice.
	GetBodyAsSlice() ([]any, error)
	// GetBodyAsMap tries to convert page body to a map.
	GetBodyAsMap() (map[string]any, error)
}

// Pager knows how to advance through a specific resource collection, one page at a time.
type Pager struct {
	client *golangsdk.ServiceClient
	Client *golangsdk.ServiceClient

	initialURL string
	InitialURL string

	createPage func(r PageResult) Page
	CreatePage func(r NewPageResult) NewPage

	Err error

	// Headers supplies additional HTTP headers to populate on each paged request.
	Headers map[string]string
}

// NewPager constructs a manually-configured pager.
// Supply the URL for the first page, a function that requests a specific page given a URL, and a function that counts a page.
func NewPager(client *golangsdk.ServiceClient, initialURL string, createPage func(r PageResult) Page) Pager {
	return Pager{
		client:     client,
		initialURL: initialURL,
		createPage: createPage,
	}
}

func (p Pager) fetchNextPage(url string) (Page, error) {
	resp, err := Request(p.client, p.Headers, url)
	if err != nil {
		return nil, err
	}

	remembered, err := PageResultFrom(resp)
	if err != nil {
		return nil, err
	}

	return p.createPage(remembered), nil
}

// EachPage iterates over each page returned by a Pager, yielding one at a time to a handler function.
// Return "false" from the handler to prematurely stop iterating.
func (p Pager) EachPage(handler func(Page) (bool, error)) error {
	if p.Err != nil {
		return p.Err
	}
	currentURL := p.initialURL
	for {
		currentPage, err := p.fetchNextPage(currentURL)
		if err != nil {
			return err
		}

		empty, err := currentPage.IsEmpty()
		if err != nil {
			return err
		}
		if empty {
			return nil
		}

		ok, err := handler(currentPage)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		currentURL, err = currentPage.NextPageURL()
		if err != nil {
			return err
		}
		if currentURL == "" {
			return nil
		}
	}
}

// AllPages returns all the pages from a `List` operation in a single page,
// allowing the user to retrieve all the pages at once.
func (p Pager) AllPages() (Page, error) {
	// body will contain the final concatenated Page body.
	var body []byte

	// Grab a test page to ascertain the page body type.
	testPage, err := p.fetchNextPage(p.initialURL)
	if err != nil {
		return nil, err
	}
	// Store the page type, so we can use reflection to create a new mega-page of
	// that type.
	pageType := reflect.TypeOf(testPage)

	// if it's a single page, just return the testPage (first page)
	if _, found := pageType.FieldByName("SinglePageBase"); found {
		return testPage, nil
	}

	if _, err := testPage.GetBodyAsSlice(); err == nil {
		var pagesSlice []any

		// Iterate over the pages to concatenate the bodies.
		err = p.EachPage(func(page Page) (bool, error) {
			b, err := page.GetBodyAsSlice()
			if err != nil {
				return false, fmt.Errorf("error paginating page with slice body: %w", err)
			}
			pagesSlice = append(pagesSlice, b...)
			return true, nil
		})
		if err != nil {
			return nil, err
		}

		body, err = json.Marshal(pagesSlice)
		if err != nil {
			return nil, err
		}
	} else if _, err := testPage.GetBodyAsMap(); err == nil {
		var pagesSlice []any

		// key is the map key for the page body if the body type is `map[string]any`.
		var key string
		// Iterate over the pages to concatenate the bodies.
		err = p.EachPage(func(page Page) (bool, error) {
			b, err := page.GetBodyAsMap()
			if err != nil {
				return false, fmt.Errorf("error paginating page with map body: %w", err)
			}
			for k, v := range b {
				// If it's a linked page, we don't want the `links`, we want the other one.
				if !strings.HasSuffix(k, "links") {
					// check the field's type. we only want []any (which is really []map[string]interface{})
					switch vt := v.(type) {
					case []any:
						key = k
						pagesSlice = append(pagesSlice, vt...)
					}
				}
			}
			return true, nil
		})
		if err != nil {
			return nil, err
		}

		mapBody := map[string]any{
			key: pagesSlice,
		}

		body, err = json.Marshal(mapBody)
		if err != nil {
			return nil, err
		}
	} else {
		var pagesSlice [][]byte

		// Iterate over the pages to concatenate the bodies.
		err = p.EachPage(func(page Page) (bool, error) {
			b := page.GetBody()
			pagesSlice = append(pagesSlice, b)
			// separate pages with a comma
			pagesSlice = append(pagesSlice, []byte{10})
			return true, nil
		})
		if err != nil {
			return nil, err
		}
		if len(pagesSlice) > 0 {
			// Remove the trailing comma.
			pagesSlice = pagesSlice[:len(pagesSlice)-1]
		}
		var b []byte
		// Combine the slice of slices in to a single slice.
		for _, slice := range pagesSlice {
			b = append(b, slice...)
		}

		body = b
	}

	// Each `Extract*` function is expecting a specific type of page coming back,
	// otherwise the type assertion in those functions will fail. pageType is needed
	// to create a type in this method that has the same type that the `Extract*`
	// function is expecting and set the Body of that object to the concatenated
	// pages.
	page := reflect.New(pageType)
	// Set the page body to be the concatenated pages.
	page.Elem().FieldByName("Body").Set(reflect.ValueOf(body))
	// Set any additional headers that were pass along. The `objectstorage` pacakge,
	// for example, passes a Content-Type header.
	h := make(http.Header)
	for k, v := range p.Headers {
		h.Add(k, v)
	}
	page.Elem().FieldByName("Header").Set(reflect.ValueOf(h))
	// Type assert the page to a Page interface so that the type assertion in the
	// `Extract*` methods will work.
	return page.Elem().Interface().(Page), err
}

// NewPage must be satisfied by the result type of any resource collection.
// It allows clients to interact with the resource uniformly, regardless of whether or not or how it's paginated.
// Generally, rather than implementing this interface directly, implementors should embed one of the concrete PageBase structs,
// instead.
// Depending on the pagination strategy of a particular resource, there may be an additional subinterface that the result type
// will need to implement.
type NewPage interface {
	// NewNextPageURL generates the URL for the page of data that follows this collection.
	// Return "" if no such page exists.
	NewNextPageURL() (string, error)

	// NewIsEmpty returns true if this Page has no items in it.
	NewIsEmpty() (bool, error)

	// NewGetBody returns the Page Body. This is used in the `AllPages` method.
	NewGetBody() []byte
	// NewGetBodyAsSlice tries to convert page body to a slice.
	NewGetBodyAsSlice() ([]any, error)
	// NewGetBodyAsMap tries to convert page body to a map.
	NewGetBodyAsMap() (map[string]any, error)
}

func (p Pager) newFetchNextPage(url string) (NewPage, error) {
	resp, err := Request(p.Client, p.Headers, url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return p.CreatePage(NewPageResult{
		Body:   rawBody,
		Header: resp.Header,
		URL:    *resp.Request.URL,
	}), nil
}

// NewEachPage iterates over each page returned by a Pager, yielding one at a time to a handler function.
// Return "false" from the handler to prematurely stop iterating.
func (p Pager) NewEachPage(handler func(NewPage) (bool, error)) error {
	if p.Err != nil {
		return p.Err
	}
	currentURL := p.InitialURL
	for {
		currentPage, err := p.newFetchNextPage(currentURL)
		if err != nil {
			return err
		}

		empty, err := currentPage.NewIsEmpty()
		if err != nil {
			return err
		}
		if empty {
			return nil
		}

		ok, err := handler(currentPage)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		currentURL, err = currentPage.NewNextPageURL()
		if err != nil {
			return err
		}
		if currentURL == "" {
			return nil
		}
	}
}

// NewAllPages returns all the pages from a `List` operation in a single page,
// allowing the user to retrieve all the pages at once.
func (p Pager) NewAllPages() (NewPage, error) {
	// body will contain the final concatenated Page body.
	var body []byte

	// Grab a test page to ascertain the page body type.
	testPage, err := p.newFetchNextPage(p.InitialURL)
	if err != nil {
		return nil, err
	}
	// Store the page type, so we can use reflection to create a new mega-page of
	// that type.
	pageType := reflect.TypeOf(testPage)

	// if it's a single page, just return the testPage (first page)
	if _, found := pageType.FieldByName("NewSinglePageBase"); found {
		return testPage, nil
	}

	if _, err := testPage.NewGetBodyAsSlice(); err == nil {
		var pagesSlice []any

		// Iterate over the pages to concatenate the bodies.
		err = p.NewEachPage(func(page NewPage) (bool, error) {
			b, err := page.NewGetBodyAsSlice()
			if err != nil {
				return false, fmt.Errorf("error paginating page with slice body: %w", err)
			}
			pagesSlice = append(pagesSlice, b...)
			return true, nil
		})
		if err != nil {
			return nil, err
		}

		body, err = json.Marshal(pagesSlice)
		if err != nil {
			return nil, err
		}
	} else if _, err := testPage.NewGetBodyAsMap(); err == nil {
		var pagesSlice []any

		// key is the map key for the page body if the body type is `map[string]any`.
		var key string
		// Iterate over the pages to concatenate the bodies.
		err = p.NewEachPage(func(page NewPage) (bool, error) {
			b, err := page.NewGetBodyAsMap()
			if err != nil {
				return false, fmt.Errorf("error paginating page with map body: %w", err)
			}
			for k, v := range b {
				// If it's a linked page, we don't want the `links`, we want the other one.
				if !strings.HasSuffix(k, "links") {
					// check the field's type. we only want []any (which is really []map[string]interface{})
					switch vt := v.(type) {
					case []any:
						key = k
						pagesSlice = append(pagesSlice, vt...)
					}
				}
			}
			return true, nil
		})
		if err != nil {
			return nil, err
		}

		mapBody := map[string]any{
			key: pagesSlice,
		}

		body, err = json.Marshal(mapBody)
		if err != nil {
			return nil, err
		}
	} else {
		var pagesSlice [][]byte

		// Iterate over the pages to concatenate the bodies.
		err = p.NewEachPage(func(page NewPage) (bool, error) {
			b := page.NewGetBody()
			pagesSlice = append(pagesSlice, b)
			// separate pages with a comma
			pagesSlice = append(pagesSlice, []byte{10})
			return true, nil
		})
		if err != nil {
			return nil, err
		}
		if len(pagesSlice) > 0 {
			// Remove the trailing comma.
			pagesSlice = pagesSlice[:len(pagesSlice)-1]
		}
		var b []byte
		// Combine the slice of slices in to a single slice.
		for _, slice := range pagesSlice {
			b = append(b, slice...)
		}

		body = b
	}

	// Each `Extract*` function is expecting a specific type of page coming back,
	// otherwise the type assertion in those functions will fail. pageType is needed
	// to create a type in this method that has the same type that the `Extract*`
	// function is expecting and set the Body of that object to the concatenated
	// pages.
	page := reflect.New(pageType)
	// Set the page body to be the concatenated pages.
	page.Elem().FieldByName("Body").Set(reflect.ValueOf(body))
	// Set any additional headers that were pass along. The `objectstorage` pacakge,
	// for example, passes a Content-Type header.
	h := make(http.Header)
	for k, v := range p.Headers {
		h.Add(k, v)
	}
	page.Elem().FieldByName("Header").Set(reflect.ValueOf(h))
	// Type assert the page to a Page interface so that the type assertion in the
	// `Extract*` methods will work.
	return page.Elem().Interface().(NewPage), err
}
