// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"testing"

	"github.com/altoros/gosigma/data"
	"github.com/altoros/gosigma/mock"
)

func TestClientServersEmpty(t *testing.T) {
	mock.ResetServers()

	cli, err := createTestClient(t)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	check := func(rqspec RequestSpec) {
		servers, err := cli.Servers(rqspec)
		if err != nil {
			t.Error(err)
		}
		if len(servers) > 0 {
			t.Errorf("%v", servers)
		}
	}

	check(RequestShort)
	check(RequestDetail)
}

func TestClientServers(t *testing.T) {
	mock.ResetServers()

	ds := newDataServer()
	mock.AddServer(ds)

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	servers, err := cli.Servers(true)
	if err != nil {
		t.Error(err)
		return
	}

	if len(servers) != 1 {
		t.Errorf("invalid len: %v", servers)
		return
	}

	s := servers[0]

	if s.String() == "" {
		t.Error("Empty string representation")
		return
	}

	checkv := func(v, wants string) {
		if v != wants {
			t.Errorf("value %s, wants %s", v, wants)
		}
	}
	checkv(s.Name(), "name")
	checkv(s.URI(), "uri")
	checkv(s.Status(), "status")
	checkv(s.UUID(), "uuid")

	checkg := func(s Server, k, wants string) {
		if v, ok := s.Get(k); !ok || v != wants {
			t.Errorf("value of Get(%q) = %q, %v; wants %s", k, v, ok, wants)
		}
	}
	checkg(s, "key1", "value1")
	checkg(s, "key2", "value2")

	// refresh
	ds.Name = "name1"
	ds.URI = "uri1"
	ds.Status = "status1"
	ds.Meta["key1"] = "value11"
	ds.Meta["key2"] = "value22"
	ds.Meta["key3"] = "value33"
	ds.Context = true
	ds.CPU = 100
	ds.CPUsInsteadOfCores = true
	ds.CPUModel = "cpu_model"
	if err := s.Refresh(); err != nil {
		t.Error(err)
		return
	}
	checkv(s.Name(), "name1")
	checkv(s.URI(), "uri1")
	checkv(s.Status(), "status1")
	checkg(s, "key1", "value11")
	checkg(s, "key2", "value22")
	checkg(s, "key3", "value33")
	if v := s.Context(); v != true {
		t.Errorf("Server.Context() failed")
	}
	if v := s.CPU(); v != 100 {
		t.Errorf("Server.CPU() failed")
	}
	if v := s.CPUsInsteadOfCores(); v != true {
		t.Errorf("Server.CPUsInsteadOfCores() failed")
	}
	if v := s.CPUModel(); v != "cpu_model" {
		t.Errorf("Server.CPUModel() failed")
	}

	// failed refresh
	mock.ResetServers()
	if err := s.Refresh(); err == nil {
		t.Error("Server refresh must fail")
		return
	}

	mock.ResetServers()
}

func TestClientServer(t *testing.T) {
	mock.ResetServers()

	ds := newDataServer()
	mock.AddServer(ds)

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	if s, err := cli.Server(""); err == nil {
		t.Errorf("Server() returned valid result for empty uuid: %#v", s)
		return
	}

	s, err := cli.Server("uuid")
	if err != nil {
		t.Error(err)
		return
	}

	if s.String() == "" {
		t.Error("Empty string representation")
	}

	checkv := func(v, wants string) {
		if v != wants {
			t.Errorf("value %s, wants %s", v, wants)
		}
	}
	checkv(s.Name(), "name")
	checkv(s.URI(), "uri")
	checkv(s.Status(), "status")
	checkv(s.UUID(), "uuid")

	checkg := func(s Server, k, wants string) {
		if v, ok := s.Get(k); !ok || v != wants {
			t.Errorf("value of Get(%q) = %q, %v; wants %s", k, v, ok, wants)
		}
	}
	checkg(s, "key1", "value1")
	checkg(s, "key2", "value2")

	// refresh
	ds.Name = "name1"
	ds.URI = "uri1"
	ds.Status = "status1"
	ds.Meta["key1"] = "value11"
	ds.Meta["key2"] = "value22"
	ds.Meta["key3"] = "value33"
	if err := s.Refresh(); err != nil {
		t.Error(err)
	}
	checkv(s.Name(), "name1")
	checkv(s.URI(), "uri1")
	checkv(s.Status(), "status1")
	checkg(s, "key1", "value11")
	checkg(s, "key2", "value22")
	checkg(s, "key3", "value33")

	// failed refresh
	mock.ResetServers()
	if err := s.Refresh(); err == nil {
		t.Error("Server refresh must fail")
	}
}

