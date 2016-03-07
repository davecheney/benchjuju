// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	"fmt"
	"regexp"

	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type spaceSuite struct{}

var _ = gc.Suite(&spaceSuite{})

var spaceNameTests = []struct {
	pattern string
	valid   bool
}{
	{pattern: "", valid: false},
	{pattern: "eth0", valid: true},
	{pattern: "-my-net-", valid: false},
	{pattern: "42", valid: true},
	{pattern: "%not", valid: false},
	{pattern: "$PATH", valid: false},
	{pattern: "but-this-works", valid: true},
	{pattern: "----", valid: false},
	{pattern: "oh--no", valid: false},
	{pattern: "777", valid: true},
	{pattern: "is-it-", valid: false},
	{pattern: "also_not", valid: false},
	{pattern: "a--", valid: false},
	{pattern: "foo-2", valid: true},
}

func (s *spaceSuite) TestSpaceNames(c *gc.C) {
	for i, test := range spaceNameTests {
		c.Logf("test %d: %q", i, test.pattern)
		c.Check(names.IsValidSpace(test.pattern), gc.Equals, test.valid)
		if test.valid {
			expectTag := fmt.Sprintf("%s-%s", names.SpaceTagKind, test.pattern)
			c.Check(names.NewSpaceTag(test.pattern).String(), gc.Equals, expectTag)
		} else {
			expectErr := fmt.Sprintf("%q is not a valid space name", test.pattern)
			testTag := func() { names.NewSpaceTag(test.pattern) }
			c.Check(testTag, gc.PanicMatches, regexp.QuoteMeta(expectErr))
		}
	}
}

var parseSpaceTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "space-1",
	expected: names.NewSpaceTag("1"),
}, {
	tag: "-space1",
	err: names.InvalidTagError("-space1", ""),
}}

func (s *spaceSuite) TestParseSpaceTag(c *gc.C) {
	for i, t := range parseSpaceTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseSpaceTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
