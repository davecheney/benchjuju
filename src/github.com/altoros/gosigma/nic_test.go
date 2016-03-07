// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"testing"

	"github.com/altoros/gosigma/data"
)

func TestNIC_Empty(t *testing.T) {
	var n NIC
	n = nic{client: nil, obj: &data.NIC{}}
	if v := n.IPv4(); v != nil {
		t.Errorf("invalid NIC.IPv4 %v, must be nil", v)
	}
	if v := n.MAC(); v != "" {
		t.Errorf("invalid NIC.MAC %q, must be empty", v)
	}
	if v := n.Model(); v != "" {
		t.Errorf("invalid NIC.Model %q, must be empty", v)
	}
	if v := n.Runtime(); v != nil {
		t.Errorf("invalid NIC.Runtime %v, must be nil", v)
	}
	if v := n.VLAN(); v != nil {
		t.Errorf("invalid NIC.VLAN %v, must be nil", v)
	}

	const str = `{Model: "", MAC: "", IPv4: <nil>, VLAN: <nil>, Runtime: <nil>}`
	if v := n.String(); v != str {
		t.Errorf("invalid String() result: %q, must be %s", v, str)
	}
}

func TestNIC_DataIP(t *testing.T) {
	var d = data.NIC{
		IPv4: &data.IPv4{
			Conf: "static",
			IP:   data.MakeIPResource("31.171.246.37"),
		},
		Model: "virtio",
		MAC:   "22:40:85:4f:d3:ce",
	}
	var n = nic{obj: &d}

	if v := n.IPv4(); v == nil {
		t.Errorf("invalid NIC.IPv4, must be not nil")
	} else {
		if c := v.Conf(); c != "static" {
			t.Errorf("invalid NIC.IPv4.Conf %q, must be static", c)
		}
		if r := v.Resource(); r == nil {
			t.Error("invalid NIC.IPv4.Resource, must be not nil")
		} else {
			if uuid := r.UUID(); uuid != "31.171.246.37" {
				t.Errorf("invalid NIC.IPv4.Resource.UUID %q", uuid)
			}
			if uri := r.URI(); uri != "/api/2.0/ips/31.171.246.37/" {
				t.Errorf("invalid NIC.IPv4.Resource.URI %q", uri)
			}
		}
	}
	if v := n.Model(); v != "virtio" {
		t.Errorf("invalid model %q, must be virtio", v)
	}
	if v := n.MAC(); v != "22:40:85:4f:d3:ce" {
		t.Errorf("invalid MAC %q, must be 22:40:85:4f:d3:ce", v)
	}

	const str = `{Model: "virtio", MAC: "22:40:85:4f:d3:ce", IPv4: {Conf: "static", {URI: "/api/2.0/ips/31.171.246.37/", UUID: "31.171.246.37"}}, VLAN: <nil>, Runtime: <nil>}`
	if v := n.String(); v != str {
		t.Errorf("invalid String() result: %q, must be %s", v, str)
	}

	if r := n.Runtime(); r != nil {
		t.Errorf("invalid NIC.Runtime, must be nil, %v", r)
	}

	if v := n.VLAN(); v != nil {
		t.Errorf("invalid NIC.VLAN, must be nil, %v", v)
	}
}

func TestNIC_DataVLan(t *testing.T) {
	var d = data.NIC{
		Model: "virtio",
		MAC:   "22:40:85:4f:d3:ce",
		VLAN:  data.MakeVLanResource("5bc05e7e-6555-4f40-add8-3b8e91447702"),
	}
	var n = nic{obj: &d}

	if v := n.Model(); v != "virtio" {
		t.Errorf("invalid model %q, must be virtio", v)
	}
	if v := n.MAC(); v != "22:40:85:4f:d3:ce" {
		t.Errorf("invalid MAC %q, must be 22:40:85:4f:d3:ce", v)
	}

	const str = `{Model: "virtio", MAC: "22:40:85:4f:d3:ce", IPv4: <nil>, VLAN: {URI: "/api/2.0/vlans/5bc05e7e-6555-4f40-add8-3b8e91447702/", UUID: "5bc05e7e-6555-4f40-add8-3b8e91447702"}, Runtime: <nil>}`
	if v := n.String(); v != str {
		t.Errorf("invalid String() result: %q, must be %s", v, str)
	}

	if i := n.IPv4(); i != nil {
		t.Errorf("invalid NIC.IPv4, must be nil, %v", i)
	}
	if r := n.Runtime(); r != nil {
		t.Errorf("invalid NIC.Runtime, must be nil, %v", r)
	}
	if v := n.VLAN(); v == nil {
		t.Error("invalid NIC.VLAN, must be not nil")
	} else {
		if uuid := v.UUID(); uuid != "5bc05e7e-6555-4f40-add8-3b8e91447702" {
			t.Errorf("invalid NIC.VLAN.UUID %q", uuid)
		}
		if uri := v.URI(); uri != "/api/2.0/vlans/5bc05e7e-6555-4f40-add8-3b8e91447702/" {
			t.Errorf("invalid NIC.VLAN.URI %q", uri)
		}
	}
}

func TestNIC_Runtime(t *testing.T) {
	var n NIC
	n = nic{client: nil, obj: &data.NIC{
		Runtime: &data.RuntimeNetwork{},
	}}
	if v := n.Runtime(); v == nil {
		t.Errorf("invalid NIC.Runtime, must be not nil")
	}
}
