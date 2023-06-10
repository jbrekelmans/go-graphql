package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// Unmarshal recursively processes the JSON data `b` and unmarshals the JSON values into the appropriate
// locations in `v`.
// `v` must be a non-nil pointer to a struct.
// Unmarshal, unlike "encoding/json".Unmarshal, handles Go structs that define inline fragments of
// GraphQL queries (and other special behaviour defined by the graphql package).
//
// The json.Unmarshaler interface is _not_ respected when JSON objects and arrays are being unmarshaled.
// That is, UnmarshalJSON will never be called with JSON that is an object or array.
//
// Unmarshaling JSON arrays into Go arrays is not supported and attempting to do so returns an error.
func Unmarshal(b []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf(`v has non-pointer type %T`, v)
	}
	if rv.IsNil() {
		return fmt.Errorf(`v is nil`)
	}
	if unwrapPointerType(rv.Type()).Kind() != reflect.Struct {
		return fmt.Errorf(`v is not a pointer-to-struct type`)
	}
	jsonDec := json.NewDecoder(bytes.NewReader(b))
	jsonDec.UseNumber()

	u := unmarshaler{tokens: jsonDec}
	err := u.Run(rv)
	if err != nil {
		return err
	}
	_, err = jsonDec.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf(`JSON data contains extraneous token`)
}
