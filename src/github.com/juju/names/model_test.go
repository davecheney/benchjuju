// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type modelSuite struct{}

var _ = gc.Suite(&modelSuite{})

var parseModelTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "model-f47ac10b-58cc-4372-a567-0e02b2c3d479",
	expected: names.NewModelTag("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
}, {
	tag: "dave",
	err: names.InvalidTagError("dave", ""),
	//}, {
	// TODO(dfc) passes, but should not
	//	tag: "model-",
	//	err: names.InvalidTagError("model", ""),
}, {
	tag: "service-dave",
	err: names.InvalidTagError("service-dave", names.ModelTagKind),
}}

func (s *modelSuite) TestParseModelTag(c *gc.C) {
	for i, t := range parseModelTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseModelTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
