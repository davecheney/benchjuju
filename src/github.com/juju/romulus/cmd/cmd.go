// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cmd

import (
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/persistent-cookiejar"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
)

// HttpCommand can instantiate http bakery clients using a commong cookie jar.
type HttpCommand struct {
	cmd.CommandBase

	cookiejar *cookiejar.Jar
}

// NewClient returns a new http bakery client for commands.
func (s *HttpCommand) NewClient() (*httpbakery.Client, error) {
	if s.cookiejar == nil {
		cookieFile := cookiejar.DefaultCookieFile()
		jar, err := cookiejar.New(&cookiejar.Options{
			Filename: cookieFile,
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
		s.cookiejar = jar
	}
	client := httpbakery.NewClient()
	client.Jar = s.cookiejar
	client.VisitWebPage = httpbakery.OpenWebBrowser
	return client, nil
}

// Close saves the persistent cookie jar used by the specified httpbakery.Client.
func (s *HttpCommand) Close() error {
	if s.cookiejar != nil {
		return s.cookiejar.Save()
	}
	return nil
}
