// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package charmrepo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	gitjujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"

	"gopkg.in/juju/charmrepo.v2-unstable"
)

type LocalRepoSuite struct {
	gitjujutesting.FakeHomeSuite
	repo        *charmrepo.LocalRepository
	charmsPath  string
	bundlesPath string
}

var _ = gc.Suite(&LocalRepoSuite{})

func (s *LocalRepoSuite) SetUpTest(c *gc.C) {
	s.FakeHomeSuite.SetUpTest(c)
	root := c.MkDir()
	s.repo = &charmrepo.LocalRepository{Path: root}
	s.bundlesPath = filepath.Join(root, "bundle")
	s.charmsPath = filepath.Join(root, "quantal")
	c.Assert(os.Mkdir(s.bundlesPath, 0777), jc.ErrorIsNil)
	c.Assert(os.Mkdir(s.charmsPath, 0777), jc.ErrorIsNil)
}

func (s *LocalRepoSuite) addCharmArchive(name string) string {
	return TestCharms.CharmArchivePath(s.charmsPath, name)
}

func (s *LocalRepoSuite) addCharmDir(name string) string {
	return TestCharms.ClonedDirPath(s.charmsPath, name)
}

func (s *LocalRepoSuite) addBundleDir(name string) string {
	return TestCharms.ClonedBundleDirPath(s.bundlesPath, name)
}

func (s *LocalRepoSuite) checkNotFoundErr(c *gc.C, err error, charmURL *charm.URL) {
	expect := `entity not found in "` + s.repo.Path + `": ` + charmURL.String()
	c.Check(err, gc.ErrorMatches, expect)
}

func (s *LocalRepoSuite) TestMissingCharm(c *gc.C) {
	for i, str := range []string{
		"local:quantal/zebra", "local:badseries/zebra",
	} {
		c.Logf("test %d: %s", i, str)
		charmURL := charm.MustParseURL(str)
		_, err := s.repo.Get(charmURL)
		s.checkNotFoundErr(c, err, charmURL)
	}
}

func (s *LocalRepoSuite) TestMissingRepo(c *gc.C) {
	c.Assert(os.RemoveAll(s.repo.Path), gc.IsNil)
	_, err := s.repo.Get(charm.MustParseURL("local:quantal/zebra"))
	c.Assert(err, gc.ErrorMatches, `no repository found at ".*"`)
	_, err = s.repo.GetBundle(charm.MustParseURL("local:bundle/wordpress-simple"))
	c.Assert(err, gc.ErrorMatches, `no repository found at ".*"`)
	c.Assert(ioutil.WriteFile(s.repo.Path, nil, 0666), gc.IsNil)
	_, err = s.repo.Get(charm.MustParseURL("local:quantal/zebra"))
	c.Assert(err, gc.ErrorMatches, `no repository found at ".*"`)
	_, err = s.repo.GetBundle(charm.MustParseURL("local:bundle/wordpress-simple"))
	c.Assert(err, gc.ErrorMatches, `no repository found at ".*"`)
}

func (s *LocalRepoSuite) TestCharmArchive(c *gc.C) {
	charmURL := charm.MustParseURL("local:quantal/dummy")
	s.addCharmArchive("dummy")

	ch, err := s.repo.Get(charmURL)
	c.Assert(err, gc.IsNil)
	c.Assert(ch.Revision(), gc.Equals, 1)
}

func (s *LocalRepoSuite) TestLogsErrors(c *gc.C) {
	err := ioutil.WriteFile(filepath.Join(s.charmsPath, "blah.charm"), nil, 0666)
	c.Assert(err, gc.IsNil)
	err = os.Mkdir(filepath.Join(s.charmsPath, "blah"), 0666)
	c.Assert(err, gc.IsNil)
	samplePath := s.addCharmDir("upgrade2")
	gibberish := []byte("don't parse me by")
	err = ioutil.WriteFile(filepath.Join(samplePath, "metadata.yaml"), gibberish, 0666)
	c.Assert(err, gc.IsNil)

	charmURL := charm.MustParseURL("local:quantal/dummy")
	s.addCharmDir("dummy")
	ch, err := s.repo.Get(charmURL)
	c.Assert(err, gc.IsNil)
	c.Assert(ch.Revision(), gc.Equals, 1)
	c.Assert(c.GetTestLog(), gc.Matches, `
.* WARNING juju.charm.charmrepo failed to load charm at ".*/quantal/blah": .*
.* WARNING juju.charm.charmrepo failed to load charm at ".*/quantal/blah.charm": .*
.* WARNING juju.charm.charmrepo failed to load charm at ".*/quantal/upgrade2": .*
`[1:])
}

func renameSibling(c *gc.C, path, name string) {
	c.Assert(os.Rename(path, filepath.Join(filepath.Dir(path), name)), gc.IsNil)
}

func (s *LocalRepoSuite) TestIgnoresUnpromisingNames(c *gc.C) {
	err := ioutil.WriteFile(filepath.Join(s.charmsPath, "blah.notacharm"), nil, 0666)
	c.Assert(err, gc.IsNil)
	err = os.Mkdir(filepath.Join(s.charmsPath, ".blah"), 0666)
	c.Assert(err, gc.IsNil)
	renameSibling(c, s.addCharmDir("dummy"), ".dummy")
	renameSibling(c, s.addCharmArchive("dummy"), "dummy.notacharm")
	charmURL := charm.MustParseURL("local:quantal/dummy")

	_, err = s.repo.Get(charmURL)
	s.checkNotFoundErr(c, err, charmURL)
	c.Assert(c.GetTestLog(), gc.Equals, "")
}

