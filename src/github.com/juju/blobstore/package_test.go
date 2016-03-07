// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package blobstore_test

import (
	"testing"

	gitjujutesting "github.com/juju/testing"
)

func Test(t *testing.T) {
	gitjujutesting.MgoTestPackage(t, nil)
}
