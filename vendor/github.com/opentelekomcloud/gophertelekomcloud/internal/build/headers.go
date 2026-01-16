package build

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/opentelekomcloud/gophertelekomcloud/internal/multierr"
)

/*
Headers is an internal function to be used by request methods in
individual resource packages.

It accepts an arbitrary tagged structure and produces a string map that's
suitable for use as the HTTP headers of an outgoing request. Field names are
mapped to header names based in "h" tags.

	type struct QueryStruct {
	  Bar string `h:"x_bar"`
	  Baz int    `h:"lorem_ipsum"`
	}

	instance := QueryStruct{
	  Bar: "AAA",
	  Baz: "BBB",
	}

will be converted into:

	map[string]string{
	  "x_bar": "AAA",
	  "lorem_ipsum": "BBB",
	}

Untagged fields and fields left at their zero values are skipped. Integers,
booleans and string values are supported.
*/
func Headers(opts interface{}) (map[string]string, error) {
	if opts == nil {
		return nil, fmt.Errorf("error building headers: %w", ErrNilOpts)
	}

	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr {
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr {
		optsType = optsType.Elem()
	}

	if optsValue.Kind() != reflect.Struct {
		// Return an error if the underlying type of 'opts' isn't a struct.
		return nil, fmt.Errorf("error building headers: options type is not a struct")
	}

	mErr := multierr.MultiError{}
	result := make(map[string]string)

	for i := 0; i < optsValue.NumField(); i++ {
		value := optsValue.Field(i)
		field := optsType.Field(i)

		headerName := field.Tag.Get("h")
		if headerName == "" {
			continue
		}

		if value.IsZero() {
			// We duplicate the check from ValidateTags to avoid double reflect package usage
			// TODO: investigate performance difference when using ValidateTags
			if structFieldRequired(field) {
				mErr = append(mErr, fmt.Errorf("required header [%s] not set", field.Name))
			}
			continue
		}

		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		var headerValue string

		// if the field is set, add it to the slice of query pieces
		switch value.Kind() {
		case reflect.String:
			headerValue = value.String()
		case reflect.Int, reflect.Int32, reflect.Int64:
			headerValue = strconv.FormatInt(value.Int(), 10)
		case reflect.Bool:
			headerValue = strconv.FormatBool(value.Bool())
		default:
			mErr = append(mErr, fmt.Errorf("value of unsupported type %s", value.Type()))
		}

		result[headerName] = headerValue
	}

	if err := mErr.ErrorOrNil(); err != nil {
		return nil, fmt.Errorf("error building headers: %w", err)
	}

	return result, nil
}
