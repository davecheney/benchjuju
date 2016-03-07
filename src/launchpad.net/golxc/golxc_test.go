// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

package golxc_test

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"launchpad.net/golxc"
)

var lxcfile = `# MIRROR to be used by ubuntu template at container creation:
# Leaving it undefined is fine
#MIRROR="http://archive.ubuntu.com/ubuntu"
# or
#MIRROR="http://<host-ip-addr>:3142/archive.ubuntu.com/ubuntu"

# LXC_AUTO - whether or not to start containers symlinked under
# /etc/lxc/auto
LXC_AUTO="true"

# Leave USE_LXC_BRIDGE as "true" if you want to use lxcbr0 for your
# containers.  Set to "false" if you'll use virbr0 or another existing
# bridge, or mavlan to your host's NIC.
USE_LXC_BRIDGE="true"

# LXC_BRIDGE and LXC_ADDR are changed against original for
# testing purposes.
# LXC_BRIDGE="lxcbr1"
LXC_BRIDGE="lxcbr9"
# LXC_ADDR="10.0.1.1"
LXC_ADDR="10.0.9.1"
LXC_NETMASK="255.255.255.0"
LXC_NETWORK="10.0.9.0/24"
LXC_DHCP_RANGE="10.0.9.2,10.0.9.254"
LXC_DHCP_MAX="253"
# And for testing LXC_BRIDGE="lxcbr99" and LXC_ADDR="10.0.99.1".

LXC_SHUTDOWN_TIMEOUT=120`

var lxcconf = map[string]string{
	"address": "10.0.9.1",
	"bridge":  "lxcbr9",
}

type ConfigSuite struct{}

var _ = gc.Suite(&ConfigSuite{})

func (s *ConfigSuite) TestReadConf(c *gc.C) {
	// Test reading the configuration.
	cf := filepath.Join(c.MkDir(), "lxc-test")
	c.Assert(ioutil.WriteFile(cf, []byte(lxcfile), 0555), gc.IsNil)

	defer golxc.SetConfPath(golxc.SetConfPath(cf))

	conf, err := golxc.ReadConf()
	c.Assert(err, gc.IsNil)
	c.Assert(conf, gc.DeepEquals, lxcconf)
}

func (s *ConfigSuite) TestReadNotExistingDefaultEnvironment(c *gc.C) {
	// Test reading a not existing environment.
	defer golxc.SetConfPath(golxc.SetConfPath(filepath.Join(c.MkDir(), "foo")))

	_, err := golxc.ReadConf()
	c.Assert(err, gc.ErrorMatches, "open .*: no such file or directory")
}

func (s *ConfigSuite) TestNetworkAttributes(c *gc.C) {
	// Test reading the network attribute form an environment.
	cf := filepath.Join(c.MkDir(), "lxc-test")
	c.Assert(ioutil.WriteFile(cf, []byte(lxcfile), 0555), gc.IsNil)

	defer golxc.SetConfPath(golxc.SetConfPath(cf))

	addr, bridge, err := golxc.NetworkAttributes()
	c.Assert(err, gc.IsNil)
	c.Assert(addr, gc.Equals, "10.0.9.1")
	c.Assert(bridge, gc.Equals, "lxcbr9")
}

type NetworkSuite struct{}

var _ = gc.Suite(&NetworkSuite{})

func (s *NetworkSuite) SetUpSuite(c *gc.C) {
	u, err := user.Current()
	c.Assert(err, gc.IsNil)
	if u.Uid != "0" {
		// Has to be run as root!
		c.Skip("tests must run as root")
	}
}

func (s *NetworkSuite) TestStartStopNetwork(c *gc.C) {
	// Test starting and stoping of the LXC network.
	initialRunning, err := golxc.IsNetworkRunning()
	c.Assert(err, gc.IsNil)
	defer func() {
		if initialRunning {
			c.Assert(golxc.StartNetwork(), gc.IsNil)
		}
	}()
	c.Assert(golxc.StartNetwork(), gc.IsNil)
	running, err := golxc.IsNetworkRunning()
	c.Assert(err, gc.IsNil)
	c.Assert(running, gc.Equals, true)
	c.Assert(golxc.StopNetwork(), gc.IsNil)
	running, err = golxc.IsNetworkRunning()
	c.Assert(err, gc.IsNil)
	c.Assert(running, gc.Equals, false)
}

