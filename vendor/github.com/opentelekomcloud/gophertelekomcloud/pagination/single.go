package pagination

import (
	"fmt"
)

// SinglePageBase may be embedded in a Page that contains all the results from an operation at once.
// Deprecated: use element slice as a return result.
type SinglePageBase PageResult

func (current SinglePageBase) GetBody() []byte {
	return current.Body
}

func (current SinglePageBase) GetBodyAsSlice() ([]interface{}, error) {
	return PageResult(current).GetBodyAsSlice()
}

func (current SinglePageBase) GetBodyAsMap() (map[string]interface{}, error) {
	return PageResult(current).GetBodyAsMap()
}

// NextPageURL always returns "" to indicate that there are no more pages to return.
func (current SinglePageBase) NextPageURL() (string, error) {
	return "", nil
}

// IsEmpty satisfies the IsEmpty method of the Page interface
func (current SinglePageBase) IsEmpty() (bool, error) {
	body, err := current.GetBodyAsSlice()
	if err != nil {
		return false, fmt.Errorf("error converting page body to slice: %w", err)
	}

	return len(body) == 0, nil
}

// NewSinglePageBase may be embedded in a Page that contains all the results from an operation at once.
type NewSinglePageBase struct {
	NewPageResult
}

// NewNextPageURL always returns "" to indicate that there are no more pages to return.
func (current NewSinglePageBase) NewNextPageURL() (string, error) {
	return "", nil
}

// NewIsEmpty satisfies the IsEmpty method of the Page interface
func (current NewSinglePageBase) NewIsEmpty() (bool, error) {
	body, err := current.NewGetBodyAsSlice()
	if err != nil {
		return false, fmt.Errorf("error converting page body to slice: %w", err)
	}

	return len(body) == 0, nil
}
