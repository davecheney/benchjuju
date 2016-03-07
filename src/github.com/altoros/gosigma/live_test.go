// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"errors"
	"flag"
	"strings"
	"testing"
	"time"
)

var live = flag.String("live", "", "run live tests against CloudSigma endpoint, specify credentials in form -live=user:pass")
var suid = flag.String("suid", "", "uuid of server at CloudSigma to run server specific tests")
var duid = flag.String("duid", "", "uuid of drive at CloudSigma to run drive specific tests")
var vlan = flag.String("vlan", "", "uuid of vlan at CloudSigma to run server specific tests")
var sshkey = flag.String("sshkey", "", "public ssh key to run server specific tests")
var force = flag.Bool("force", false, "force start/stop live tests")
var lib = flag.Bool("lib", false, "duid is library drive")
var size = flag.Uint64("size", 0, "size for operations: TestLiveDriveResize")

func libFlag() LibrarySpec {
	if *lib {
		return LibraryMedia
	}
	return LibraryAccount
}

func parseCredentials() (u string, p string, e error) {
	if *live == "" {
		return
	}

	parts := strings.SplitN(*live, ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		e = errors.New("Invalid credentials: " + *live)
		return
	}

	u, p = parts[0], parts[1]

	return
}

func skipTest(t *testing.T, e error) {
	if e == nil {
		t.SkipNow()
	} else {
		t.Error(e)
	}
}

func TestLiveServers(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	ii, err := cli.Servers(false)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%v", ii)
}

func TestLiveServerGet(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *suid == "" {
		t.Skip("-suid=<server-uuid> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	s, err := cli.Server(*suid)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%v", s)
}

func TestLiveServerStart(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *suid == "" {
		t.Skip("-suid=<server-uuid> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	s, err := cli.Server(*suid)
	if err != nil {
		t.Error(err)
		return
	}

	if s.Status() != ServerStopped && !*force {
		t.Skip("wrong server status", s.Status())
		return
	}

	if err := s.Start(); err != nil {
		t.Error(err)
	}
}

func TestLiveServerStartWait(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *suid == "" {
		t.Skip("-suid=<server-uuid> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	s, err := cli.Server(*suid)
	if err != nil {
		t.Error(err)
		return
	}

	if s.Status() != ServerStopped && !*force {
		t.Skip("wrong server status", s.Status())
		return
	}

	if err := s.StartWait(); err != nil {
		t.Error(err)
	}
}

func TestLiveServerStop(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *suid == "" {
		t.Skip("-suid=<server-uuid> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	s, err := cli.Server(*suid)
	if err != nil {
		t.Error(err)
		return
	}

	if s.Status() != ServerRunning && !*force {
		t.Skip("wrong server status", s.Status())
		return
	}

	if err := s.Stop(); err != nil {
		t.Error(err)
	}
}

func TestLiveDriveGet(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *duid == "" {
		t.Skip("-duid=<drive-uuid> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	d, err := cli.Drive(*duid, libFlag())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%v", d)
}

func TestLiveDriveList(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	dd, err := cli.Drives(true, libFlag())
	if err != nil {
		t.Error(err)
		return
	}

	for _, d := range dd {
		t.Logf("%v", d)
	}
}

func TestLiveDriveClone(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *duid == "" {
		t.Skip("-duid=<drive-uuid> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	d, err := cli.Drive(*duid, libFlag())
	if err != nil {
		t.Error(err)
		return
	}

	var cloneParams CloneParams
	cloneParams.Name = "LiveTest-" + time.Now().Format("15-04-05-999999999")
	newDrive, err := d.CloneWait(cloneParams, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%v", newDrive)
}

func TestLiveServerClone(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *duid == "" {
		t.Skip("-duid=<drive-uuid> must be specified")
		return
	}

	if *vlan == "" {
		t.Skip("-vlan=<vlan-uuid> must be specified")
		return
	}

	if *sshkey == "" {
		t.Skip("-sshkey=<ssh-public-key> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	originalDrive, err := cli.Drive(*duid, libFlag())
	if err != nil {
		t.Error(err)
		return
	}

	stamp := time.Now().Format("15-04-05-999999999")

	var cloneParams CloneParams
	cloneParams.Name = "LiveTest-drv-" + stamp
	newDrive, err := originalDrive.CloneWait(cloneParams, nil)
	if err != nil {
		t.Error(err)
		return
	}

	var c Components
	c.SetName("LiveTest-srv-" + stamp)
	c.SetCPU(2000)
	c.SetMem(2 * Gigabyte)
	c.SetVNCPassword("test-vnc-password")
	c.SetSSHPublicKey(*sshkey)
	c.SetDescription("test-description")
	c.AttachDrive(1, "0:0", "virtio", newDrive.UUID())
	c.NetworkDHCP4(ModelVirtio)
	c.NetworkVLan(ModelVirtio, *vlan)

	s, err := cli.CreateServer(c)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%v", s)
}

func TestLiveServerRemove(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *suid == "" {
		t.Skip("-suid=<server-uuid> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error("create client", err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	s, err := cli.Server(*suid)
	if err != nil {
		t.Error("query server:", err)
		return
	}

	if s.Status() != ServerStopped {
		if err := s.StopWait(); err != nil {
			t.Error("stop server:", err)
			return
		}
	}

	err = s.Remove(RecurseAllDrives)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("Server deleted")
}

func TestLiveDriveResize(t *testing.T) {
	u, p, err := parseCredentials()
	if u == "" {
		skipTest(t, err)
		return
	}

	if *duid == "" {
		t.Skip("-duid=<drive-uuid> must be specified")
		return
	}

	if *size == 0 {
		t.Skip("-size=<drive-size> must be specified")
		return
	}

	cli, err := NewClient(DefaultRegion, u, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	d, err := cli.Drive(*duid, libFlag())
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Current drive size: %d", d.Size())
	if err := d.ResizeWait(*size); err != nil {
		t.Error(err)
		return
	}
	t.Logf("Resulting drive size: %d", d.Size())
	t.Logf("%v", d)
}
