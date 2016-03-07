package jujusvg

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"
)

type IconFetcherSuite struct{}

var _ = gc.Suite(&IconFetcherSuite{})

func (s *IconFetcherSuite) TestLinkFetchIcons(c *gc.C) {
	tests := map[string][]byte{
		"~charming-devs/precise/elasticsearch-2": []byte(`
			<svg xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/~charming-devs/precise/elasticsearch-2.svg" />
			</svg>`),
		"~juju-jitsu/precise/charmworld-58": []byte(`
			<svg xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/~juju-jitsu/precise/charmworld-58.svg" />
			</svg>`),
		"precise/mongodb-21": []byte(`
			<svg xmlns:xlink="http://www.w3.org/1999/xlink">
				<image width="96" height="96" xlink:href="/precise/mongodb-21.svg" />
			</svg>`),
	}
	iconURL := func(ref *charm.URL) string {
		return "/" + ref.Path() + ".svg"
	}
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil, nil)
	c.Assert(err, gc.IsNil)
	fetcher := LinkFetcher{
		IconURL: iconURL,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	for charm, link := range tests {
		assertXMLEqual(c, []byte(iconMap[charm]), []byte(link))
	}
}

func (s *IconFetcherSuite) TestHTTPFetchIcons(c *gc.C) {
	fetchCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		fmt.Fprintln(w, fmt.Sprintf("<svg>%s</svg>", r.URL.Path))
	}))
	defer ts.Close()

	tsIconURL := func(ref *charm.URL) string {
		return ts.URL + "/" + ref.Path() + ".svg"
	}
	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil, nil)
	c.Assert(err, gc.IsNil)
	// Only one copy of precise/mongodb-21
	b.Services["duplicateService"] = &charm.ServiceSpec{
		Charm:    "cs:precise/mongodb-21",
		NumUnits: 1,
	}
	fetcher := HTTPFetcher{
		Concurrency: 1,
		IconURL:     tsIconURL,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	c.Assert(iconMap, gc.DeepEquals, map[string][]byte{
		"~charming-devs/precise/elasticsearch-2": []byte("<svg>/~charming-devs/precise/elasticsearch-2.svg</svg>\n"),
		"~juju-jitsu/precise/charmworld-58":      []byte("<svg>/~juju-jitsu/precise/charmworld-58.svg</svg>\n"),
		"precise/mongodb-21":                     []byte("<svg>/precise/mongodb-21.svg</svg>\n"),
	})

	fetcher.Concurrency = 10
	iconMap, err = fetcher.FetchIcons(b)
	c.Assert(err, gc.IsNil)
	c.Assert(iconMap, gc.DeepEquals, map[string][]byte{
		"~charming-devs/precise/elasticsearch-2": []byte("<svg>/~charming-devs/precise/elasticsearch-2.svg</svg>\n"),
		"~juju-jitsu/precise/charmworld-58":      []byte("<svg>/~juju-jitsu/precise/charmworld-58.svg</svg>\n"),
		"precise/mongodb-21":                     []byte("<svg>/precise/mongodb-21.svg</svg>\n"),
	})
	c.Assert(fetchCount, gc.Equals, 6)
}

func (s *IconFetcherSuite) TestHTTPBadIconURL(c *gc.C) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad-wolf", http.StatusForbidden)
		return
	}))
	defer ts.Close()

	tsIconURL := func(ref *charm.URL) string {
		return ts.URL + "/" + ref.Path() + ".svg"
	}

	b, err := charm.ReadBundleData(strings.NewReader(bundle))
	c.Assert(err, gc.IsNil)
	err = b.Verify(nil, nil)
	c.Assert(err, gc.IsNil)
	fetcher := HTTPFetcher{
		Concurrency: 1,
		IconURL:     tsIconURL,
	}
	iconMap, err := fetcher.FetchIcons(b)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("cannot retrieve icon from %s.+\\.svg: 403 Forbidden.*", ts.URL))
	c.Assert(iconMap, gc.IsNil)

	fetcher.Concurrency = 10
	iconMap, err = fetcher.FetchIcons(b)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("cannot retrieve icon from %s.+\\.svg: 403 Forbidden.*", ts.URL))
	c.Assert(iconMap, gc.IsNil)
}
