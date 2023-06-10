package graphql

import (
	"encoding/json"
	"fmt"
)

// ID corresponds to the GraphQL ID type.
//
// Users should only need this type when passing IDs as
// variable values to queries/mutations, so this package
// can declare the variable in GraphQL with the correct
// type.
type ID struct {
	S string
}

func (i ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.S)
}

func (i *ID) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf(`error in (*ID).UnmarshalJSON: %w`, err)
	}
	i.S = s
	return nil
}
