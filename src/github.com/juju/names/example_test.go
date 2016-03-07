// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import "fmt"

func ExampleParseTag() {
	tag, err := ParseTag("user-100")
	if err != nil {
		panic(err)
	}
	switch tag := tag.(type) {
	case UserTag:
		fmt.Printf("User tag, id: %s\n", tag.Id())
	default:
		fmt.Printf("Unknown tag, type %T\n", tag)
	}
}
