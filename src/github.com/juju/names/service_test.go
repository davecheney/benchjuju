// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type serviceSuite struct{}

var _ = gc.Suite(&serviceSuite{})

var serviceNameTests = []struct {
	pattern string
	valid   bool
}{
	{pattern: "", valid: false},
	{pattern: "wordpress", valid: true},
	{pattern: "foo42", valid: true},
	{pattern: "doing55in54", valid: true},
	{pattern: "%not", valid: false},
	{pattern: "42also-not", valid: false},
	{pattern: "but-this-works", valid: true},
	{pattern: "so-42-far-not-good", valid: false},
	{pattern: "foo/42", valid: false},
	{pattern: "is-it-", valid: false},
	{pattern: "broken2-", valid: false},
	{pattern: "foo2", valid: true},
	{pattern: "foo-2", valid: false},
}

func (s *serviceSuite) TestServiceNameFormats(c *gc.C) {
	assertService := func(s string, expect bool) {
		c.Assert(names.IsValidService(s), gc.Equals, expect)
		// Check that anything that is considered a valid service name
		// is also (in)valid if a(n) (in)valid unit designator is added
		// to it.
		c.Assert(names.IsValidUnit(s+"/0"), gc.Equals, expect)
		c.Assert(names.IsValidUnit(s+"/99"), gc.Equals, expect)
		c.Assert(names.IsValidUnit(s+"/-1"), gc.Equals, false)
		c.Assert(names.IsValidUnit(s+"/blah"), gc.Equals, false)
		c.Assert(names.IsValidUnit(s+"/"), gc.Equals, false)
	}

	for i, test := range serviceNameTests {
		c.Logf("test %d: %q", i, test.pattern)
		assertService(test.pattern, test.valid)
	}
}

var parseServiceTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "service-dave",
	expected: names.NewServiceTag("dave"),
}, {
	tag: "dave",
	err: names.InvalidTagError("dave", ""),
}, {
	tag: "service-dave/0",
	err: names.InvalidTagError("service-dave/0", names.ServiceTagKind),
}, {
	tag: "service",
	err: names.InvalidTagError("service", ""),
}, {
	tag: "user-dave",
	err: names.InvalidTagError("user-dave", names.ServiceTagKind),
}}

func (s *serviceSuite) TestParseServiceTag(c *gc.C) {
	for i, t := range parseServiceTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseServiceTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