func (s *NetworkSuite) TestNotExistingNetworkAttributes(c *gc.C) {
	// Test reading of network attributes from a not existing environment.
	defer golxc.SetConfPath(golxc.SetConfPath(filepath.Join(c.MkDir(), "foo")))

	_, _, err := golxc.NetworkAttributes()
	c.Assert(err, gc.ErrorMatches, "open .*: no such file or directory")
}

type LXCSuite struct {
	factory golxc.ContainerFactory
}

var _ = gc.Suite(&LXCSuite{golxc.Factory()})

func (s *LXCSuite) SetUpSuite(c *gc.C) {
	u, err := user.Current()
	c.Assert(err, gc.IsNil)
	if u.Uid != "0" {
		// Has to be run as root!
		c.Skip("tests must run as root")
	}
}

func (s *LXCSuite) createContainer(c *gc.C) golxc.Container {
	container := s.factory.New("golxc")
	c.Assert(container.IsConstructed(), gc.Equals, false)
	err := container.Create("", "ubuntu", nil, nil, nil)
	c.Assert(err, gc.IsNil)
	c.Assert(container.IsConstructed(), gc.Equals, true)
	return container
}

func (s *LXCSuite) TestCreateDestroy(c *gc.C) {
	// Test clean creation and destroying of a container.
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	home := golxc.ContainerHome(lc)
	_, err := os.Stat(home)
	c.Assert(err, gc.ErrorMatches, "stat .*: no such file or directory")
	err = lc.Create("", "ubuntu", nil, nil, nil)
	c.Assert(err, gc.IsNil)
	c.Assert(lc.IsConstructed(), gc.Equals, true)
	defer func() {
		err = lc.Destroy()
		c.Assert(err, gc.IsNil)
		_, err = os.Stat(home)
		c.Assert(err, gc.ErrorMatches, "stat .*: no such file or directory")
	}()
	fi, err := os.Stat(golxc.ContainerHome(lc))
	c.Assert(err, gc.IsNil)
	c.Assert(fi.IsDir(), gc.Equals, true)
}

func (s *LXCSuite) TestCreateTwice(c *gc.C) {
	// Test that a container cannot be created twice.
	lc1 := s.createContainer(c)
	c.Assert(lc1.IsConstructed(), gc.Equals, true)
	defer func() {
		c.Assert(lc1.Destroy(), gc.IsNil)
	}()
	lc2 := s.factory.New("golxc")
	err := lc2.Create("", "ubuntu", nil, nil, nil)
	c.Assert(err, gc.ErrorMatches, "container .* is already created")
}

func (s *LXCSuite) TestCreateIllegalTemplate(c *gc.C) {
	// Test that a container creation fails correctly in
	// case of an illegal template.
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	err := lc.Create("", "name-of-a-not-existing-template-for-golxc", nil, nil, nil)
	c.Assert(err, gc.ErrorMatches, `error executing "lxc-create": .*bad template.*`)
	c.Assert(lc.IsConstructed(), gc.Equals, false)
}

func (s *LXCSuite) TestDestroyNotCreated(c *gc.C) {
	// Test that a non-existing container can't be destroyed.
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	err := lc.Destroy()
	c.Assert(err, gc.ErrorMatches, "container .* is not yet created")
}

func contains(lcs []golxc.Container, lc golxc.Container) bool {
	for _, clc := range lcs {
		if clc.Name() == lc.Name() {
			return true
		}
	}
	return false
}

func (s *LXCSuite) TestList(c *gc.C) {
	// Test the listing of created containers.
	lcs, err := s.factory.List()
	oldLen := len(lcs)
	c.Assert(err, gc.IsNil)
	c.Assert(oldLen >= 0, gc.Equals, true)
	lc := s.createContainer(c)
	defer func() {
		c.Assert(lc.Destroy(), gc.IsNil)
	}()
	lcs, _ = s.factory.List()
	newLen := len(lcs)
	c.Assert(newLen == oldLen+1, gc.Equals, true)
	c.Assert(contains(lcs, lc), gc.Equals, true)
}

