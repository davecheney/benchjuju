// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	"fmt"

	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type unitSuite struct{}

var _ = gc.Suite(&unitSuite{})

func (s *unitSuite) TestUnitTag(c *gc.C) {
	c.Assert(names.NewUnitTag("wordpress/2").String(), gc.Equals, "unit-wordpress-2")
}

var unitNameTests = []struct {
	pattern string
	valid   bool
	service string
}{
	{pattern: "wordpress/42", valid: true, service: "wordpress"},
	{pattern: "rabbitmq-server/123", valid: true, service: "rabbitmq-server"},
	{pattern: "foo", valid: false},
	{pattern: "foo/", valid: false},
	{pattern: "bar/foo", valid: false},
	{pattern: "20/20", valid: false},
	{pattern: "foo-55", valid: false},
	{pattern: "foo-bar/123", valid: true, service: "foo-bar"},
	{pattern: "foo-bar/123/", valid: false},
	{pattern: "foo-bar/123-not", valid: false},
}

func (s *unitSuite) TestUnitNameFormats(c *gc.C) {
	for i, test := range unitNameTests {
		c.Logf("test %d: %q", i, test.pattern)
		c.Assert(names.IsValidUnit(test.pattern), gc.Equals, test.valid)
	}
}

func (s *unitSuite) TestInvalidUnitTagFormats(c *gc.C) {
	for i, test := range unitNameTests {
		if !test.valid {
			c.Logf("test %d: %q", i, test.pattern)
			expect := fmt.Sprintf("%q is not a valid unit name", test.pattern)
			testUnitTag := func() { names.NewUnitTag(test.pattern) }
			c.Assert(testUnitTag, gc.PanicMatches, expect)
		}
	}
}

func (s *serviceSuite) TestUnitService(c *gc.C) {
	for i, test := range unitNameTests {
		c.Logf("test %d: %q", i, test.pattern)
		if !test.valid {
			expect := fmt.Sprintf("%q is not a valid unit name", test.pattern)
			_, err := names.UnitService(test.pattern)
			c.Assert(err, gc.ErrorMatches, expect)
		} else {
			result, err := names.UnitService(test.pattern)
			c.Assert(err, gc.IsNil)
			c.Assert(result, gc.Equals, test.service)
		}
	}
}

var parseUnitTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "unit-dave/0",
	expected: names.NewUnitTag("dave/0"),
}, {
	tag: "dave",
	err: names.InvalidTagError("dave", ""),
}, {
	tag: "unit-dave",
	err: names.InvalidTagError("unit-dave", names.UnitTagKind), // not a valid unit name either
}, {
	tag: "service-dave",
	err: names.InvalidTagError("service-dave", names.UnitTagKind),
}}

func (s *unitSuite) TestParseUnitTag(c *gc.C) {
	for i, t := range parseUnitTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseUnitTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
