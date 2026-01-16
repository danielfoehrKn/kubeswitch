package build

import (
	"encoding/json"
	"fmt"
)

// Body wraps original request data providing root tag support.
//
// Example:
//
//	wrapped := structure {
//		Field string `json:"field"`
//	} {"data"}
//
//	Body{
//		RootTag: "root",
//		Wrapped: wrapped,
//	}
//
// Will produce the following json:
//
//	{
//	  "root": {"field": "data"}
//	}
type Body struct {
	RootTag string
	Wrapped interface{}
}

// prepared returns request body, wrapped into a root tag, if required.
func (r Body) prepared() interface{} {
	if r.RootTag == "" {
		return r.Wrapped
	}

	return map[string]interface{}{
		r.RootTag: r.Wrapped,
	}
}

// MarshalJSON satisfies json.Marshaler interface.
func (r Body) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.prepared())
}

// String allows simple pretty-print of prepared value.
func (r Body) String() string {
	jsonData, err := json.MarshalIndent(r.prepared(), "", "  ")
	if err != nil {
		return fmt.Sprintf("!err: %s", err.Error())
	}

	return string(jsonData)
}

// RequestBody validates given structure by its tags and build the body ready to be marshalled to the JSON.
func RequestBody(opts interface{}, parent string) (*Body, error) {
	if opts == nil {
		return nil, fmt.Errorf("error building request body: %w", ErrNilOpts)
	}

	if err := ValidateTags(opts); err != nil {
		return nil, fmt.Errorf("error building request body: %w", err)
	}

	return &Body{
		RootTag: parent,
		Wrapped: opts,
	}, nil
}

func (r Body) ToMap() (map[string]interface{}, error) {
	var res map[string]interface{}

	marshal, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(marshal, &res)
	return res, err
}

func RequestBodyMap(opts interface{}, parent string) (map[string]interface{}, error) {
	body, err := RequestBody(opts, parent)
	if err != nil {
		return nil, err
	}

	return body.ToMap()
}