func (s *LXCSuite) TestClone(c *gc.C) {
	// Test the cloning of an existing container.
	lc1 := s.createContainer(c)
	defer func() {
		c.Assert(lc1.Destroy(), gc.IsNil)
	}()
	lcs, _ := s.factory.List()
	oldLen := len(lcs)
	lc2, err := lc1.Clone("golxcclone", nil, nil)
	c.Assert(err, gc.IsNil)
	c.Assert(lc2.IsConstructed(), gc.Equals, true)
	defer func() {
		c.Assert(lc2.Destroy(), gc.IsNil)
	}()
	lcs, _ = s.factory.List()
	newLen := len(lcs)
	c.Assert(newLen == oldLen+1, gc.Equals, true)
	c.Assert(contains(lcs, lc1), gc.Equals, true)
	c.Assert(contains(lcs, lc2), gc.Equals, true)
}

func (s *LXCSuite) TestCloneNotCreated(c *gc.C) {
	// Test the cloning of a non-existing container.
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	_, err := lc.Clone("golxcclone", nil, nil)
	c.Assert(err, gc.ErrorMatches, "container .* is not yet created")
}

func (s *LXCSuite) TestStartStop(c *gc.C) {
	// Test starting and stopping a container.
	lc := s.createContainer(c)
	defer func() {
		c.Assert(lc.Destroy(), gc.IsNil)
	}()
	c.Assert(lc.Start("", ""), gc.IsNil)
	c.Assert(lc.IsRunning(), gc.Equals, true)
	c.Assert(lc.Stop(), gc.IsNil)
	c.Assert(lc.IsRunning(), gc.Equals, false)
}

func (s *LXCSuite) TestStartNotCreated(c *gc.C) {
	// Test that a non-existing container can't be started.
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	c.Assert(lc.Start("", ""), gc.ErrorMatches, "container .* is not yet created")
}

func (s *LXCSuite) TestStopNotRunning(c *gc.C) {
	// Test that a not running container can't be stopped.
	lc := s.createContainer(c)
	defer func() {
		c.Assert(lc.Destroy(), gc.IsNil)
	}()
	c.Assert(lc.Stop(), gc.IsNil)
}

func (s *LXCSuite) TestWait(c *gc.C) {
	// Test waiting for one of a number of states of a container.
	// ATTN: Using a not reached state blocks the test until timeout!
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	c.Assert(lc.Wait(), gc.ErrorMatches, "no states specified")
	c.Assert(lc.Wait(golxc.StateStopped), gc.IsNil)
	c.Assert(lc.Wait(golxc.StateStopped, golxc.StateRunning), gc.IsNil)
	c.Assert(lc.Create("", "ubuntu", nil, nil, nil), gc.IsNil)
	defer func() {
		c.Assert(lc.Destroy(), gc.IsNil)
	}()
	go func() {
		c.Assert(lc.Start("", ""), gc.IsNil)
	}()
	c.Assert(lc.Wait(golxc.StateRunning), gc.IsNil)
}

func (s *LXCSuite) TestFreezeUnfreeze(c *gc.C) {
	// Test the freezing and unfreezing of a started container.
	lc := s.createContainer(c)
	defer func() {
		c.Assert(lc.Destroy(), gc.IsNil)
	}()
	c.Assert(lc.Start("", ""), gc.IsNil)
	defer func() {
		c.Assert(lc.Stop(), gc.IsNil)
	}()
	c.Assert(lc.IsRunning(), gc.Equals, true)
	c.Assert(lc.Freeze(), gc.IsNil)
	c.Assert(lc.IsRunning(), gc.Equals, false)
	c.Assert(lc.Unfreeze(), gc.IsNil)
	c.Assert(lc.IsRunning(), gc.Equals, true)
}

func (s *LXCSuite) TestFreezeNotStarted(c *gc.C) {
	// Test that a not running container can't be frozen.
	lc := s.createContainer(c)
	defer func() {
		c.Assert(lc.Destroy(), gc.IsNil)
	}()
	c.Assert(lc.Freeze(), gc.ErrorMatches, "container .* is not running")
}

func (s *LXCSuite) TestFreezeNotCreated(c *gc.C) {
	// Test that a non-existing container can't be frozen.
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	c.Assert(lc.Freeze(), gc.ErrorMatches, "container .* is not yet created")
}

func (s *LXCSuite) TestUnfreezeNotCreated(c *gc.C) {
	// Test that a non-existing container can't be unfrozen.
	lc := s.factory.New("golxc")
	c.Assert(lc.IsConstructed(), gc.Equals, false)
	c.Assert(lc.Unfreeze(), gc.ErrorMatches, "container .* is not yet created")
}