func TestClientServerNotFound(t *testing.T) {
	mock.ResetServers()

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	s, err := cli.Server("uuid1234567")
	if err == nil {
		t.Errorf("found server %#v", s)
	}

	t.Log(err)
	cs, ok := err.(*Error)
	if !ok {
		t.Error("error required to be gosigma.Error")
	}
	if cs.ServiceError.Message != "notfound" {
		t.Error("invalid error message from mock server")
	}
}

func TestClientServersShort(t *testing.T) {
	mock.ResetServers()

	ds := newDataServer()
	mock.AddServer(ds)

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	ss, err := cli.Servers(false)
	if err != nil {
		t.Error(err)
		return
	}

	if v, ok := ss[0].Get("key1"); ok || len(v) > 0 {
		t.Error("Error getting short server list")
	}

	ss, err = cli.Servers(true)
	if err != nil {
		t.Error(err)
	}

	if v, ok := ss[0].Get("key1"); !ok || len(v) == 0 {
		t.Error("Error getting detailed server list")
	}

	mock.ResetServers()
}

func TestClientStartServerInvalidUUID(t *testing.T) {
	mock.ResetServers()

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	// No Server
	if err := cli.StartServer("uuid-123", nil); err == nil {
		t.Error("Start server must fail here")
	} else {
		t.Log("Start server:", err)
	}

	// No Server with empty non-nil avoid
	if err := cli.StartServer("uuid-123", []string{}); err == nil {
		t.Error("Start server must fail here")
	} else {
		t.Log("Start server:", err)
	}

	// No Server with non-empty non-nil avoid
	if err := cli.StartServer("uuid-123", []string{"non-uuid"}); err == nil {
		t.Error("Start server must fail here")
	} else {
		t.Log("Start server:", err)
	}
}

func TestClientStartServer(t *testing.T) {
	mock.ResetServers()

	ds := newDataServer()
	ds.Status = "stopped"
	mock.AddServer(ds)

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	s, err := cli.Server("uuid")
	if err != nil {
		t.Error(err)
		return
	}

	if err := s.Start(); err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10 && s.Status() != ServerRunning; i++ {
		if err := s.Refresh(); err != nil {
			t.Error(err)
			return
		}
	}

	if s.Status() != ServerRunning {
		t.Error("Server status must be running")
	}
}

func TestClientStopServer(t *testing.T) {
	mock.ResetServers()

	ds := newDataServer()
	ds.Status = "running"
	mock.AddServer(ds)

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	s, err := cli.Server("uuid")
	if err != nil {
		t.Error(err)
		return
	}

	if err := s.Stop(); err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10 && s.Status() != ServerStopped; i++ {
		if err := s.Refresh(); err != nil {
			t.Error(err)
			return
		}
	}

	if s.Status() != ServerStopped {
		t.Error("Server status must be stopped")
	}
}

