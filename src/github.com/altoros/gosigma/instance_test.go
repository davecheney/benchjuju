// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"flag"
	"testing"
	"time"
)

var instance = flag.Bool("instance", false, "run instance tests, should be run inside CloudSigma server instance")

func TestReadContext(t *testing.T) {
	if !*instance {
		t.SkipNow()
		return
	}

	client, err := NewClient("zrh", "user", "password", nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		client.Logger(t)
	}

	client.ReadWriteTimeout(2 * time.Second)

	ctx, err := client.ReadContext()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(ctx)

	nics := ctx.NICs()
	for i, n := range nics {
		t.Logf("nic #%d: %s", i, n)
	}
}
