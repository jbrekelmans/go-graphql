package json

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jbrekelmans/go-graphql/mapping"
)

type receivers []reflect.Value

func newReceivers(rv reflect.Value) receivers {
	var r receivers
	r.add(rv)
	return r
}

func (r *receivers) append(rv reflect.Value) {
	*r = append(*r, rv)
}

// add adds rv. If rv is a (pointer to) struct then the fields of the struct that define inline fragments are also added.
// This process is recursive. See code for the exact criteria for when fields of structs are added.
func (r *receivers) add(rv reflect.Value) {
	t := unwrapPointerType(rv.Type())
	if t.Kind() != reflect.Struct {
		r.append(rv)
		return
	}
	// rv should be added to r, which is done below so we can lazily initialize structs.

	// Recursively walk through struct fields.
	// We use a queue instead of recursive calls of add,
	// so we can use the local variable "seen".

	// Initialize queue with single element rv.
	// The initial memory storing the queue elements is stack-allocated.
	var queueArray [3]reflect.Value
	queueArray[0] = rv
	queue := stack[reflect.Value](queueArray[:1])

	// Track types to avoid infinite recursion for self-referential types.
	var seenArray [3]reflect.Type
	seenArray[0] = t
	seen := stack[reflect.Type](seenArray[:1])

	for !queue.empty() {
		rv := queue.pop()
		rvAdded := false
		structType := unwrapPointerType(rv.Type())
		for i := 0; i < structType.NumField(); i++ {
			structField := structType.Field(i)
			if !structField.IsExported() {
				continue
			}
			fieldInfo := mapping.NewFieldInfo(structField)
			if !fieldInfo.Inline() && !fieldInfo.IsInlineFragment() {
				continue
			}
			t := unwrapPointerType(structField.Type)
			if t.Kind() != reflect.Struct || stackContains(seen, t) {
				continue
			}
			if !rvAdded {
				rv = elemIfPointer(rv)
				r.append(rv)
				rvAdded = true
			}
			seen.push(t)
			queue.push(rv.Field(i))
		}
		if !rvAdded {
			r.append(rv)
		}
	}
}

func (r receivers) copy() receivers {
	return append(receivers(nil), r...)
}

// mapArrayElement derives a set of receivers from r that should receive an element of a JSON array.
func (r receivers) mapArrayElement() receivers {
	var recvNext receivers
	for _, rv := range r {
		// rv must be a slice type.
		if rv.Kind() != reflect.Slice {
			panic(fmt.Errorf(`receivers.mapArrayElement: r contains non-slice type %v`, rv.Type()))
		}
		elemType := rv.Type().Elem()
		zeroRV := reflect.Zero(elemType)
		rv.Set(reflect.Append(rv, zeroRV))
		rv := rv.Index(rv.Len() - 1)
		recvNext.add(rv)
	}
	return recvNext
}

func (r receivers) mapArrayStartInPlace() error {
	// TODO support unmarshaling JSON arrays into Go arrays.
	for i, rv := range r {
		t := unwrapPointerType(rv.Type())
		if t.Kind() != reflect.Slice {
			return fmt.Errorf(`cannot unmarshal JSON array into non-slice type %v`, rv.Type())
		}
		rv := elemIfPointer(rv)
		if n := rv.Len(); n > 0 {
			rv.SetLen(0)
		}
		r[i] = rv
	}
	return nil
}

// mapPropertyName derives a set of receivers from r that should receive the value of a JSON object
// property named propertyName.
// For each receiver in r: if the receiver is a struct and has a field mapped to propertyName, then
// field is added to the result.
// Returns an error if the result would be empty (as this indicates a bug in the bigger picture: we selected a field in
// GraphQL but there is no location to unmarshal).
func (r receivers) mapPropertyName(propertyName string) (receivers, error) {
	var recvNext receivers
	for _, rv := range r {
		t := unwrapPointerType(rv.Type())
		if t.Kind() != reflect.Struct {
			continue
		}
		for i := 0; i < t.NumField(); i++ {
			structField := t.Field(i)
			if !structField.IsExported() {
				continue
			}
			fieldInfo := mapping.NewFieldInfo(structField)
			fieldName := fieldInfo.FieldName()
			if fieldName == "" {
				continue
			}
			if !strings.EqualFold(fieldName, propertyName) {
				continue
			}
			rv = elemIfPointer(rv)
			recvNext.add(rv.Field(i))
		}
	}
	if len(recvNext) == 0 {
		// TODO add path of property to error msg
		return nil, fmt.Errorf(`JSON object has property named %#v but no receiver struct has a field mapped to that property`,
			propertyName)
	}
	return recvNext, nil
}

// same returns true if the first elements of r and other are stored at the same memory address.
// If r or other is empty then returns false.
func (r receivers) same(other receivers) bool {
	return len(r) > 0 && len(other) > 0 && &r[0] == &other[0]
}

// unmarshal unmarshals JSON into each receiver in r.
func (r receivers) unmarshalJSON(jsonBytes []byte) error {
	if string(jsonBytes) == "null" {
		return nil
	}
	for _, rv := range r {
		rv := elemIfPointer(rv)
		receiverInterf := rv.Addr().Interface()
		if err := json.Unmarshal(jsonBytes, receiverInterf); err != nil {
			return err
		}
	}
	return nil
}

// unmarshalAny marshals v to JSON and then unmarshals the result into each receiver in r.
func (r receivers) unmarshalAny(v any) error {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.unmarshalJSON(jsonBytes)
}
