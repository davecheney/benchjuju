// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"flag"
	"testing"
)

var trace = flag.Bool("trace", false, "trace test requests/responses")

func TestVersionStringMatches(t *testing.T) {
	if vs, vns := Version(), VersionNumber().String(); vs != vns {
		t.Errorf("Version() != VersionNumber().String(): '%s' != '%s'", vs, vns)
	}
}
