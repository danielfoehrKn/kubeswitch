package pagination

import (
	"fmt"
	"strconv"
)

type OffsetPage interface {
	// LastElement returning index of the last element of the page
	LastElement() int
}

type OffsetPageBase struct {
	Offset int
	Limit  int

	PageResult
}

func (p OffsetPageBase) LastElement() int {
	q := p.URL.Query()
	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil {
		offset = p.Offset
		q.Set("offset", strconv.Itoa(offset))
	}
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil {
		limit = p.Limit
		q.Set("limit", strconv.Itoa(limit))
	}
	return offset + limit
}

func (p OffsetPageBase) NextPageURL() (string, error) {
	currentURL := p.URL
	q := currentURL.Query()
	if q.Get("offset") == "" && q.Get("limit") == "" {
		// without offset and limit it's just a SinglePageBase
		return "", nil
	}
	q.Set("offset", strconv.Itoa(p.LastElement()))
	currentURL.RawQuery = q.Encode()
	return currentURL.String(), nil
}

// IsEmpty returns true if this Page has no items in it.
func (p OffsetPageBase) IsEmpty() (bool, error) {
	body, err := p.GetBodyAsSlice()
	if err != nil {
		return false, fmt.Errorf("error converting page body to slice: %w", err)
	}

	return len(body) == 0, nil
}

// GetBody returns the Page Body. This is used in the `AllPages` method.
func (p OffsetPageBase) GetBody() []byte {
	return p.Body
}
