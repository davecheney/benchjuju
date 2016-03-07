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
	"testing"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type GoCommonTestSuite struct {
}

var _ = gc.Suite(&GoCommonTestSuite{})
