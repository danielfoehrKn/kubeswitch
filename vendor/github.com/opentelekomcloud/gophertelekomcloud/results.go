package golangsdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/opentelekomcloud/gophertelekomcloud/internal/extract"
)

/*
Result is an internal type to be used by individual resource packages, but its
methods will be available on a wide variety of user-facing embedding types.

It acts as a base struct that other Result types, returned from request
functions, can embed for convenience. All Results capture basic information
from the HTTP transaction that was performed, including the response body,
HTTP headers, and any errors that happened.

Generally, each Result type will have an Extract method that can be used to
further interpret the result's payload in a specific context. Extensions or
providers can then provide additional extraction functions to pull out
provider- or extension-specific information as well.

Deprecated: use functions from internal/extract package instead
*/
type Result struct {
	// Body is the payload of the HTTP response from the server.
	Body []byte

	// Header contains the HTTP header structure from the original response.
	Header http.Header

	// Err is an error that occurred during the operation. It's deferred until
	// extraction to make it easier to chain the Extract call.
	Err error
}

// BodyReader returns cached body as *bytes.Reader
func (r Result) BodyReader() io.Reader {
	return bytes.NewReader(r.Body)
}

type JsonRDSInstanceStatus struct {
	Instances  []JsonRDSInstanceField `json:"instances"`
	TotalCount int                    `json:"total_count"`
}

type JsonRDSInstanceField struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ExtractInto allows users to provide an object into which `Extract` will extract
// the `Result.Body`. This would be useful for OpenStack providers that have
// different fields in the response object than OpenStack proper.
//
// Deprecated: use extract.Into function instead
func (r Result) ExtractInto(to interface{}) error {
	if r.Err != nil {
		return r.Err
	}

	return extract.Into(bytes.NewReader(r.Body), to)
}

// ExtractIntoStructPtr will unmarshal the Result (r) into the provided
// interface{} (to).
//
// NOTE: For internal use only
//
// `to` must be a pointer to an underlying struct type
//
// If provided, `label` will be filtered out of the response
// body prior to `r` being unmarshalled into `to`.
//
// Deprecated: use extract.IntoStructPtr function instead
func (r Result) ExtractIntoStructPtr(to interface{}, label string) error {
	if r.Err != nil {
		return r.Err
	}

	return extract.IntoStructPtr(bytes.NewReader(r.Body), to, label)
}

// ExtractIntoSlicePtr will unmarshal the Result (r) into the provided
// interface{} (to).
//
// NOTE: For internal use only
//
// `to` must be a pointer to an underlying slice type
//
// If provided, `label` will be filtered out of the response
// body prior to `r` being unmarshalled into `to`.
//
// Deprecated: use extract.IntoSlicePtr function instead
func (r Result) ExtractIntoSlicePtr(to interface{}, label string) error {
	if r.Err != nil {
		return r.Err
	}

	return extract.IntoSlicePtr(bytes.NewReader(r.Body), to, label)
}

func PrettyPrintJSON(body interface{}) string {
	pretty, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return string(pretty)
}

// PrettyPrintJSON creates a string containing the full response body as
// pretty-printed JSON. It's useful for capturing test fixtures and for
// debugging extraction bugs. If you include its output in an issue related to
// a buggy extraction function, we will all love you forever.
func (r Result) PrettyPrintJSON() string {
	return PrettyPrintJSON(r.Body)
}

// ErrResult is an internal type to be used by individual resource packages, but
// its methods will be available on a wide variety of user-facing embedding
// types.
//
// It represents results that only contain a potential error and
// nothing else. Usually, if the operation executed successfully, the Err field
// will be nil; otherwise it will be stocked with a relevant error. Use the
// ExtractErr method
// to cleanly pull it out.
//
// Deprecated: use plain err return instead
type ErrResult struct {
	Result
}

// ExtractErr is a function that extracts error information, or nil, from a result.
func (r ErrResult) ExtractErr() error {
	return r.Err
}

