// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	"fmt"

	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type volumeSuite struct{}

var _ = gc.Suite(&volumeSuite{})

func (s *volumeSuite) TestVolumeTag(c *gc.C) {
	c.Assert(names.NewVolumeTag("1").String(), gc.Equals, "volume-1")
}

func (s *volumeSuite) TestVolumeNameValidity(c *gc.C) {
	assertVolumeNameValid(c, "0")
	assertVolumeNameValid(c, "0/lxc/0/0")
	assertVolumeNameValid(c, "1000")
	assertVolumeNameInvalid(c, "-1")
	assertVolumeNameInvalid(c, "")
	assertVolumeNameInvalid(c, "one")
	assertVolumeNameInvalid(c, "#")
	assertVolumeNameInvalid(c, "0/0/0") // 0/0 is not a valid machine ID
}

func (s *volumeSuite) TestParseVolumeTag(c *gc.C) {
	assertParseVolumeTag(c, "volume-0", names.NewVolumeTag("0"))
	assertParseVolumeTag(c, "volume-88", names.NewVolumeTag("88"))
	assertParseVolumeTag(c, "volume-0-lxc-0-88", names.NewVolumeTag("0/lxc/0/88"))
	assertParseVolumeTagInvalid(c, "", names.InvalidTagError("", ""))
	assertParseVolumeTagInvalid(c, "one", names.InvalidTagError("one", ""))
	assertParseVolumeTagInvalid(c, "volume-", names.InvalidTagError("volume-", names.VolumeTagKind))
	assertParseVolumeTagInvalid(c, "machine-0", names.InvalidTagError("machine-0", names.VolumeTagKind))
}

func (s *volumeSuite) TestVolumeMachine(c *gc.C) {
	assertVolumeMachine(c, "0/0", names.NewMachineTag("0"))
	assertVolumeMachine(c, "0/lxc/0/0", names.NewMachineTag("0/lxc/0"))
	assertVolumeNoMachine(c, "0")
}

func assertVolumeMachine(c *gc.C, id string, expect names.MachineTag) {
	t, ok := names.VolumeMachine(names.NewVolumeTag(id))
	c.Assert(ok, gc.Equals, true)
	c.Assert(t, gc.Equals, expect)
}

func assertVolumeNoMachine(c *gc.C, id string) {
	_, ok := names.VolumeMachine(names.NewVolumeTag(id))
	c.Assert(ok, gc.Equals, false)
}

func assertVolumeNameValid(c *gc.C, name string) {
	c.Assert(names.IsValidVolume(name), gc.Equals, true)
	names.NewVolumeTag(name)
}

func assertVolumeNameInvalid(c *gc.C, name string) {
	c.Assert(names.IsValidVolume(name), gc.Equals, false)
	testVolumeTag := func() { names.NewVolumeTag(name) }
	expect := fmt.Sprintf("%q is not a valid volume ID", name)
	c.Assert(testVolumeTag, gc.PanicMatches, expect)
}

func assertParseVolumeTag(c *gc.C, tag string, expect names.VolumeTag) {
	t, err := names.ParseVolumeTag(tag)
	c.Assert(err, gc.IsNil)
	c.Assert(t, gc.Equals, expect)
}

func assertParseVolumeTagInvalid(c *gc.C, tag string, expect error) {
	_, err := names.ParseVolumeTag(tag)
	c.Assert(err, gc.ErrorMatches, expect.Error())
}
