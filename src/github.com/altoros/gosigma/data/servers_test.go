// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestDataServersReaderFail(t *testing.T) {
	r := failReader{}

	if _, err := ReadServer(r); err == nil || err.Error() != "test error" {
		t.Error("Fail")
	}

	if _, err := ReadServers(r); err == nil || err.Error() != "test error" {
		t.Error("Fail")
	}
}

func TestDataServersUnmarshal(t *testing.T) {
	var ss Servers
	ss.Meta.Limit = 12345
	ss.Meta.Offset = 12345
	ss.Meta.TotalCount = 12345
	err := json.Unmarshal([]byte(jsonServersData), &ss)
	if err != nil {
		t.Error(err)
	}

	verifyMeta(t, &ss.Meta, 0, 0, 5)

	for i := 0; i < len(serversData); i++ {
		compareServers(t, i, &ss.Objects[i], &serversData[i])
	}
}

func TestDataServersReadServers(t *testing.T) {
	ss, err := ReadServers(strings.NewReader(jsonServersData))
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(serversData); i++ {
		compareServers(t, i, &ss[i], &serversData[i])
	}
}

func TestDataServersDetailUnmarshal(t *testing.T) {
	var ss Servers
	ss.Meta.Limit = 12345
	ss.Meta.Offset = 12345
	ss.Meta.TotalCount = 12345
	err := json.Unmarshal([]byte(jsonServersDetailData), &ss)
	if err != nil {
		t.Error(err)
	}

	verifyMeta(t, &ss.Meta, 0, 0, 5)

	for i := 0; i < len(serversDetailData); i++ {
		compareServers(t, i, &ss.Objects[i], &serversDetailData[i])
	}
}

func TestDataServersReadServersDetail(t *testing.T) {
	ss, err := ReadServers(strings.NewReader(jsonServersDetailData))
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(serversDetailData); i++ {
		compareServers(t, i, &ss[i], &serversDetailData[i])
	}
}

func TestDataServersReadServerDetail(t *testing.T) {
	s, err := ReadServer(strings.NewReader(jsonServerData))
	if err != nil {
		t.Error(err)
	}
	compareServers(t, 0, s, &serverData)
}

func compareNICs(t *testing.T, i int, value, wants *NIC) {
	if value.IPv4 != nil && wants.IPv4 != nil {
		if value.IPv4.Conf != wants.IPv4.Conf {
			t.Errorf("NIC.IPv4.Conf error [%d]: found %#v, wants %#v", i, value.IPv4.Conf, wants.IPv4.Conf)
		}
		if value.IPv4.IP != nil && wants.IPv4.IP != nil {
			if *value.IPv4.IP != *wants.IPv4.IP {
				t.Errorf("NIC.IPv4.IP error [%d]: found %#v, wants %#v", i, value.IPv4.IP, wants.IPv4.IP)
			}
		} else if value.IPv4.IP != nil || wants.IPv4.IP != nil {
			t.Errorf("NIC.IPv4.IP error [%d]: found %#v, wants %#v", i, value.IPv4.IP, wants.IPv4.IP)
		}
	} else if value.IPv4 != nil || wants.IPv4 != nil {
		t.Errorf("NIC.IPv4 error [%d]: found %#v, wants %#v", i, value.IPv4, wants.IPv4)
	}

	if value.Model != wants.Model {
		t.Errorf("NIC.Model error [%d]: found %#v, wants %#v", i, value.Model, wants.Model)
	}
	if value.MAC != wants.MAC {
		t.Errorf("NIC.MAC error [%d]: found %#v, wants %#v", i, value.MAC, wants.MAC)
	}

	if value.VLAN != nil && wants.VLAN != nil {
		if *value.VLAN != *wants.VLAN {
			t.Errorf("NIC.VLAN error [%d]: found %#v, wants %#v", i, value.VLAN, wants.VLAN)
		}
	} else if value.VLAN != nil || wants.VLAN != nil {
		t.Errorf("NIC.VLAN error [%d]: found %#v, wants %#v", i, value.VLAN, wants.VLAN)
	}
}

func compareServers(t *testing.T, i int, value, wants *Server) {
	if value.Resource != wants.Resource {
		t.Errorf("Resource error [%d]: found %#v, wants %#v", i, value.Resource, wants.Resource)
	}

	if value.Context != wants.Context {
		t.Errorf("Server.Context error [%d]: found %#v, wants %#v", i, value.Context, wants.Context)
	}
	if value.CPU != wants.CPU {
		t.Errorf("Server.CPU error [%d]: found %#v, wants %#v", i, value.CPU, wants.CPU)
	}
	if value.CPUsInsteadOfCores != wants.CPUsInsteadOfCores {
		t.Errorf("Server.CPUsInsteadOfCores error [%d]: found %#v, wants %#v", i, value.CPUsInsteadOfCores, wants.CPUsInsteadOfCores)
	}
	if value.CPUModel != wants.CPUModel {
		t.Errorf("Server.CPUModel error [%d]: found %#v, wants %#v", i, value.CPUModel, wants.CPUModel)
	}

	if len(value.Drives) != len(wants.Drives) {
		t.Errorf("Server.Drives error [%d]: found %#v, wants %#v", i, value.Drives, wants.Drives)
	}
	for i := 0; i < len(value.Drives); i++ {
		if value.Drives[i] != wants.Drives[i] {
			t.Errorf("Server.Drives error [%d]: found %#v, wants %#v", i, value.Drives[i], wants.Drives[i])
		}
	}

	if value.Mem != wants.Mem {
		t.Errorf("Server.Mem error [%d]: found %#v, wants %#v", i, value.Mem, wants.Mem)
	}

	compareMeta(t, fmt.Sprintf("Server.Meta error [%d]", i), value.Meta, wants.Meta)

	if value.Name != wants.Name {
		t.Errorf("Name error [%d]: found %#v, wants %#v", i, value.Name, wants.Name)
	}

	if len(value.NICs) != len(wants.NICs) {
		t.Errorf("Server.NICs error [%d]: found %#v, wants %#v", i, value.NICs, wants.NICs)
	}
	for i := 0; i < len(value.NICs); i++ {
		compareNICs(t, i, &value.NICs[i], &wants.NICs[i])
	}

	if value.SMP != wants.SMP {
		t.Errorf("Server.SMP error [%d]: found %#v, wants %#v", i, value.SMP, wants.SMP)
	}

	if value.Status != wants.Status {
		t.Errorf("Status error [%d]: found %#v, wants %#v", i, value.Status, wants.Status)
	}

	if value.VNCPassword != wants.VNCPassword {
		t.Errorf("Server.VNCPassword error [%d]: found %#v, wants %#v", i, value.VNCPassword, wants.VNCPassword)
	}
}