func (s *LXCSuite) TestUnfreezeNotFrozen(c *gc.C) {
	// Test that a running container can't be unfrozen.
	lc := s.createContainer(c)
	defer func() {
		c.Assert(lc.Destroy(), gc.IsNil)
	}()
	c.Assert(lc.Start("", ""), gc.IsNil)
	defer func() {
		c.Assert(lc.Stop(), gc.IsNil)
	}()
	c.Assert(lc.Unfreeze(), gc.ErrorMatches, "container .* is not frozen")
}

type commandArgs struct {
	testing.IsolationSuite
	originalPath string
}

var _ = gc.Suite(&commandArgs{})

func (s *commandArgs) SetUpSuite(c *gc.C) {
	// lxc-create requires the PATH to be set.
	s.originalPath = os.Getenv("PATH")
	s.IsolationSuite.SetUpSuite(c)
}

func (s *commandArgs) setupLxcStart(c *gc.C) {
	dir := c.MkDir()
	// Make the rootfs for the "test" container so it thinks it is created.
	rootfs := filepath.Join(dir, "test", "rootfs")
	err := os.MkdirAll(rootfs, 0755)
	c.Assert(err, gc.IsNil)
	c.Assert(rootfs, jc.IsDirectory)

	s.PatchValue(&golxc.ContainerDir, dir)
	s.PatchEnvironment("PATH", s.originalPath)
	testing.PatchExecutableAsEchoArgs(c, s, "lxc-wait")
	testing.PatchExecutableAsEchoArgs(c, s, "lxc-start")
	info := "PID: 1234\nSTATE: RUNNING"
	testing.PatchExecutable(c, s, "lxc-info", "#!/bin/sh\necho '"+info+"'\n")
}

func (s *commandArgs) TestStartArgs(c *gc.C) {
	s.setupLxcStart(c)
	factory := golxc.Factory()
	container := factory.New("test")
	err := container.Start("config-file", "console-log")
	c.Assert(err, gc.IsNil)
	testing.AssertEchoArgs(
		c, "lxc-start",
		"--daemon",
		"--name", "test",
		"--rcfile", "config-file",
		"--console-log", "console-log")
}

func (s *commandArgs) TestStartArgsNoConsoleLog(c *gc.C) {
	s.setupLxcStart(c)
	factory := golxc.Factory()
	container := factory.New("test")
	err := container.Start("config-file", "")
	c.Assert(err, gc.IsNil)
	testing.AssertEchoArgs(
		c, "lxc-start",
		"--daemon",
		"--name", "test",
		"--rcfile", "config-file")
}

func (s *commandArgs) TestStartArgsFallback(c *gc.C) {
	dir := c.MkDir()
	// Make the rootfs for the "test" container so it thinks it is created.
	rootfs := filepath.Join(dir, "test", "rootfs")
	err := os.MkdirAll(rootfs, 0755)
	c.Assert(err, gc.IsNil)
	c.Assert(rootfs, jc.IsDirectory)

	s.PatchValue(&golxc.ContainerDir, dir)
	s.PatchEnvironment("PATH", s.originalPath)
	testing.PatchExecutableAsEchoArgs(c, s, "lxc-wait")
	EchoQuotedArgs := `
name=` + "`basename $0`" + `
argfile="$name.out"
rm -f $argfile
printf "%s" $name | tee -a $argfile
for arg in "$@"; do
  printf " \"%s\""  "$arg" | tee -a $argfile
done
printf "\n" | tee -a $argfile
`
	// Make lxc-start error if called with --console-log but succeed otherwise.
	// On success, echo the args.
	testing.PatchExecutable(
		c, s, "lxc-start", "#!/bin/bash\nif [ $6 == '--console-log' ]; then\nexit 1\nelse\n"+EchoQuotedArgs+"fi")
	info := "PID: 1234\nSTATE: RUNNING"
	testing.PatchExecutable(c, s, "lxc-info", "#!/bin/sh\necho '"+info+"'\n")

	factory := golxc.Factory()
	container := factory.New("test")
	err = container.Start("config-file", "console-log")
	c.Assert(err, gc.IsNil)
	testing.AssertEchoArgs(
		c, "lxc-start",
		"--daemon",
		"--name", "test",
		"--rcfile", "config-file",
		"--console", "console-log")
}

