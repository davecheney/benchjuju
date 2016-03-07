// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type subnetSuite struct{}

var _ = gc.Suite(&subnetSuite{})

func (s *subnetSuite) TestNewSubnetTag(c *gc.C) {
	cidr := "10.20.0.0/16"
	tag := names.NewSubnetTag(cidr)
	parsed, err := names.ParseSubnetTag(tag.String())
	c.Assert(err, gc.IsNil)
	c.Assert(parsed.Kind(), gc.Equals, names.SubnetTagKind)
	c.Assert(parsed.Id(), gc.Equals, cidr)
	c.Assert(parsed.String(), gc.Equals, names.SubnetTagKind+"-"+cidr)

	f := func() {
		tag = names.NewSubnetTag("foo")
	}
	c.Assert(f, gc.PanicMatches, "foo is not a valid subnet CIDR")
}

var parseSubnetTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "subnet-10.20.0.0/16",
	expected: names.NewSubnetTag("10.20.0.0/16"),
}, {
	tag:      "subnet-2001:db8::/32",
	expected: names.NewSubnetTag("2001:db8::/32"),
}, {
	tag: "subnet-fe80::3%zone1/10",
	err: names.InvalidTagError("subnet-fe80::3%zone1/10", names.SubnetTagKind),
}, {
	tag: "subnet-10.20.30.40/16",
	err: names.InvalidTagError("subnet-10.20.30.40/16", names.SubnetTagKind),
}, {
	tag: "subnet-2001:db8::123/32",
	err: names.InvalidTagError("subnet-2001:db8::123/32", names.SubnetTagKind),
}, {
	tag: "subnet-foo",
	err: names.InvalidTagError("subnet-foo", names.SubnetTagKind),
}, {
	tag: "subnet-",
	err: names.InvalidTagError("subnet-", names.SubnetTagKind),
}, {
	tag: "foobar",
	err: names.InvalidTagError("foobar", ""),
}, {
	tag: "unit-foo-0",
	err: names.InvalidTagError("unit-foo-0", names.SubnetTagKind),
}}

func (s *subnetSuite) TestParseSubnetTag(c *gc.C) {
	for i, t := range parseSubnetTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseSubnetTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
