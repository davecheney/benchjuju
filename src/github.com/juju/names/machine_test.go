// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	stdtesting "testing"

	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type machineSuite struct{}

var _ = gc.Suite(&machineSuite{})

func Test(t *stdtesting.T) {
	gc.TestingT(t)
}

func (s *machineSuite) TestMachineTag(c *gc.C) {
	c.Assert(names.NewMachineTag("10").String(), gc.Equals, "machine-10")
	// Check a container id.
	c.Assert(names.NewMachineTag("10/lxc/1").String(), gc.Equals, "machine-10-lxc-1")
}

var machineIdTests = []struct {
	pattern   string
	valid     bool
	container bool
}{
	{pattern: "42", valid: true},
	{pattern: "042", valid: false},
	{pattern: "0", valid: true},
	{pattern: "foo", valid: false},
	{pattern: "/", valid: false},
	{pattern: "55/", valid: false},
	{pattern: "1/foo", valid: false},
	{pattern: "2/foo/", valid: false},
	{pattern: "3/lxc/42", valid: true, container: true},
	{pattern: "3/lxc-nodash/42", valid: false},
	{pattern: "0/lxc/00", valid: false},
	{pattern: "0/lxc/0/", valid: false},
	{pattern: "03/lxc/42", valid: false},
	{pattern: "3/lxc/042", valid: false},
	{pattern: "4/foo/bar", valid: false},
	{pattern: "5/lxc/42/foo", valid: false},
	{pattern: "6/lxc/42/kvm/0", valid: true, container: true},
	{pattern: "06/lxc/42/kvm/0", valid: false},
	{pattern: "6/lxc/042/kvm/0", valid: false},
	{pattern: "6/lxc/42/kvm/00", valid: false},
	{pattern: "06/lxc/042/kvm/00", valid: false},
}

func (s *machineSuite) TestMachineIdFormats(c *gc.C) {
	for i, test := range machineIdTests {
		c.Logf("test %d: %q", i, test.pattern)
		c.Assert(names.IsValidMachine(test.pattern), gc.Equals, test.valid)
		c.Assert(names.IsContainerMachine(test.pattern), gc.Equals, test.container)
	}
}

var parseMachineTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{{
	tag: "",
	err: names.InvalidTagError("", ""),
}, {
	tag:      "machine-0",
	expected: names.NewMachineTag("0"),
}, {
	tag: "machine-one",
	err: names.InvalidTagError("machine-one", names.MachineTagKind),
}, {
	tag: "dave",
	err: names.InvalidTagError("dave", ""),
}, {
	tag: "user-one",
	err: names.InvalidTagError("user-one", names.MachineTagKind),
}}

func (s *machineSuite) TestParseMachineTag(c *gc.C) {
	for i, t := range parseMachineTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseMachineTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}
