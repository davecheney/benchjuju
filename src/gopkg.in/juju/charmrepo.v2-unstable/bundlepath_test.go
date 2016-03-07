// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package charmrepo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/yaml.v1"

	"gopkg.in/juju/charmrepo.v2-unstable"
)

type bundlePathSuite struct {
	repoPath string
}

var _ = gc.Suite(&bundlePathSuite{})

func (s *bundlePathSuite) SetUpTest(c *gc.C) {
	s.repoPath = c.MkDir()
}

func (s *bundlePathSuite) cloneCharmDir(path, name string) string {
	return TestCharms.ClonedDirPath(path, name)
}

func (s *bundlePathSuite) TestNoPath(c *gc.C) {
	_, _, err := charmrepo.NewBundleAtPath("")
	c.Assert(err, gc.ErrorMatches, "path to bundle not specified")
}

func (s *bundlePathSuite) TestInvalidPath(c *gc.C) {
	_, _, err := charmrepo.NewBundleAtPath("/foo")
	c.Assert(err, gc.Equals, os.ErrNotExist)
}

func (s *bundlePathSuite) TestRepoURL(c *gc.C) {
	_, _, err := charmrepo.NewCharmAtPath("cs:foo", "trusty")
	c.Assert(err, gc.Equals, os.ErrNotExist)
}

func (s *bundlePathSuite) TestInvalidRelativePath(c *gc.C) {
	_, _, err := charmrepo.NewBundleAtPath("./foo")
	c.Assert(err, gc.Equals, os.ErrNotExist)
}

func (s *bundlePathSuite) TestRelativePath(c *gc.C) {
	relDir := filepath.Join(TestCharms.Path(), "bundle")
	cwd, err := os.Getwd()
	c.Assert(err, jc.ErrorIsNil)
	defer os.Chdir(cwd)
	c.Assert(os.Chdir(relDir), jc.ErrorIsNil)
	_, _, err = charmrepo.NewBundleAtPath("openstack")
	c.Assert(charmrepo.IsInvalidPathError(err), jc.IsTrue)
}

func (s *bundlePathSuite) TestNoBundleAtPath(c *gc.C) {
	_, _, err := charmrepo.NewBundleAtPath(c.MkDir())
	c.Assert(err, gc.ErrorMatches, `bundle not found:.*`)
}

func (s *bundlePathSuite) TestGetBundle(c *gc.C) {
	bundleDir := filepath.Join(TestCharms.Path(), "bundle", "openstack")
	b, url, err := charmrepo.NewBundleAtPath(bundleDir)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(b.Data(), jc.DeepEquals, TestCharms.BundleDir("openstack").Data())
	c.Assert(url, gc.DeepEquals, charm.MustParseURL("local:bundle/openstack-0"))
}

func (s *bundlePathSuite) TestGetBundleSymlink(c *gc.C) {
	realPath := TestCharms.ClonedBundleDirPath(c.MkDir(), "wordpress-simple")
	bundlesPath := c.MkDir()
	linkPath := filepath.Join(bundlesPath, "wordpress-simple")
	err := os.Symlink(realPath, linkPath)
	c.Assert(err, jc.ErrorIsNil)
	url := charm.MustParseURL("local:bundle/wordpress-simple")

	b, url, err := charmrepo.NewBundleAtPath(filepath.Join(bundlesPath, "wordpress-simple"))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(b.Data(), jc.DeepEquals, TestCharms.BundleDir("wordpress-simple").Data())
	c.Assert(url, gc.DeepEquals, charm.MustParseURL("local:bundle/wordpress-simple-0"))
}

func (s *bundlePathSuite) TestGetBundleLocalFile(c *gc.C) {
	bundlePath := filepath.Join(c.MkDir(), "mybundle")
	data := `
services:
  wordpress:
    charm: wordpress
    num_units: 1
`[1:]
	err := ioutil.WriteFile(bundlePath, []byte(data), 0644)
	c.Assert(err, jc.ErrorIsNil)

	bundleData, err := charmrepo.ReadBundleFile(bundlePath)
	c.Assert(err, jc.ErrorIsNil)
	out, err := yaml.Marshal(bundleData)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), jc.DeepEquals, data)
}

func (s *bundlePathSuite) TestGetBundleLocalFileNotExists(c *gc.C) {
	bundlePath := filepath.Join(c.MkDir(), "mybundle")
	_, err := charmrepo.ReadBundleFile(bundlePath)
	c.Assert(err, gc.ErrorMatches, `bundle not found:.*`)
}
