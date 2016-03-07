// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import "testing"

func testMarshalComponents(t *testing.T, c Components, test, wants string) bool {
	s, err := c.marshalString()
	if err != nil {
		t.Error(t)
		return false
	}
	if s != wants {
		t.Errorf("invalid %s, s=%v, wants=%v", test, s, wants)
		return false
	}
	return true
}

func TestComponentsNil(t *testing.T) {
	var c Components
	testMarshalComponents(t, c, `marshal empty struct`, `{}`)
}

func TestComponentsName(t *testing.T) {
	var c Components

	c.SetName("test")
	testMarshalComponents(t, c, `SetName("test")`, `{"name":"test"}`)

	c.SetName("")
	testMarshalComponents(t, c, `SetName("")`, `{}`)
}

func TestComponentsCPU(t *testing.T) {
	var c Components

	c.SetCPU(2000)
	testMarshalComponents(t, c, `SetCPU(2000)`, `{"cpu":2000}`)

	c.SetCPU(0)
	testMarshalComponents(t, c, `SetCPU(0)`, `{}`)
}

func TestComponentsMem(t *testing.T) {
	var c Components

	c.SetMem(5368709120)
	testMarshalComponents(t, c, `SetMem(5368709120)`, `{"mem":5368709120}`)

	c.SetMem(0)
	testMarshalComponents(t, c, `SetMem(0)`, `{}`)
}

func TestComponentsVNCPassword(t *testing.T) {
	var c Components

	c.SetVNCPassword("password")
	testMarshalComponents(t, c, `SetVNCPassword("password")`, `{"vnc_password":"password"}`)

	c.SetVNCPassword("")
	testMarshalComponents(t, c, `SetVNCPassword("")`, `{}`)
}

func TestComponentsDescription(t *testing.T) {
	var c Components

	c.SetDescription("description")
	testMarshalComponents(t, c, `SetDescription("description")`, `{"meta":{"description":"description"}}`)

	c.SetDescription("")
	testMarshalComponents(t, c, `SetDescription("")`, `{}`)
}

func TestComponentsSSHPublicKey(t *testing.T) {
	var c Components

	c.SetSSHPublicKey("key")
	testMarshalComponents(t, c, `SetSSHPublicKey("key")`, `{"meta":{"ssh_public_key":"key"}}`)

	c.SetSSHPublicKey("")
	testMarshalComponents(t, c, `SetSSHPublicKey("")`, `{}`)
}

func TestComponentsAttachDrive(t *testing.T) {
	var c Components
	c.AttachDrive(1, "0:0", "virtio", "uuid")
	testMarshalComponents(t, c, "AttachDrive",
		`{"drives":[{"boot_order":1,"dev_channel":"0:0","device":"virtio","drive":{"resource_uri":"/api/2.0/drives/uuid/","uuid":"uuid"}}]}`)
}

func TestComponentsNetworkDHCP4(t *testing.T) {
	var c Components
	c.NetworkDHCP4("virtio")
	testMarshalComponents(t, c, `NetworkDHCP4("virtio")`,
		`{"nics":[{"ip_v4_conf":{"conf":"dhcp"},"model":"virtio"}]}`)
}

func TestComponentsNetworkStatic4(t *testing.T) {
	var c Components
	c.NetworkStatic4("virtio", "ipaddr")
	testMarshalComponents(t, c, `NetworkStatic4("virtio", "ipaddr")`,
		`{"nics":[{"ip_v4_conf":{"conf":"static","ip":{"resource_uri":"/api/2.0/ips/ipaddr/","uuid":"ipaddr"}},"model":"virtio"}]}`)
}

func TestComponentsNetworkManual4(t *testing.T) {
	var c Components
	c.NetworkManual4("virtio")
	testMarshalComponents(t, c, `NetworkManual4("virtio")`,
		`{"nics":[{"ip_v4_conf":{"conf":"manual"},"model":"virtio"}]}`)
}

func TestComponentsNetworkVLan(t *testing.T) {
	var c Components
	c.NetworkVLan("virtio", "vlanuuid")
	testMarshalComponents(t, c, `NetworkVLan("virtio", "vlanuuid")`,
		`{"nics":[{"model":"virtio","vlan":{"resource_uri":"/api/2.0/vlans/vlanuuid/","uuid":"vlanuuid"}}]}`)
}
