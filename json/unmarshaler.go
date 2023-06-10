package json

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type stateItem struct {
	// inObject is true if the inner-most array or object being decoded is an object.
	// false if it is an array.
	inObject bool
	recv     receivers
}

type unmarshaler struct {
	tokens *json.Decoder
	state  stack[stateItem]
}

// Run recursively walks through JSON values and unmarshals them into the appropriate values.
// The json.Unmarshaler interface is _not_ respected when JSON objects and arrays are being decoded.
// That is, UnmarshalJSON will never be called with JSON that is an object or array.
func (u *unmarshaler) Run(rv reflect.Value) error {
	token, err := u.tokens.Token()
	if err != nil {
		return eofToUnexpected(err)
	}
	if token != json.Delim('{') {
		return fmt.Errorf(`JSON value must be an object`)
	}
	u.state = stack[stateItem]{
		stateItem{
			inObject: true,
			recv:     newReceivers(rv),
		},
	}
	for len(u.state) > 0 {
		token, err = u.tokens.Token()
		if err != nil {
			return eofToUnexpected(err)
		}
		s := u.state.top()
		recv := s.recv
		if s.inObject {
			if token != json.Delim('}') {
				// Token is name of property of object.
				propertyName := token.(string)
				recv, err = recv.mapPropertyName(propertyName)
				if err != nil {
					return err
				}
				token, err = u.tokens.Token()
				if err != nil {
					return eofToUnexpected(err)
				}
			}
		} else if token != json.Delim(']') {
			// In JSON array.
			recv = recv.mapArrayElement()
		}
		switch {
		case token == json.Delim('{'):
			u.state.push(stateItem{
				inObject: true,
				recv:     recv,
			})
		case token == json.Delim('}'):
			u.state.pop()
			// TODO consider recycling receivers
		case token == json.Delim('['):
			recvNext := recv
			if u.state.top().recv.same(recv) {
				recvNext = recv.copy()
			}
			recvNext.mapArrayStartInPlace()
			u.state.push(stateItem{
				recv: recv,
			})
		case token == json.Delim(']'):
			u.state.pop()
			// TODO consider recycling receivers
		default:
			if err := recv.unmarshalAny(token); err != nil {
				return err
			}
		}
	}
	return nil
}

func elemIfPointer(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	return rv
}

func eofToUnexpected(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

func unwrapPointerType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
