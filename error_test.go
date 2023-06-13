package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ErrorItem(t *testing.T) {
	t.Run("MarshalJSON", func(t *testing.T) {
		t.Run("Case1", func(t *testing.T) {
			var e *ErrorItem
			b, err := e.MarshalJSON()
			if assert.NoError(t, err) {
				assert.Equal(t, string(b), "null")
			}
		})
		t.Run("Case2", func(t *testing.T) {
			e := ErrorItem{
				Raw: map[string]json.RawMessage{
					"type": json.RawMessage("null"),
				},
			}
			b, err := e.MarshalJSON()
			if assert.NoError(t, err) {
				assert.Equal(t, string(b), `{"type":null}`)
			}
		})
	})
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("Case1", func(t *testing.T) {
			e := ErrorItem{
				Message: "this-should-be-unchanged",
			}
			err := e.UnmarshalJSON([]byte("null"))
			if assert.NoError(t, err) {
				assert.Equal(t, ErrorItem{
					Message: "this-should-be-unchanged",
				}, e)
			}
		})
		t.Run("Case2", func(t *testing.T) {
			var e ErrorItem
			err := e.UnmarshalJSON([]byte("bad-json"))
			assert.Truef(t, err != nil && strings.HasPrefix(err.Error(), `error in (*graphql.ErrorItem).UnmarshalJSON: `), "unexpected err: %v", err)
		})
		t.Run("Case3", func(t *testing.T) {
			var e ErrorItem
			err := e.UnmarshalJSON([]byte(`{"message":"some msg"}`))
			if assert.NoError(t, err) {
				assert.Equal(t, ErrorItem{
					Message: "some msg",
					Raw: map[string]json.RawMessage{
						"message": json.RawMessage(`"some msg"`),
					},
				}, e)
			}
		})
		t.Run("Case4", func(t *testing.T) {
			var e ErrorItem
			err := e.UnmarshalJSON([]byte(`{"type":"X"}`))
			if assert.NoError(t, err) {
				assert.Equal(t, ErrorItem{
					Raw: map[string]json.RawMessage{
						"type": json.RawMessage(`"X"`),
					},
				}, e)
			}
		})
	})
}

func Test_getOrCreateError(t *testing.T) {
	t.Run("Case1", func(t *testing.T) {
		in := &Error{}
		out := getOrCreateError(in)
		assert.Same(t, in, out)
	})
	t.Run("Case2", func(t *testing.T) {
		innerError := errors.New("oops")
		outerError := fmt.Errorf("error doing XYZ: %w", innerError)
		in := outerError
		out := getOrCreateError(in)
		assert.Equal(t, &Error{
			Err:     outerError,
			Message: outerError.Error(),
		}, out)
		assert.Same(t, outerError, out.Err)
	})
}

func Test_setErrorOperation(t *testing.T) {
	t.Run("Case1", func(t *testing.T) {
		var in error
		out := setErrorOperation(in, "query{me{name}}")
		assert.Equal(t, nil, out)
	})
	t.Run("Case2", func(t *testing.T) {
		in := errors.New("oops")
		out := setErrorOperation(in, "query{me{name}}")
		if assert.Equal(t, &Error{
			Err:       in,
			Message:   "oops",
			Operation: "query{me{name}}",
		}, out) {
			out2 := out.(*Error)
			assert.Same(t, in, out2.Err)
		}
	})
	t.Run("Case3", func(t *testing.T) {
		in := &Error{Message: "oops"}
		out := setErrorOperation(in, "query{me{name}}")
		if assert.Equal(t, &Error{
			Message:   "oops",
			Operation: "query{me{name}}",
		}, out) {
			out2 := out.(*Error)
			assert.Same(t, in, out2)
		}
	})
}

func Test_setErrorItems(t *testing.T) {
	t.Run("Case1", func(t *testing.T) {
		var in error
		out := setErrorItems(in, []ErrorItem{{Message: "x"}})
		assert.Equal(t, nil, out)
	})
	t.Run("Case2", func(t *testing.T) {
		in := errors.New("oops")
		errItems := []ErrorItem{{Message: "x"}}
		out := setErrorItems(in, errItems)
		if assert.Equal(t, &Error{
			Err:     in,
			Message: "oops",
			Errors:  []ErrorItem{{Message: "x"}},
		}, out) {
			out2 := out.(*Error)
			assert.Same(t, in, out2.Err)
			assert.Same(t, &errItems[0], &out2.Errors[0])
		}
	})
	t.Run("Case3", func(t *testing.T) {
		in := &Error{Message: "oops"}
		errItems := []ErrorItem{{Message: "x"}}
		out := setErrorItems(in, errItems)
		if assert.Equal(t, &Error{
			Message: "oops",
			Errors:  []ErrorItem{{Message: "x"}},
		}, out) {
			out2 := out.(*Error)
			assert.Same(t, in, out2)
			assert.Same(t, &errItems[0], &out2.Errors[0])
		}
	})
}
