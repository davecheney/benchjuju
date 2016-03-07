// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"testing"

	"github.com/altoros/gosigma/data"
)

func TestRuntimeNIC_Empty(t *testing.T) {
	var n RuntimeNIC
	n = runtimeNIC{&data.RuntimeNetwork{}}
	if v := n.Type(); v != "" {
		t.Errorf("invalid RuntimeNIC.Type %q, must be empty", v)
	}
	if v := n.IPv4(); v != nil {
		t.Errorf("invalid RuntimeNIC.IPv4 %v, must be empty", v)
	}

	const str = `{Type: "", IPv4: <nil>}`
	if v := n.String(); v != str {
		t.Errorf("invalid String() result: %q, must be %s", v, str)
	}
}

func TestRuntimeNIC_Public(t *testing.T) {
	var obj = data.RuntimeNetwork{
		InterfaceType: "public",
		IPv4:          data.MakeIPResource("10.11.12.13"),
	}
	var n RuntimeNIC
	n = runtimeNIC{obj: &obj}
	if v := n.Type(); v != "public" {
		t.Errorf("invalid type %q, must be public", v)
	}
	if v := n.IPv4(); v.UUID() != "10.11.12.13" {
		t.Errorf("invalid address %q", v)
	}
}

func TestRuntimeNIC_Private(t *testing.T) {
	var obj = data.RuntimeNetwork{
		InterfaceType: "private",
	}
	var n RuntimeNIC
	n = runtimeNIC{obj: &obj}
	if v := n.Type(); v != "private" {
		t.Errorf("invalid type %q, must be private", v)
	}
	if v := n.IPv4(); v != nil {
		t.Errorf("invalid address %q, must be nil", v)
	}
}
