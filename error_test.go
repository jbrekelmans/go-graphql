package graphql

import (
	"encoding/json"
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
