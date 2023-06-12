// FieldInfo defines how fields of Go structs map to GraphQL.
package mapping

import (
	"reflect"
	"strings"

	"github.com/shurcooL/graphql/ident"
)

// FieldInfo defines how a field of a Go struct maps to GraphQL,
// both when constructing operations (queries/mutations) and when
// unmarshalling the response.
type FieldInfo struct {
	inline  bool
	graphQL string
}

// NewFieldInfo maps a field of a Go struct to GraphQL.
func NewFieldInfo(f reflect.StructField) FieldInfo {
	var fieldInfo FieldInfo
	tag, hasTag := f.Tag.Lookup("graphql")
	fieldInfo.inline = f.Anonymous && !hasTag
	if !fieldInfo.inline {
		if hasTag {
			// TODO validate tag further (i.e. we need to remove commas, reject # chars)
			fieldInfo.graphQL = tag
		} else {
			fieldInfo.graphQL = ident.ParseMixedCaps(f.Name).ToLowerCamelCase()
		}
	}
	// TODO validate that type is a struct (wrapped by any amount of pointers) if it's an inline fragment.
	return fieldInfo
}

// IsInlineFragment returns true if f.GraphQL() is an InlineFragment production
// (without the trailing selection set). I.e. the field of the Go struct defines
// an inline fragment.
// See https://spec.graphql.org/October2021/#sec-Selection-Sets.
func (f FieldInfo) IsInlineFragment() bool {
	return strings.HasPrefix(strings.TrimSpace(f.graphQL), "...")
}

// GraphQL returns a GraphQL snippet. Recall that Go structs correspond to selection sets
// in GraphQL and Go struct fields corresond to selections.
// See https://spec.graphql.org/October2021/#sec-Selection-Sets.
// In particular, the returned GraphQL snippet satisfies the InlineFragment or Field
// production in the language, except for the trailing SelectionSet.
// For example, this can return "... on User" or "name".
func (f FieldInfo) GraphQL() string {
	return f.graphQL
}

// FieldName names the field in GraphQL.
// Aliases are not supported.
func (f FieldInfo) FieldName() string {
	if f.Inline() || f.IsInlineFragment() {
		return ""
	}
	graphQL := strings.TrimSpace(f.graphQL)
	if i := strings.IndexAny(graphQL, "(:@"); i >= 0 {
		return strings.TrimSpace(graphQL[:i])
	}
	return graphQL
}

// Inline returns true if the Go struct field has a struct type (i.e. a struct contained in
// a struct) and the "inner" struct fields should be "inlined" in the "outer" struct fields
// when constructing GraphQL selection sets.
func (f FieldInfo) Inline() bool {
	return f.inline
}