// ----------------------------------------------------------------------------

type ErrRespond struct {
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type ErrWithResult struct {
	ErrResult
}

func (r Result) Extract() (*ErrRespond, error) {
	var s = ErrRespond{}
	if r.Err != nil {
		return nil, r.Err
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return nil, fmt.Errorf("failed to extract Error with Result")
	}
	return &s, nil
}

// ----------------------------------------------------------------------------

/*
HeaderResult is an internal type to be used by individual resource packages, but
its methods will be available on a wide variety of user-facing embedding types.

It represents a result that only contains an error (possibly nil) and an
http.Header. This is used, for example, by the objectstorage packages in
openstack, because most of the operations don't return response bodies, but do
have relevant information in headers.
*/
type HeaderResult struct {
	Result
}

// ExtractInto allows users to provide an object into which `Extract` will
// extract the http.Header headers of the result.
func (r HeaderResult) ExtractInto(to interface{}) error {
	if r.Err != nil {
		return r.Err
	}

	tmpHeaderMap := map[string]string{}
	for k, v := range r.Header {
		if len(v) > 0 {
			tmpHeaderMap[k] = v[0]
		}
	}

	b, err := extract.JsonMarshal(tmpHeaderMap)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, to)

	return err
}

// RFC3339Milli describes a common time format used by some API responses.
const RFC3339Milli = "2006-01-02T15:04:05.999999Z"

type JSONRFC3339Milli time.Time

func (jt *JSONRFC3339Milli) UnmarshalJSON(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := json.NewDecoder(b)
	var s string
	if err := dec.Decode(&s); err != nil {
		return err
	}
	t, err := time.Parse(RFC3339Milli, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC3339Milli(t)
	return nil
}

const RFC3339MilliNoZ = "2006-01-02T15:04:05.999999"

type JSONRFC3339MilliNoZ time.Time

func (jt *JSONRFC3339MilliNoZ) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(RFC3339MilliNoZ, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC3339MilliNoZ(t)
	return nil
}

type JSONRFC1123 time.Time

func (jt *JSONRFC1123) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC1123, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC1123(t)
	return nil
}

type JSONUnix time.Time

func (jt *JSONUnix) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	unix, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	t = time.Unix(unix, 0)
	*jt = JSONUnix(t)
	return nil
}

// RFC3339NoZ is the time format used in Heat (Orchestration).
const RFC3339NoZ = "2006-01-02T15:04:05"

type JSONRFC3339NoZ time.Time

func (jt *JSONRFC3339NoZ) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(RFC3339NoZ, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC3339NoZ(t)
	return nil
}

// RFC3339ZNoT is the time format used in Zun (Containers Service).
const RFC3339ZNoT = "2006-01-02 15:04:05-07:00"

type JSONRFC3339ZNoT time.Time

func (jt *JSONRFC3339ZNoT) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(RFC3339ZNoT, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC3339ZNoT(t)
	return nil
}

// RFC3339ZNoTNoZ is another time format used in Zun (Containers Service).
const RFC3339ZNoTNoZ = "2006-01-02 15:04:05"

type JSONRFC3339ZNoTNoZ time.Time

func (jt *JSONRFC3339ZNoTNoZ) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(RFC3339ZNoTNoZ, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC3339ZNoTNoZ(t)
	return nil
}

/*
Link is an internal type to be used in packages of collection resources that are
paginated in a certain way.

It's a response substructure common to many paginated collection results that is
used to point to related pages. Usually, the one we care about is the one with
Rel field set to "next".
*/
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

/*
ExtractNextURL is an internal function useful for packages of collection
resources that are paginated in a certain way.

It attempts to extract the "next" URL from slice of Link structs, or
"" if no such URL is present.
*/
func ExtractNextURL(links []Link) (string, error) {
	var url string

	for _, l := range links {
		if l.Rel == "next" {
			url = l.Href
		}
	}

	if url == "" {
		return "", nil
	}

	return url, nil
}
