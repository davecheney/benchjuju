// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package charmrepo_test

import (
	"os"
	"path/filepath"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"

	"gopkg.in/juju/charmrepo.v2-unstable"
)

type charmPathSuite struct {
	repoPath string
}

var _ = gc.Suite(&charmPathSuite{})

func (s *charmPathSuite) SetUpTest(c *gc.C) {
	s.repoPath = c.MkDir()
}

func (s *charmPathSuite) cloneCharmDir(path, name string) string {
	return TestCharms.ClonedDirPath(path, name)
}

func (s *charmPathSuite) TestNoPath(c *gc.C) {
	_, _, err := charmrepo.NewCharmAtPath("", "trusty")
	c.Assert(err, gc.ErrorMatches, "empty charm path")
}

func (s *charmPathSuite) TestInvalidPath(c *gc.C) {
	_, _, err := charmrepo.NewCharmAtPath("/foo", "trusty")
	c.Assert(err, gc.Equals, os.ErrNotExist)
}

func (s *charmPathSuite) TestRepoURL(c *gc.C) {
	_, _, err := charmrepo.NewCharmAtPath("cs:foo", "trusty")
	c.Assert(err, gc.Equals, os.ErrNotExist)
}

func (s *charmPathSuite) TestInvalidRelativePath(c *gc.C) {
	_, _, err := charmrepo.NewCharmAtPath("./foo", "trusty")
	c.Assert(err, gc.Equals, os.ErrNotExist)
}

func (s *charmPathSuite) TestRelativePath(c *gc.C) {
	s.cloneCharmDir(s.repoPath, "mysql")
	cwd, err := os.Getwd()
	c.Assert(err, jc.ErrorIsNil)
	defer os.Chdir(cwd)
	c.Assert(os.Chdir(s.repoPath), jc.ErrorIsNil)
	_, _, err = charmrepo.NewCharmAtPath("mysql", "trusty")
	c.Assert(charmrepo.IsInvalidPathError(err), jc.IsTrue)
}

func (s *charmPathSuite) TestNoCharmAtPath(c *gc.C) {
	_, _, err := charmrepo.NewCharmAtPath(c.MkDir(), "trusty")
	c.Assert(err, gc.ErrorMatches, "charm not found.*")
}

func (s *charmPathSuite) TestCharm(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "mysql")
	s.cloneCharmDir(s.repoPath, "mysql")
	ch, url, err := charmrepo.NewCharmAtPath(charmDir, "quantal")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ch.Meta().Name, gc.Equals, "mysql")
	c.Assert(ch.Revision(), gc.Equals, 1)
	c.Assert(url, gc.DeepEquals, charm.MustParseURL("local:quantal/mysql-1"))
}

func (s *charmPathSuite) TestNoSeriesSpecified(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "mysql")
	s.cloneCharmDir(s.repoPath, "mysql")
	_, _, err := charmrepo.NewCharmAtPath(charmDir, "")
	c.Assert(err, gc.ErrorMatches, "series not specified and charm does not define any")
}

func (s *charmPathSuite) TestNoSeriesSpecifiedForceStillFails(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "mysql")
	s.cloneCharmDir(s.repoPath, "mysql")
	_, _, err := charmrepo.NewCharmAtPathForceSeries(charmDir, "", true)
	c.Assert(err, gc.ErrorMatches, "series not specified and charm does not define any")
}

func (s *charmPathSuite) TestMuliSeriesDefault(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "multi-series")
	s.cloneCharmDir(s.repoPath, "multi-series")
	ch, url, err := charmrepo.NewCharmAtPath(charmDir, "")
	c.Assert(err, gc.IsNil)
	c.Assert(ch.Meta().Name, gc.Equals, "new-charm-with-multi-series")
	c.Assert(ch.Revision(), gc.Equals, 7)
	c.Assert(url, gc.DeepEquals, charm.MustParseURL("local:precise/multi-series-7"))
}

func (s *charmPathSuite) TestMuliSeries(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "multi-series")
	s.cloneCharmDir(s.repoPath, "multi-series")
	ch, url, err := charmrepo.NewCharmAtPath(charmDir, "trusty")
	c.Assert(err, gc.IsNil)
	c.Assert(ch.Meta().Name, gc.Equals, "new-charm-with-multi-series")
	c.Assert(ch.Revision(), gc.Equals, 7)
	c.Assert(url, gc.DeepEquals, charm.MustParseURL("local:trusty/multi-series-7"))
}

func (s *charmPathSuite) TestUnsupportedSeries(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "multi-series")
	s.cloneCharmDir(s.repoPath, "multi-series")
	_, _, err := charmrepo.NewCharmAtPath(charmDir, "wily")
	c.Assert(err, gc.ErrorMatches, `series "wily" not supported by charm, supported series are.*`)
}

func (s *charmPathSuite) TestUnsupportedSeriesNoForce(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "multi-series")
	s.cloneCharmDir(s.repoPath, "multi-series")
	_, _, err := charmrepo.NewCharmAtPathForceSeries(charmDir, "wily", false)
	c.Assert(err, gc.ErrorMatches, `series "wily" not supported by charm, supported series are.*`)
}

func (s *charmPathSuite) TestUnsupportedSeriesForce(c *gc.C) {
	charmDir := filepath.Join(s.repoPath, "multi-series")
	s.cloneCharmDir(s.repoPath, "multi-series")
	ch, url, err := charmrepo.NewCharmAtPathForceSeries(charmDir, "wily", true)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ch.Meta().Name, gc.Equals, "new-charm-with-multi-series")
	c.Assert(ch.Revision(), gc.Equals, 7)
	c.Assert(url, gc.DeepEquals, charm.MustParseURL("local:wily/multi-series-7"))
}

func (s *charmPathSuite) TestFindsSymlinks(c *gc.C) {
	realPath := TestCharms.ClonedDirPath(c.MkDir(), "dummy")
	charmsPath := c.MkDir()
	linkPath := filepath.Join(charmsPath, "dummy")
	err := os.Symlink(realPath, linkPath)
	c.Assert(err, gc.IsNil)

	ch, url, err := charmrepo.NewCharmAtPath(filepath.Join(charmsPath, "dummy"), "quantal")
	c.Assert(err, gc.IsNil)
	c.Assert(ch.Revision(), gc.Equals, 1)
	c.Assert(ch.Meta().Name, gc.Equals, "dummy")
	c.Assert(ch.Config().Options["title"].Default, gc.Equals, "My Title")
	c.Assert(ch.(*charm.CharmDir).Path, gc.Equals, linkPath)
	c.Assert(url, gc.DeepEquals, charm.MustParseURL("local:quantal/dummy-1"))
}
