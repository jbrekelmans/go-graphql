package mapping

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewFieldInfo(t *testing.T) {
	t.Run("Case1", func(t *testing.T) {
		rt := reflect.TypeOf(struct {
			ID string
		}{})
		id, _ := rt.FieldByName("ID")
		actual := NewFieldInfo(id)
		assert.Equal(t, "id", actual.graphQL)
	})
	t.Run("Case2", func(t *testing.T) {
		rt := reflect.TypeOf(struct {
			OwnerID string
		}{})
		ownerID, _ := rt.FieldByName("OwnerID")
		actual := NewFieldInfo(ownerID)
		assert.Equal(t, "ownerId", actual.graphQL)
	})
	t.Run("Case3", func(t *testing.T) {
		rt := reflect.TypeOf(struct {
			User struct {
				Bio string
			} `graphql:"... on User"`
		}{})
		user, _ := rt.FieldByName("User")
		actual := NewFieldInfo(user)
		assert.Equal(t, "... on User", actual.graphQL)
	})
	t.Run("Case4", func(t *testing.T) {
		rt := reflect.TypeOf(struct {
			Bio string
		}{})
		bio, _ := rt.FieldByName("Bio")
		actual := NewFieldInfo(bio)
		assert.Equal(t, "bio", actual.graphQL)
	})
}