func (s *commandArgs) TestCreateArgs(c *gc.C) {
	s.PatchValue(&golxc.ContainerDir, c.MkDir())
	s.PatchEnvironment("PATH", s.originalPath)
	testing.PatchExecutableAsEchoArgs(c, s, "lxc-create")

	factory := golxc.Factory()
	container := factory.New("test")
	err := container.Create(
		"config-file", "template",
		[]string{"extra-1", "extra-2"},
		[]string{"template-1", "template-2"},
		nil,
	)
	c.Assert(err, gc.IsNil)
	testing.AssertEchoArgs(
		c, "lxc-create",
		"-n", "test",
		"-t", "template",
		"-f", "config-file",
		"extra-1", "extra-2",
		"--", "template-1", "template-2")
}

func (s *commandArgs) TestCreateWithEnv(c *gc.C) {
	s.PatchValue(&golxc.ContainerDir, c.MkDir())
	s.PatchEnvironment("PATH", s.originalPath)

	EchoEnvHello := `#!/bin/bash
name=` + "`basename $0`" + `
argfile="$name.out"
rm -f $argfile
printf "%s" $name | tee -a $argfile
printf " \"%s\""  "$HELLO" | tee -a $argfile
if [ -n "$FOO" ]; then
    printf " \"%s\""  "$FOO" | tee -a $argfile
fi
printf "\n" | tee -a $argfile
`

	testing.PatchExecutable(
		c, s, "lxc-create", EchoEnvHello)
	s.AddCleanup(func(*gc.C) {
		os.Remove("lxc-create.out")
	})

	// Add an environment variable to the calling process
	// to show that these don't get passed through.
	s.PatchEnvironment("FOO", "BAR")
	// Only the explicitly set env variables are propagated.
	envArgs := []string{"HELLO=WORLD!"}
	factory := golxc.Factory()
	container := factory.New("test")
	err := container.Create(
		"config-file", "template",
		[]string{"extra-1", "extra-2"},
		[]string{"template-1", "template-2"},
		envArgs,
	)
	c.Assert(err, gc.IsNil)
	testing.AssertEchoArgs(c, "lxc-create", "WORLD!")
}

func (s *commandArgs) TestCloneArgs(c *gc.C) {
	dir := c.MkDir()
	s.PatchValue(&golxc.ContainerDir, dir)
	s.PatchEnvironment("PATH", s.originalPath)
	// Patch lxc-info too as clone checks to see if it is running.
	testing.PatchExecutableAsEchoArgs(c, s, "lxc-info")
	testing.PatchExecutableAsEchoArgs(c, s, "lxc-clone")

	factory := golxc.Factory()
	container := factory.New("test")
	// Make the rootfs for the "test" container so it thinks it is created.
	rootfs := filepath.Join(dir, "test", "rootfs")
	err := os.MkdirAll(rootfs, 0755)
	c.Assert(err, gc.IsNil)
	c.Assert(rootfs, jc.IsDirectory)
	c.Assert(container.IsConstructed(), jc.IsTrue)
	clone, err := container.Clone(
		"name",
		[]string{"extra-1", "extra-2"},
		[]string{"template-1", "template-2"},
	)
	c.Assert(err, gc.IsNil)
	testing.AssertEchoArgs(
		c, "lxc-clone",
		"-o", "test",
		"-n", "name",
		"extra-1", "extra-2",
		"--", "template-1", "template-2")
	c.Assert(clone.Name(), gc.Equals, "name")
}

type UtilsSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&UtilsSuite{})

func (s *UtilsSuite) TestGetDefaultLXCContainerDirDefaultValue(c *gc.C) {
	testing.PatchExecutable(c, s, "lxc-config", "#!/bin/sh\nexit -1\n")
	c.Assert(golxc.GetDefaultLXCContainerDir(), gc.Equals, golxc.DefaultLXCDir)
}

func (s *UtilsSuite) TestGetDefaultLXCContainerDir(c *gc.C) {
	const path = "/var/lib/non-standard-lxc"
	testing.PatchExecutable(c, s, "lxc-config", "#!/bin/sh\necho '"+path+"'\n")
	c.Assert(golxc.GetDefaultLXCContainerDir(), gc.Equals, path)
}
