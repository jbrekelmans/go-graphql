package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ID(t *testing.T) {
	t.Run("MarshalJSON", func(t *testing.T) {
		t.Run("Case1", func(t *testing.T) {
			id := ID{S: "x"}
			b, err := id.MarshalJSON()
			if assert.NoError(t, err) {
				assert.Equal(t, []byte(`"x"`), b)
			}
		})
	})
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("Case1", func(t *testing.T) {
			var id ID
			err := id.UnmarshalJSON([]byte(`"x"`))
			if assert.NoError(t, err) {
				assert.Equal(t, ID{S: "x"}, id)
			}
		})
		t.Run("Case2", func(t *testing.T) {
			id := ID{S: "x"}
			err := id.UnmarshalJSON([]byte(`null`))
			if assert.NoError(t, err) {
				assert.Equal(t, ID{S: "x"}, id)
			}
		})
		t.Run("Case3", func(t *testing.T) {
			var id ID
			err := id.UnmarshalJSON([]byte(`1`))
			assert.Error(t, err)
		})
	})
}
