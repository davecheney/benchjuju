// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"testing"
	"time"

	"github.com/altoros/gosigma/data"
	"github.com/altoros/gosigma/mock"
)

var mockEndpoint string

func init() {
	mock.Start()
	mockEndpoint = mock.Endpoint("")
}

func newDataServer() *data.Server {
	return &data.Server{
		Resource: data.Resource{URI: "uri", UUID: "uuid"},
		Meta:     map[string]string{"key1": "value1", "key2": "value2"},
		Name:     "name",
		Status:   "status",
	}
}

func createTestClient(t *testing.T) (*Client, error) {
	cli, err := NewClient(mockEndpoint, mock.TestUser, mock.TestPassword, nil)
	if err != nil {
		return nil, err
	}

	if *trace {
		cli.Logger(t)
	}

	return cli, nil
}

type testLog struct{ written int }

func (l *testLog) Log(args ...interface{})                 { l.written++ }
func (l *testLog) Logf(format string, args ...interface{}) { l.written++ }

func TestClientLogger(t *testing.T) {
	cli, err := NewClient("https://0.1.2.3:2000/api/2.0/", mock.TestUser, mock.TestPassword, nil)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	var log testLog
	cli.Logger(&log)

	cli.ConnectTimeout(100 * time.Millisecond)
	cli.ReadWriteTimeout(100 * time.Millisecond)

	ssf, err := cli.Servers(false)
	if err == nil || ssf != nil {
		t.Error("Servers(false) returned valid result for unavailable endpoint")
		return
	}

	if log.written == 0 {
		t.Error("no writes to log")
	}
}

func TestClientTimeouts(t *testing.T) {
	cli, err := createTestClient(t)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	cli.ConnectTimeout(100 * time.Millisecond)
	if v := cli.GetConnectTimeout(); v != 100*time.Millisecond {
		t.Error("ConnectTimeout check failed")
	}

	cli.ReadWriteTimeout(200 * time.Millisecond)
	if v := cli.GetReadWriteTimeout(); v != 200*time.Millisecond {
		t.Error("ReadWriteTimeout check failed")
	}

	cli.OperationTimeout(300 * time.Millisecond)
	if v := cli.GetOperationTimeout(); v != 300*time.Millisecond {
		t.Error("OperationTimeout check failed")
	}
}

func TestClientEmptyUUID(t *testing.T) {
	cli, err := createTestClient(t)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	if _, err := cli.Server(""); err != errEmptyUUID {
		t.Error("Server('') must fail with errEmptyUUID")
	}
	if err := cli.StartServer("", nil); err != errEmptyUUID {
		t.Error("StartServer('') must fail with errEmptyUUID")
	}
	if err := cli.StopServer(""); err != errEmptyUUID {
		t.Error("StopServer('') must fail with errEmptyUUID")
	}
	if err := cli.RemoveServer("", RecurseAllDrives); err != errEmptyUUID {
		t.Error("RemoveServer('') must fail with errEmptyUUID")
	}
	if _, err := cli.Drive("", true); err != errEmptyUUID {
		t.Error("Drive('') must fail with errEmptyUUID")
	}
	if _, err := cli.Job(""); err != errEmptyUUID {
		t.Error("Job('') must fail with errEmptyUUID")
	}
	if _, err := cli.CloneDrive("", false, CloneParams{}, nil); err != errEmptyUUID {
		t.Error("CloneDrive('') must fail with errEmptyUUID")
	}
	if err := cli.RemoveDrive("", false); err != errEmptyUUID {
		t.Error("RemoveDrive('') must fail with errEmptyUUID")
	}
}

