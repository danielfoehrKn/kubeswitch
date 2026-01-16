package build

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/opentelekomcloud/gophertelekomcloud/internal/multierr"
)

/*
QueryString is an internal function to be used by request methods in
individual resource packages.

It accepts a tagged structure and expands it into a URL struct. Field names are
converted into query parameters based on a "q" tag. For example:

	type QueryStruct struct {
	   Bar string `q:"x_bar"`
	   Baz int    `q:"lorem_ipsum"`
	}

	instance := QueryStruct{
	   Bar: "AAA",
	   Baz: "BBB",
	}

will be converted into "?x_bar=AAA&lorem_ipsum=BBB".

The struct's fields may be strings, integers, or boolean values. Fields left at
their type's zero value will be omitted from the query.
*/
func QueryString(opts interface{}) (*url.URL, error) {
	if opts == nil {
		return nil, fmt.Errorf("error building query string: %w", ErrNilOpts)
	}

	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr {
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr {
		optsType = optsType.Elem()
	}

	params := url.Values{}

	if optsValue.Kind() != reflect.Struct {
		// Return an error if the underlying type of 'opts' isn't a struct.
		return nil, fmt.Errorf("error building query string: options type is not a struct")
	}

	mErr := multierr.MultiError{}

	for i := 0; i < optsValue.NumField(); i++ {
		v := optsValue.Field(i)
		field := optsType.Field(i)

		// Otherwise, the field is not set.
		// We duplicate the check from ValidateTags to avoid double reflect package usage
		// TODO: investigate performance difference when using ValidateTags
		if v.IsZero() {
			if structFieldRequired(field) {
				// And the field is required. Return an error.
				mErr = append(mErr, fmt.Errorf("required query parameter [%s] not set", field.Name))
			}
			continue // skip empty fields
		}

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// if the field is set, add it to the slice of query pieces
		qTag := field.Tag.Get("q")

		// if the field has a 'q' tag, it goes in the query string
		if qTag == "" {
			continue
		}

		tags := strings.Split(qTag, ",")

		switch v.Kind() {
		case reflect.String:
			params.Add(tags[0], v.String())
		case reflect.Int, reflect.Int32, reflect.Int64:
			params.Add(tags[0], strconv.FormatInt(v.Int(), 10))
		case reflect.Bool:
			params.Add(tags[0], strconv.FormatBool(v.Bool()))
		case reflect.Slice:
			switch v.Type().Elem() {
			case reflect.TypeOf(0):
				for i := 0; i < v.Len(); i++ {
					params.Add(tags[0], strconv.FormatInt(v.Index(i).Int(), 10))
				}
			default:
				for i := 0; i < v.Len(); i++ {
					params.Add(tags[0], v.Index(i).String())
				}
			}
		case reflect.Map:
			keyKind := v.Type().Key().Kind()
			valueKind := v.Type().Elem().Kind()
			if keyKind == reflect.String && valueKind == reflect.String {
				var s []string
				for _, k := range v.MapKeys() {
					value := v.MapIndex(k).String()
					s = append(s, fmt.Sprintf("'%s':'%s'", k.String(), value))
				}
				params.Add(tags[0], fmt.Sprintf("{%s}", strings.Join(s, ", ")))
			} else {
				mErr = append(mErr, fmt.Errorf("expected map[string]string, got map[%s]%s", keyKind, valueKind))
			}
		}
	}

	if err := mErr.ErrorOrNil(); err != nil {
		return nil, fmt.Errorf("error building query string: %w", err)
	}

	return &url.URL{RawQuery: params.Encode()}, nil
}
