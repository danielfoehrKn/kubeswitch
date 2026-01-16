package build

import (
	"fmt"
	"reflect"

	"github.com/opentelekomcloud/gophertelekomcloud/internal/multierr"
)

type fieldValue struct {
	// Name of the field in the type.
	Name string
	// Value of the field in the structure.
	Value reflect.Value
	// `required` tag value.
	TagRequired bool
	// `xor` tag value.
	TagXOR string
	// `or` tag value.
	TagOR string
}

// ValidateTags validating structure by tags.
//
// Supported validations:
//
//	required (`required:"true"`)  - mark field required, returns error if it is empty.
//	or:      (`or:"OtherField"`)  - requires at least one field to be not empty.
//	xor:     (`xor:"OtherField"`) - requires exactly of this and the other field to be set.
func ValidateTags(opts interface{}) error {
	if opts == nil {
		return nil // nil is an ideal value
	}

	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr {
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr {
		optsType = optsType.Elem()
	}

	fields := make(map[string]fieldValue)

	if optsValue.Kind() != reflect.Struct {
		return nil // no need to go deep
	}

	// fill the structure fields map
	for i := 0; i < optsValue.NumField(); i++ {
		value := optsValue.Field(i)
		field := optsType.Field(i)

		fields[field.Name] = fieldValue{
			Name:        field.Name,
			Value:       value,
			TagRequired: structFieldRequired(field),
			TagXOR:      field.Tag.Get("xor"),
			TagOR:       field.Tag.Get("or"),
		}
	}

	errors := multierr.MultiError{}

	for name, field := range fields {
		fieldErrors := make([]error, 0)

		if field.TagRequired && field.Value.IsZero() {
			fieldErrors = append(fieldErrors,
				fmt.Errorf("missing input for argument [%s]", name),
			)
		}

		orField := field.TagOR
		if orField != "" && field.Value.IsZero() && fields[orField].Value.IsZero() {
			fieldErrors = append(fieldErrors,
				fmt.Errorf("at least one of %s and %s must be provided", name, orField),
			)
		}

		xorField := field.TagXOR
		if xorField != "" && (field.Value.IsZero() == fields[xorField].Value.IsZero()) {
			fieldErrors = append(fieldErrors,
				fmt.Errorf("exactly one of %s and %s must be provided", name, xorField),
			)
		}

		errors = append(errors, fieldErrors...)
	}

	return errors.ErrorOrNil()
}

func structFieldRequired(field reflect.StructField) bool {
	return field.Tag.Get("required") == "true"
}