func TestClientEndpointUnavailableSoft(t *testing.T) {
	cli, err := NewClient(mockEndpoint+"1", mock.TestUser, mock.TestPassword, nil)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	ssf, err := cli.Servers(false)
	if err == nil || ssf != nil {
		t.Errorf("Servers(false) returned valid result for unavailable endpoint: %#v", ssf)
		return
	}
	t.Log("OK: Servers(false)", err)

	sst, err := cli.Servers(true)
	if err == nil || sst != nil {
		t.Errorf("Servers(true) returned valid result for unavailable endpoint: %#v", sst)
		return
	}
	t.Log("OK: Servers(true)", err)

	s, err := cli.Server("uuid")
	if err == nil {
		t.Errorf("Server() returned valid result with for unavailable endpoint: %#v", s)
		return
	}
	t.Log("OK, Server():", err)

	if _, err = cli.CreateServer(Components{}); err == nil {
		t.Error("CreateServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, CreateServer():", err)

	err = cli.StartServer("uuid", nil)
	if err == nil {
		t.Error("StartServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, StartServer():", err)

	err = cli.StopServer("uuid")
	if err == nil {
		t.Error("StopServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, StopServer():", err)

	err = cli.RemoveServer("uuid", RecurseAllDrives)
	if err == nil {
		t.Error("RemoveServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, RemoveServer():", err)

	dd, err := cli.Drives(false, false)
	if err == nil {
		t.Errorf("Drives(false) returned valid result for unavailable endpoint: %#v", dd)
		return
	}
	t.Log("OK, Drives(false):", err)

	d, err := cli.Drive("uuid", false)
	if err == nil {
		t.Errorf("Drive(false) returned valid result for unavailable endpoint: %#v", d)
		return
	}
	t.Log("OK, Drive(false):", err)

	cd, err := cli.CloneDrive("uuid", false, CloneParams{}, nil)
	if err == nil {
		t.Errorf("CloneDrive() returned valid result for unavailable endpoint: %#v", cd)
		return
	}

	err = cli.RemoveDrive("uuid", false)
	if err == nil {
		t.Error("RemoveDrive() returned valid result for unavailable endpoint")
		return
	}

	j, err := cli.Job("uuid")
	if err == nil {
		t.Errorf("Job() returned valid result for unavailable endpoint: %#v", j)
		return
	}
	t.Log("OK, Job():", err)
}

func TestClientEndpointUnavailableHard(t *testing.T) {
	cli, err := NewClient("https://0.1.2.3:2000/api/2.0/", mock.TestUser, mock.TestPassword, nil)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	cli.ConnectTimeout(100 * time.Millisecond)
	cli.ReadWriteTimeout(100 * time.Millisecond)

	ssf, err := cli.Servers(false)
	if err == nil || ssf != nil {
		t.Error("Servers(false) returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK: Servers(false)", err)

	sst, err := cli.Servers(true)
	if err == nil || sst != nil {
		t.Error("Servers(true) returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK: Servers(true)", err)

	s, err := cli.Server("uuid")
	if err == nil {
		t.Errorf("Server() returned valid result for unavailable endpoint: %#v", s)
		return
	}
	t.Log("OK, Server():", err)

	if _, err = cli.CreateServer(Components{}); err == nil {
		t.Error("CreateServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, CreateServer():", err)

	err = cli.StartServer("uuid", nil)
	if err == nil {
		t.Error("StartServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, StartServer():", err)

	err = cli.StopServer("uuid")
	if err == nil {
		t.Error("StopServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, StopServer():", err)

	err = cli.RemoveServer("uuid", RecurseAllDrives)
	if err == nil {
		t.Error("RemoveServer() returned valid result for unavailable endpoint")
		return
	}
	t.Log("OK, RemoveServer():", err)

	dd, err := cli.Drives(false, false)
	if err == nil {
		t.Errorf("Drives(false) returned valid result for unavailable endpoint: %#v", dd)
		return
	}
	t.Log("OK, Drives(false):", err)

	d, err := cli.Drive("uuid", false)
	if err == nil {
		t.Errorf("Drive(false) returned valid result for unavailable endpoint: %#v", d)
		return
	}
	t.Log("OK, Drive(false):", err)

	cd, err := cli.CloneDrive("uuid", false, CloneParams{}, nil)
	if err == nil {
		t.Errorf("CloneDrive() returned valid result for unavailable endpoint: %#v", cd)
		return
	}

	err = cli.RemoveDrive("uuid", false)
	if err == nil {
		t.Error("RemoveDrive() returned valid result for unavailable endpoint")
		return
	}

	j, err := cli.Job("uuid")
	if err == nil {
		t.Errorf("Job() returned valid result for unavailable endpoint: %#v", j)
		return
	}
	t.Log("OK, Job():", err)
}