func TestClientCreateServer(t *testing.T) {
	mock.ResetServers()

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	var c Components
	c.SetName("test")
	c.SetCPU(2000)
	c.SetMem(2147483648)
	c.SetSMP(20)
	c.SetVNCPassword("testserver")
	c.NetworkDHCP4("virtio")
	c.NetworkManual4("virtio")
	c.NetworkStatic4("virtio", "ipaddr")
	c.NetworkVLan("virtio", "vlanid")
	c.AttachDrive(1, "0:0", "virtio", "uuid")

	s, err := cli.CreateServer(c)
	if err != nil {
		t.Error(err)
		return
	}

	if s.Name() != "test" {
		t.Error("Invalid name")
	}
	if s.CPU() != 2000 {
		t.Error("Invalid cpu")
	}
	if s.Mem() != 2147483648 {
		t.Error("Invalid mem")
	}
	if s.SMP() != 20 {
		t.Error("Invalid mem")
	}
	if s.Status() != ServerStopped {
		t.Error("Server status must be stopped")
	}
	if s.VNCPassword() != "testserver" {
		t.Error("VNCPassword invalid")
	}
	if v := s.IPv4(); len(v) != 0 {
		t.Error("IPv4 invalid:", v)
	}

	nics := s.NICs()
	if len(nics) != 4 {
		t.Errorf("NICs error: %#v", nics)
	}

	n := nics[0]
	if c := n.IPv4().Conf(); c != "dhcp" {
		t.Errorf("NIC.Conf [0]: %q", c)
	}
	if n.Model() != "virtio" {
		t.Errorf("NIC.Model [0]: %q", n.Model())
	}
	if n.MAC() != "" {
		t.Errorf("NIC.MAC [0]: %q", n.MAC())
	}
	if v := n.VLAN(); v != nil {
		t.Errorf("NIC.VLAN [0] must be nil, %v", v)
	}

	n = nics[1]
	if c := n.IPv4().Conf(); c != "manual" {
		t.Errorf("NIC.Conf [1]: %q", c)
	}
	if n.Model() != "virtio" {
		t.Errorf("NIC.Model [1]: %q", n.Model())
	}
	if n.MAC() != "" {
		t.Errorf("NIC.MAC [1]: %q", n.MAC())
	}
	if v := n.VLAN(); v != nil {
		t.Errorf("NIC.VLAN [1] must be nil, %v", v)
	}

	n = nics[2]
	if c := n.IPv4().Conf(); c != "static" {
		t.Errorf("NIC.Conf [2]: %q", c)
	}
	if n.Model() != "virtio" {
		t.Errorf("NIC.Model [2]: %q", n.Model())
	}
	if n.MAC() != "" {
		t.Errorf("NIC.MAC [2]: %q", n.MAC())
	}
	if v := n.VLAN(); v != nil {
		t.Errorf("NIC.VLAN [2] must be nil, %v", v)
	}

	n = nics[3]
	if v := n.IPv4(); v != nil {
		t.Errorf("NIC.IPV4 [3]: %q", v)
	}
	if n.Model() != "virtio" {
		t.Errorf("NIC.Model [3]: %q", n.Model())
	}
	if n.MAC() != "" {
		t.Errorf("NIC.MAC [3]: %q", n.MAC())
	}
	if v := n.VLAN(); v == nil {
		t.Error("NIC.VLAN [3] must be not nil")
	} else if vv := v.UUID(); vv != "vlanid" {
		t.Errorf("NIC.VLAN [3]: %q", vv)
	}

	drives := s.Drives()
	if len(drives) != 1 {
		t.Errorf("Drives error: %#v", drives)
	}

	dd := drives[0]
	if v := dd.BootOrder(); v != 1 {
		t.Errorf("ServerDrive.BootOrder: %#v", v)
	}
	if v := dd.Channel(); v != "0:0" {
		t.Errorf("ServerDrive.BootOrder: %#v", v)
	}
	if v := dd.Device(); v != "virtio" {
		t.Errorf("ServerDrive.Device: %#v", v)
	}
	if v := dd.UUID(); v != "uuid" {
		t.Errorf("ServerDrive.UUID: %#v", v)
	}
	if v := dd.URI(); v != "/api/2.0/drives/uuid/" {
		t.Errorf("ServerDrive.URI: %#v", v)
	}
	if v := dd.String(); v != `{BootOrder: 1, Channel: "0:0", Device: "virtio", UUID: "uuid"}` {
		t.Errorf("ServerDrive.String: %#v", v)
	}

	ddd := dd.Drive()
	if v := ddd.UUID(); v != "uuid" {
		t.Errorf("Drive.UUID: %#v", v)
	}
	if v := ddd.URI(); v != data.MakeDriveResource("uuid").URI {
		t.Errorf("Drive.URI: %#v", v)
	}
	if v := ddd.Name(); v != "" {
		t.Errorf("Drive.Name: %#v", v)
	}
	if v := ddd.Status(); v != "" {
		t.Errorf("Drive.Status: %#v", v)
	}
	if v := ddd.Media(); v != "" {
		t.Errorf("Drive.Media: %#v", v)
	}
	if v := ddd.StorageType(); v != "" {
		t.Errorf("Drive.StorageType: %#v", v)
	}
	if v := ddd.Size(); v != 0 {
		t.Errorf("Drive.Size: %#v", v)
	}
}

func TestServerIPv4(t *testing.T) {
	s := &server{obj: &data.Server{
		NICs: []data.NIC{
			data.NIC{},
			data.NIC{Runtime: &data.RuntimeNetwork{}},
		},
	}}

	if ips := s.IPv4(); len(ips) != 0 {
		t.Errorf("invalid Server.IPv4(): %v", ips)
	}

	nic0 := data.NIC{Runtime: &data.RuntimeNetwork{
		IPv4: data.MakeIPResource("0.1.2.3"),
	}}
	s.obj.NICs = append(s.obj.NICs, nic0)

	if ips := s.IPv4(); len(ips) != 1 || ips[0] != "0.1.2.3" {
		t.Errorf("invalid Server.IPv4(): %v", ips)
	}

	nic1 := data.NIC{Runtime: &data.RuntimeNetwork{
		IPv4: data.MakeIPResource("0.2.3.4"),
	}}
	s.obj.NICs = append(s.obj.NICs, nic1)

	if ips := s.IPv4(); len(ips) != 2 || ips[0] != "0.1.2.3" || ips[1] != "0.2.3.4" {
		t.Errorf("invalid Server.IPv4(): %v", ips)
	}
}
