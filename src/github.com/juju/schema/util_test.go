// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema_test

import (
	gc "launchpad.net/gocheck"

	"github.com/juju/schema"
)

var aPath = []string{"<pa", "th>"}

type baseSuite struct {
	sch schema.Checker
}

func (s *baseSuite) SetUpTest(c *gc.C) {
}

type Dummy struct{}

func (d *Dummy) Coerce(value interface{}, path []string) (coerced interface{}, err error) {
	return "i-am-dummy", nil
}
