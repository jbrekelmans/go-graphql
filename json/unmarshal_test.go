package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Unmarshal(t *testing.T) {
	t.Run("Correctly unmarshals null (case 1)", func(t *testing.T) {
		json := `{"name":null}`
		var q struct {
			Name *string
		}
		err := Unmarshal([]byte(json), &q)
		if assert.NoError(t, err) {
			assert.Equal(t, (*string)(nil), q.Name)
		}
	})
	t.Run("Correctly unmarshals null in arrays", func(t *testing.T) {
		json := `{"names":[null]}`
		var q struct {
			Names []*string
		}
		err := Unmarshal([]byte(json), &q)
		if assert.NoError(t, err) {
			assert.Equal(t, []*string{nil}, q.Names)
		}
	})
	t.Run("Lazily initializes values of type pointer-to-struct", func(t *testing.T) {
		json := `{"firstName":"Henk"}`
		var q struct {
			Person *struct {
				FirstName string
			} `graphql:"... on Person"`
			Animal *struct {
				Legs int
			} `graphql:"... on Animal"`
		}
		err := Unmarshal([]byte(json), &q)
		if assert.NoError(t, err) {
			assert.Nil(t, q.Animal)
			if assert.NotNil(t, q.Person) {
				assert.Equal(t, "Henk", q.Person.FirstName)
			}
		}
	})
	t.Run("Resets slices", func(t *testing.T) {
		var q struct {
			Names []string
		}
		q.Names = []string{"n1", "n2"}
		json := `{"names":["n3"]}`
		err := Unmarshal([]byte(json), &q)
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"n3"}, q.Names)
		}
	})
	t.Run("Initializes slices", func(t *testing.T) {
		var q struct {
			Person struct {
				Names []string
			} `graphql:"... on Person"`
			Animal struct {
				Legs []struct {
					Length int
				}
			} `graphql:"... on Animal"`
		}
		json := `{"names":["n1"]}`
		err := Unmarshal([]byte(json), &q)
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"n1"}, q.Person.Names)
			assert.Nil(t, q.Animal.Legs)
		}
	})
	t.Run("Branches", func(t *testing.T) {
		var q struct {
			Person struct {
				Age string
			} `graphql:"... on Person"`
			Animal struct {
				Age        string
				NotAPerson struct{}
			} `graphql:"... on Animal"`
		}
		json := `{"age":"123"}`
		err := Unmarshal([]byte(json), &q)
		if assert.NoError(t, err) {
			assert.Equal(t, "123", q.Person.Age)
			assert.Equal(t, "123", q.Animal.Age)
		}
	})
}
