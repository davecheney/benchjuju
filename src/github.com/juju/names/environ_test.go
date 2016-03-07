// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type environSuite struct{}

var _ = gc.Suite(&environSuite{})

var parseEnvironTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "environment-f47ac10b-58cc-4372-a567-0e02b2c3d479",
	expected: names.NewEnvironTag("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
}, {
	tag: "dave",
	err: names.InvalidTagError("dave", ""),
	//}, {
	// TODO(dfc) passes, but should not
	//	tag: "environment-",
	//	err: names.InvalidTagError("environment", ""),
}, {
	tag: "service-dave",
	err: names.InvalidTagError("service-dave", names.EnvironTagKind),
}}

func (s *environSuite) TestParseEnvironTag(c *gc.C) {
	for i, t := range parseEnvironTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseEnvironTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
