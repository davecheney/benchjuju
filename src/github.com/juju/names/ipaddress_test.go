// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type ipAddressSuite struct{}

var _ = gc.Suite(&ipAddressSuite{})

func (s *ipAddressSuite) TestNewIPAddressTag(c *gc.C) {
	uuid := utils.MustNewUUID()
	tag := names.NewIPAddressTag(uuid.String())
	parsed, err := names.ParseIPAddressTag(tag.String())
	c.Assert(err, gc.IsNil)
	c.Assert(parsed.Kind(), gc.Equals, names.IPAddressTagKind)
	c.Assert(parsed.Id(), gc.Equals, uuid.String())
	c.Assert(parsed.String(), gc.Equals, names.IPAddressTagKind+"-"+uuid.String())

	f := func() {
		tag = names.NewIPAddressTag("42")
	}
	c.Assert(f, gc.PanicMatches, `invalid UUID: "42"`)
}

var parseIPAddressTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{
	{tag: "", err: names.InvalidTagError("", "")},
	{tag: "ipaddress-42424242-1111-2222-3333-0123456789ab", expected: names.NewIPAddressTag("42424242-1111-2222-3333-0123456789ab")},
	{tag: "ipaddress-012345678", err: names.InvalidTagError("ipaddress-012345678", names.IPAddressTagKind)},
	{tag: "ipaddress-42", err: names.InvalidTagError("ipaddress-42", names.IPAddressTagKind)},
	{tag: "foobar", err: names.InvalidTagError("foobar", "")},
	{tag: "space-yadda", err: names.InvalidTagError("space-yadda", names.IPAddressTagKind)}}

func (s *ipAddressSuite) TestParseIPAddressTag(c *gc.C) {
	for i, t := range parseIPAddressTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseIPAddressTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