func (s *LocalRepoSuite) TestFindsSymlinks(c *gc.C) {
	realPath := TestCharms.ClonedDirPath(c.MkDir(), "dummy")
	linkPath := filepath.Join(s.charmsPath, "dummy")
	err := os.Symlink(realPath, linkPath)
	c.Assert(err, gc.IsNil)
	ch, err := s.repo.Get(charm.MustParseURL("local:quantal/dummy"))
	c.Assert(err, gc.IsNil)
	c.Assert(ch.Revision(), gc.Equals, 1)
	c.Assert(ch.Meta().Name, gc.Equals, "dummy")
	c.Assert(ch.Config().Options["title"].Default, gc.Equals, "My Title")
	c.Assert(ch.(*charm.CharmDir).Path, gc.Equals, linkPath)
}

func (s *LocalRepoSuite) TestResolve(c *gc.C) {
	// Add some charms to the local repo.
	s.addCharmDir("upgrade1")
	s.addCharmDir("upgrade2")
	s.addCharmDir("wordpress")
	s.addCharmDir("riak")
	s.addCharmDir("multi-series")
	s.addCharmDir("multi-series-bad")

	// Define the tests to be run.
	tests := []struct {
		id     string
		url    string
		series []string
		err    string
	}{{
		id:  "local:quantal/upgrade",
		url: "local:quantal/upgrade-2",
	}, {
		id:  "local:quantal/upgrade-1",
		url: "local:quantal/upgrade-1",
	}, {
		id:  "local:quantal/wordpress",
		url: "local:quantal/wordpress-3",
	}, {
		id:  "local:quantal/riak",
		url: "local:quantal/riak-7",
	}, {
		id:  "local:quantal/wordpress-3",
		url: "local:quantal/wordpress-3",
	}, {
		id:  "local:quantal/wordpress-2",
		url: "local:quantal/wordpress-2",
	}, {
		id:     "local:quantal/new-charm-with-multi-series",
		url:    "local:quantal/new-charm-with-multi-series-7",
		series: []string{},
	}, {
		id:  "local:quantal/multi-series-bad",
		err: `series \"quantal\" not supported by charm, supported series are: precise,trusty`,
	}, {
		id:  "local:bundle/openstack",
		url: "local:bundle/openstack-0",
	}, {
		id:  "local:bundle/openstack-42",
		url: "local:bundle/openstack-42",
	}, {
		id:  "local:trusty/riak",
		err: "entity not found .*: local:trusty/riak",
	}, {
		id:  "local:quantal/no-such",
		err: "entity not found .*: local:quantal/no-such",
	}, {
		id:  "local:upgrade",
		err: "no series specified for local:upgrade",
	}}

	// Run the tests.
	for i, test := range tests {
		c.Logf("test %d: %s", i, test.id)
		ref, series, err := s.repo.Resolve(charm.MustParseURL(test.id))
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
			c.Assert(ref, gc.IsNil)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(ref, jc.DeepEquals, charm.MustParseURL(test.url))
		c.Assert(series, jc.DeepEquals, test.series)
	}
}

func (s *LocalRepoSuite) TestGetBundle(c *gc.C) {
	url := charm.MustParseURL("local:bundle/openstack")
	s.addBundleDir("openstack")
	b, err := s.repo.GetBundle(url)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(b.Data(), jc.DeepEquals, TestCharms.BundleDir("openstack").Data())
}

func (s *LocalRepoSuite) TestGetBundleSymlink(c *gc.C) {
	realPath := TestCharms.ClonedBundleDirPath(c.MkDir(), "wordpress-simple")
	linkPath := filepath.Join(s.bundlesPath, "wordpress-simple")
	err := os.Symlink(realPath, linkPath)
	c.Assert(err, jc.ErrorIsNil)
	url := charm.MustParseURL("local:bundle/wordpress-simple")
	b, err := s.repo.GetBundle(url)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(b.Data(), jc.DeepEquals, TestCharms.BundleDir("wordpress-simple").Data())
}

func (s *LocalRepoSuite) TestGetBundleErrorNotFound(c *gc.C) {
	url := charm.MustParseURL("local:bundle/no-such")
	b, err := s.repo.GetBundle(url)
	s.checkNotFoundErr(c, err, url)
	c.Assert(b, gc.IsNil)
}

var invalidURLTests = []struct {
	about  string
	bundle bool
	url    string
	err    string
}{{
	about: "get charm: non-local schema",
	url:   "cs:trusty/django-42",
	err:   `local repository got URL with non-local schema: "cs:trusty/django-42"`,
}, {
	about:  "get bundle: non-local schema",
	bundle: true,
	url:    "cs:bundle/django-scalable",
	err:    `local repository got URL with non-local schema: "cs:bundle/django-scalable"`,
}, {
	about: "get charm: bundle provided",
	url:   "local:bundle/rails",
	err:   `expected a charm URL, got bundle URL "local:bundle/rails"`,
}, {
	about:  "get bundle: charm provided",
	bundle: true,
	url:    "local:trusty/rails",
	err:    `expected a bundle URL, got charm URL "local:trusty/rails"`,
}}

func (s *LocalRepoSuite) TestInvalidURLTest(c *gc.C) {
	var err error
	var e interface{}
	for i, test := range invalidURLTests {
		c.Logf("test %d: %s", i, test.about)
		curl := charm.MustParseURL(test.url)
		if test.bundle {
			e, err = s.repo.GetBundle(curl)
		} else {
			e, err = s.repo.Get(curl)
		}
		c.Assert(e, gc.IsNil)
		c.Assert(err, gc.ErrorMatches, test.err)
	}
}
