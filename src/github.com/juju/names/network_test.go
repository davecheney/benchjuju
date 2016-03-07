// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	"fmt"
	"regexp"

	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type networkSuite struct{}

var _ = gc.Suite(&networkSuite{})

var networkNameTests = []struct {
	pattern string
	valid   bool
}{
	{pattern: "", valid: false},
	{pattern: "eth0", valid: true},
	{pattern: "-my-net-", valid: true},
	{pattern: "42", valid: true},
	{pattern: "%not", valid: false},
	{pattern: "$PATH", valid: false},
	{pattern: "but-this-works", valid: true},
	{pattern: "----", valid: true},
	{pattern: "oh--no", valid: true},
	{pattern: "777", valid: true},
	{pattern: "is-it-", valid: true},
	{pattern: "also_not", valid: true},
	{pattern: "a--", valid: true},
	{pattern: "foo-2", valid: true},
	{pattern: "MAAS_n3t-w0rK-", valid: true},
}

func (s *networkSuite) TestNetworkNames(c *gc.C) {
	for i, test := range networkNameTests {
		c.Logf("test %d: %q", i, test.pattern)
		c.Check(names.IsValidNetwork(test.pattern), gc.Equals, test.valid)
		if test.valid {
			expectTag := fmt.Sprintf("%s-%s", names.NetworkTagKind, test.pattern)
			c.Check(names.NewNetworkTag(test.pattern).String(), gc.Equals, expectTag)
		} else {
			expectErr := fmt.Sprintf("%q is not a valid network name", test.pattern)
			testNetworkTag := func() { names.NewNetworkTag(test.pattern) }
			c.Check(testNetworkTag, gc.PanicMatches, regexp.QuoteMeta(expectErr))
		}
	}
}

var parseNetworkTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "network-dave",
	expected: names.NewNetworkTag("dave"),
}, {
	tag: "dave",
	err: names.InvalidTagError("dave", ""),
}, {
	tag: "network-dave/0",
	err: names.InvalidTagError("network-dave/0", names.NetworkTagKind),
}, {
	tag: "network",
	err: names.InvalidTagError("network", ""),
}, {
	tag: "user-dave",
	err: names.InvalidTagError("user-dave", names.NetworkTagKind),
}}

func (s *networkSuite) TestParseNetworkTag(c *gc.C) {
	for i, t := range parseNetworkTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseNetworkTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
