//
// gosign - Go HTTP signing library for the Joyent Public Cloud and Joyent Manta
//
//
// Copyright (c) 2013 Joyent Inc.
//
// Written by Daniele Stroppa <daniele.stroppa@joyent.com>
//

package gosign

import (
	gc "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type GoSignTestSuite struct {
}

var _ = gc.Suite(&GoSignTestSuite{})
