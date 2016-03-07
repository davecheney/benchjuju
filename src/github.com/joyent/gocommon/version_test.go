//
// gocommon - Go library to interact with the JoyentCloud
//
//
// Copyright (c) 2013 Joyent Inc.
//
// Written by Daniele Stroppa <daniele.stroppa@joyent.com>
//

package gocommon

import (
	gc "launchpad.net/gocheck"
)

type VersionTestSuite struct {
}

var _ = gc.Suite(&VersionTestSuite{})

func (s *VersionTestSuite) TestStringMatches(c *gc.C) {
	c.Assert(Version, gc.Equals, VersionNumber.String())
}
