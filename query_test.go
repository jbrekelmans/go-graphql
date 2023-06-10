package graphql

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_queryBuilder(t *testing.T) {
	t.Run("selectionSetHelper", func(t *testing.T) {
		t.Run("Case1", func(t *testing.T) {
			type Struct1 struct {
				Field1 string
			}
			type Struct2 struct {
				Field2 string
				Struct1
			}
			var qb queryBuilder
			qb.selectionSetHelper(reflect.TypeOf(Struct2{}), false)
			assert.Equal(t, "{field2,field1}", qb.String())
		})
	})
	t.Run("operation", func(t *testing.T) {
		t.Run("Case1", func(t *testing.T) {
			var q int
			var qb queryBuilder
			err := qb.operation(`query`, &q, nil)
			assert.Truef(t,
				err != nil && err.Error() == `invalid query type *int`,
				"unexpected err: %v", err,
			)
		})
	})
	t.Run("varDefs", func(t *testing.T) {
		t.Run("Case1", func(t *testing.T) {
			var qb queryBuilder
			qb.varDefs(map[string]any{
				"id": ID{"123"},
			})
			assert.Equal(t, "($id:ID!)", qb.String())
		})
	})
}
