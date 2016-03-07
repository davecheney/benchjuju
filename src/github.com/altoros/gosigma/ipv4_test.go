// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"testing"

	"github.com/altoros/gosigma/data"
)

func TestIPv4String(t *testing.T) {
	i := &ipv4{obj: &data.IPv4{}}
	if s := i.String(); s != `{Conf: "", <nil>}` {
		t.Errorf("invalid IPv4.String(): `%s`", s)
	}

	i.obj.Conf = "conf"
	if s := i.String(); s != `{Conf: "conf", <nil>}` {
		t.Errorf("invalid IPv4.String(): `%s`", s)
	}

	i.obj.IP = data.MakeIPResource("0.1.2.3")
	if s := i.String(); s != `{Conf: "conf", {URI: "/api/2.0/ips/0.1.2.3/", UUID: "0.1.2.3"}}` {
		t.Errorf("invalid IPv4.String(): `%s`", s)
	}
}
