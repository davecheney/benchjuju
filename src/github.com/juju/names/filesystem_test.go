// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	"fmt"

	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type filesystemSuite struct{}

var _ = gc.Suite(&filesystemSuite{})

func (s *filesystemSuite) TestFilesystemTag(c *gc.C) {
	c.Assert(names.NewFilesystemTag("1").String(), gc.Equals, "filesystem-1")
}

func (s *filesystemSuite) TestFilesystemIdValidity(c *gc.C) {
	assertFilesystemIdValid(c, "0")
	assertFilesystemIdValid(c, "0/lxc/0/0")
	assertFilesystemIdValid(c, "1000")
	assertFilesystemIdInvalid(c, "-1")
	assertFilesystemIdInvalid(c, "")
	assertFilesystemIdInvalid(c, "one")
	assertFilesystemIdInvalid(c, "#")
	assertFilesystemIdInvalid(c, "0/0/0") // 0/0 is not a valid machine ID
}

func (s *filesystemSuite) TestParseFilesystemTag(c *gc.C) {
	assertParseFilesystemTag(c, "filesystem-0", names.NewFilesystemTag("0"))
	assertParseFilesystemTag(c, "filesystem-88", names.NewFilesystemTag("88"))
	assertParseFilesystemTag(c, "filesystem-0-lxc-0-88", names.NewFilesystemTag("0/lxc/0/88"))
	assertParseFilesystemTagInvalid(c, "", names.InvalidTagError("", ""))
	assertParseFilesystemTagInvalid(c, "one", names.InvalidTagError("one", ""))
	assertParseFilesystemTagInvalid(c, "filesystem-", names.InvalidTagError("filesystem-", names.FilesystemTagKind))
	assertParseFilesystemTagInvalid(c, "machine-0", names.InvalidTagError("machine-0", names.FilesystemTagKind))
}

func (s *filesystemSuite) TestFilesystemMachine(c *gc.C) {
	assertFilesystemMachine(c, "0/0", names.NewMachineTag("0"))
	assertFilesystemMachine(c, "0/lxc/0/0", names.NewMachineTag("0/lxc/0"))
	assertFilesystemNoMachine(c, "0")
}

func assertFilesystemMachine(c *gc.C, id string, expect names.MachineTag) {
	t, ok := names.FilesystemMachine(names.NewFilesystemTag(id))
	c.Assert(ok, gc.Equals, true)
	c.Assert(t, gc.Equals, expect)
}

func assertFilesystemNoMachine(c *gc.C, id string) {
	_, ok := names.FilesystemMachine(names.NewFilesystemTag(id))
	c.Assert(ok, gc.Equals, false)
}

func assertFilesystemIdValid(c *gc.C, name string) {
	c.Assert(names.IsValidFilesystem(name), gc.Equals, true)
	names.NewFilesystemTag(name)
}

func assertFilesystemIdInvalid(c *gc.C, name string) {
	c.Assert(names.IsValidFilesystem(name), gc.Equals, false)
	testFilesystemTag := func() { names.NewFilesystemTag(name) }
	expect := fmt.Sprintf("%q is not a valid filesystem id", name)
	c.Assert(testFilesystemTag, gc.PanicMatches, expect)
}

func assertParseFilesystemTag(c *gc.C, tag string, expect names.FilesystemTag) {
	t, err := names.ParseFilesystemTag(tag)
	c.Assert(err, gc.IsNil)
	c.Assert(t, gc.Equals, expect)
}

func assertParseFilesystemTagInvalid(c *gc.C, tag string, expect error) {
	_, err := names.ParseFilesystemTag(tag)
	c.Assert(err, gc.ErrorMatches, expect.Error())
}
