// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

import (
	"fmt"
	"reflect"
)

// Const returns a Checker that only succeeds if the input matches
// value exactly.  The value is compared with reflect.DeepEqual.
func Const(value interface{}) Checker {
	return constC{value}
}

type constC struct {
	value interface{}
}

func (c constC) Coerce(v interface{}, path []string) (interface{}, error) {
	if reflect.DeepEqual(v, c.value) {
		return v, nil
	}
	return nil, error_{fmt.Sprintf("%#v", c.value), v, path}
}
