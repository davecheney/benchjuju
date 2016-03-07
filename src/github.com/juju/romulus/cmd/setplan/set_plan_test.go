// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package setplan_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	stdtesting "testing"

	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/errors"
	jjjtesting "github.com/juju/juju/juju/testing"
	"github.com/juju/juju/testcharms"
	jjtesting "github.com/juju/juju/testing"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/macaroon-bakery.v1/bakery"
	"gopkg.in/macaroon-bakery.v1/bakery/checkers"
	"gopkg.in/macaroon.v1"

	"github.com/juju/romulus/cmd/setplan"
)

func TestPackage(t *stdtesting.T) {
	jjtesting.MgoTestPackage(t)
}

var _ = gc.Suite(&setPlanCommandSuite{})

type setPlanCommandSuite struct {
	jjjtesting.JujuConnSuite

	mockAPI  *mockapi
	charmURL string
}

func (s *setPlanCommandSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	ch := testcharms.Repo.CharmDir("dummy")
	curl := charm.MustParseURL(
		fmt.Sprintf("local:quantal/%s-%d", ch.Meta().Name, ch.Revision()),
	)
	s.charmURL = curl.String()
	dummyCharm, err := s.State.AddCharm(ch, curl, "dummy-path", "dummy-1")
	c.Assert(err, jc.ErrorIsNil)
	s.AddTestingService(c, "mysql", dummyCharm)

	mockAPI, err := newMockAPI()
	c.Assert(err, jc.ErrorIsNil)
	s.mockAPI = mockAPI

	s.PatchValue(setplan.NewAuthorizationClient, setplan.APIClientFnc(s.mockAPI))
}

func (s setPlanCommandSuite) TestSetPlanCommand(c *gc.C) {
	tests := []struct {
		about    string
		plan     string
		service  string
		err      string
		apiErr   error
		apiCalls []testing.StubCall
	}{{
		about:   "all is well",
		plan:    "bob/default",
		service: "mysql",
		apiCalls: []testing.StubCall{{
			FuncName: "Authorize",
			Args: []interface{}{
				s.State.ModelUUID(),
				s.charmURL,
				"mysql",
			},
		}},
	}, {
		about:   "invalid service name",
		plan:    "bob/default",
		service: "mysql-0",
		err:     "invalid service name \"mysql-0\"",
	}, {
		about:   "unknown service",
		plan:    "bob/default",
		service: "wordpress",
		err:     "service \"wordpress\" not found.*",
	}, {
		about:   "unknown service",
		plan:    "bob/default",
		service: "mysql",
		apiErr:  errors.New("some strange error"),
		err:     "some strange error",
	},
	}
	for i, test := range tests {
		c.Logf("running test %d: %v", i, test.about)
		s.mockAPI.ResetCalls()
		if test.apiErr != nil {
			s.mockAPI.SetErrors(test.apiErr)
		}
		_, err := cmdtesting.RunCommand(c, setplan.NewSetPlanCommand(), test.service, test.plan)
		if test.err == "" {
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 1)
			s.mockAPI.CheckCalls(c, test.apiCalls)

			svc, err := s.State.Service("mysql")
			c.Assert(err, jc.ErrorIsNil)
			svcMacaroon := svc.MetricCredentials()
			data, err := json.Marshal(macaroon.Slice{s.mockAPI.macaroon})
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(svcMacaroon, gc.DeepEquals, data)
		} else {
			c.Assert(err, gc.ErrorMatches, test.err)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 0)
		}
	}
}

func newMockAPI() (*mockapi, error) {
	kp, err := bakery.GenerateKey()
	if err != nil {
		return nil, errors.Trace(err)
	}
	svc, err := bakery.NewService(bakery.NewServiceParams{
		Location: "omnibus",
		Key:      kp,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &mockapi{
		service: svc,
	}, nil
}

type mockapi struct {
	testing.Stub

	service  *bakery.Service
	macaroon *macaroon.Macaroon
}

func (m *mockapi) Authorize(environmentUUID, charmURL, serviceName, plan string, visitWebPage func(*url.URL) error) (*macaroon.Macaroon, error) {
	err := m.NextErr()
	if err != nil {
		return nil, errors.Trace(err)
	}
	m.AddCall("Authorize", environmentUUID, charmURL, serviceName)
	macaroon, err := m.service.NewMacaroon(
		"",
		nil,
		[]checkers.Caveat{
			checkers.DeclaredCaveat("environment", environmentUUID),
			checkers.DeclaredCaveat("charm", charmURL),
			checkers.DeclaredCaveat("service", serviceName),
			checkers.DeclaredCaveat("plan", plan),
		},
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	m.macaroon = macaroon
	return m.macaroon, nil
}
