package extract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

func intoPtr(body io.Reader, to interface{}, label string) error {
	if label == "" {
		return Into(body, &to)
	}

	var m map[string]interface{}
	err := Into(body, &m)
	if err != nil {
		return err
	}

	b, err := JsonMarshal(m[label])
	if err != nil {
		return err
	}

	toValue := reflect.ValueOf(to)
	if toValue.Kind() == reflect.Ptr {
		toValue = toValue.Elem()
	}

	switch toValue.Kind() {
	case reflect.Slice:
		typeOfV := toValue.Type().Elem()
		if typeOfV.Kind() == reflect.Struct {
			if typeOfV.NumField() > 0 && typeOfV.Field(0).Anonymous {
				newSlice := reflect.MakeSlice(reflect.SliceOf(typeOfV), 0, 0)

				for _, v := range m[label].([]interface{}) {
					// For each iteration of the slice, we create a new struct.
					// This is to work around a bug where elements of a slice
					// are reused and not overwritten when the same copy of the
					// struct is used:
					//
					// https://github.com/golang/go/issues/21092
					// https://github.com/golang/go/issues/24155
					// https://play.golang.org/p/NHo3ywlPZli
					newType := reflect.New(typeOfV).Elem()

					b, err := JsonMarshal(v)
					if err != nil {
						return err
					}

					// This is needed for structs with an UnmarshalJSON method.
					// Technically this is just unmarshalling the response into
					// a struct that is never used, but it's good enough to
					// trigger the UnmarshalJSON method.
					for i := 0; i < newType.NumField(); i++ {
						s := newType.Field(i).Addr().Interface()

						// Unmarshal is used rather than NewDecoder to also work
						// around the above-mentioned bug.
						err = json.Unmarshal(b, s)
						if err != nil {
							continue
						}
					}

					newSlice = reflect.Append(newSlice, newType)
				}

				// "to" should now be properly modeled to receive the
				// JSON response body and unmarshal into all the correct
				// fields of the struct or composed extension struct
				// at the end of this method.
				toValue.Set(newSlice)
			}
		}
	case reflect.Struct:
		typeOfV := toValue.Type()
		if typeOfV.NumField() > 0 && typeOfV.Field(0).Anonymous {
			for i := 0; i < toValue.NumField(); i++ {
				toField := toValue.Field(i)
				if toField.Kind() == reflect.Struct {
					s := toField.Addr().Interface()
					err = json.NewDecoder(bytes.NewReader(b)).Decode(s)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	err = json.Unmarshal(b, &to)
	return err
}

// JsonMarshal marshals input to bytes via buffer with disabled HTML escaping.
func JsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)
	err := enc.Encode(t)
	return buffer.Bytes(), err
}

// Into parses input as JSON and convert to a structure.
func Into(body io.Reader, to interface{}) error {
	if closer, ok := body.(io.ReadCloser); ok {
		defer closer.Close()
	}

	byteBody, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("error reading from stream: %w", err)
	}

	if len(byteBody) == 0 {
		return nil // empty body - nothing to extract
	}

	err = json.Unmarshal(byteBody, to)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error extracting %s into %T: %w", byteBody, to, err)
	}

	return nil
}

func typeCheck(to interface{}, kind reflect.Kind) error {
	t := reflect.TypeOf(to)
	if k := t.Kind(); k != reflect.Ptr {
		return fmt.Errorf("expected pointer, got %v", k)
	}

	if kind != t.Elem().Kind() {
		return fmt.Errorf("expected pointer to %v, got: %v", kind.String(), t)
	}

	return nil
}

// IntoStructPtr will unmarshal the given body into the provided Struct.
func IntoStructPtr(body io.Reader, to interface{}, label string) error {
	err := typeCheck(to, reflect.Struct)
	if err != nil {
		return err
	}

	return intoPtr(body, to, label)
}

// IntoSlicePtr will unmarshal the provided body into the provided Slice.
func IntoSlicePtr(body io.Reader, to interface{}, label string) error {
	err := typeCheck(to, reflect.Slice)
	if err != nil {
		return err
	}

	return intoPtr(body, to, label)
}
