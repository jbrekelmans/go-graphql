package graphql

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/jbrekelmans/go-graphql/mapping"
)

type queryBuilder struct {
	b bytes.Buffer
}

func (qb *queryBuilder) operation(operationType string, q any, variables map[string]any) error {
	qb.raw(operationType)
	qb.varDefs(variables)
	n := qb.b.Len()
	qb.selectionSetHelper(reflect.TypeOf(q), false)
	if qb.b.Len() == n {
		return fmt.Errorf(`invalid %s type %T`, operationType, q)
	}
	return nil
}

func (qb *queryBuilder) selectionSetHelper(t reflect.Type, inline bool) {
	switch t.Kind() {
	// NOTE: even if we add support for arrays here, unmarshaling JSON into Go arrays is not supported.
	// TODO add support for unmarshaling into Go arrays.
	case reflect.Ptr, reflect.Slice:
		qb.selectionSetHelper(t.Elem(), false)
	case reflect.Struct:
		if !inline {
			qb.b.WriteByte('{')
		}
		n := 0
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			n++
			if n == 2 {
				qb.b.WriteByte(',')
			}
			x := mapping.NewFieldInfo(t.Field(i))
			if !x.Inline() {
				qb.raw(x.GraphQL())
			}
			qb.selectionSetHelper(f.Type, x.Inline())

		}
		if !inline {
			qb.b.WriteByte('}')
		}
	}
}

func (qb *queryBuilder) raw(s string) {
	qb.b.WriteString(s)
}

func (qb *queryBuilder) String() string {
	return qb.b.String()
}

func (qb *queryBuilder) Type(t reflect.Type) {
	nonNull := true
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		nonNull = false
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		qb.b.WriteByte('[')
		qb.Type(t.Elem())
		qb.b.WriteByte(']')
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		qb.raw("Int")
	case reflect.String:
		qb.raw("String")
	case reflect.Float32, reflect.Float64:
		qb.raw("Float")
	case reflect.Bool:
		qb.raw("Boolean")
	default:
		name := t.Name()
		qb.raw(name)
	}
	if nonNull {
		qb.b.WriteByte('!')
	}
}

func (qb *queryBuilder) varDefs(variables map[string]any) {
	// https://spec.graphql.org/October2021/#VariableDefinitions
	n := len(variables)
	if n == 0 {
		return
	}
	qb.b.WriteByte('(')
	for varName, v := range variables {
		qb.b.WriteByte('$')
		qb.raw(varName)
		qb.b.WriteByte(':')
		t := reflect.TypeOf(v)
		qb.Type(t)
	}
	qb.b.WriteByte(')')
}
