// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package allocate_test

import (
	"github.com/juju/cmd"
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/errors"
	"github.com/juju/juju/jujuclient"
	"github.com/juju/juju/jujuclient/jujuclienttesting"
	"github.com/juju/testing"
	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/cmd/allocate"
)

var _ = gc.Suite(&allocateSuite{})

type allocateSuite struct {
	jujutesting.FakeHomeSuite
	stub    *testing.Stub
	mockAPI *mockapi
	store   jujuclient.ClientStore
}

func (s *allocateSuite) SetUpTest(c *gc.C) {
	s.FakeHomeSuite.SetUpTest(c)
	s.store = &jujuclienttesting.MemStore{
		Controllers: map[string]jujuclient.ControllerDetails{
			"controller": {},
		},
		Models: map[string]jujuclient.ControllerAccountModels{
			"controller": {
				AccountModels: map[string]*jujuclient.AccountModels{
					"admin@local": {
						Models: map[string]jujuclient.ModelDetails{
							"model": {"model-uuid"},
						},
						CurrentModel: "model",
					},
				},
			},
		},
		Accounts: map[string]*jujuclient.ControllerAccounts{
			"controller": {
				Accounts: map[string]jujuclient.AccountDetails{
					"admin@local": {},
				},
				CurrentAccount: "admin@local",
			},
		},
	}
	s.stub = &testing.Stub{}
	s.mockAPI = newMockAPI(s.stub)
}

func (s *allocateSuite) run(c *gc.C, args ...string) (*cmd.Context, error) {
	alloc := allocate.NewAllocateCommandForTest(s.mockAPI, s.store)
	a := []string{"-m", "controller:model"}
	a = append(a, args...)
	return cmdtesting.RunCommand(c, alloc, a...)
}

func (s *allocateSuite) TestAllocate(c *gc.C) {
	s.mockAPI.resp = "allocation updated"
	ctx, err := s.run(c, "name:100", "db")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cmdtesting.Stdout(ctx), jc.DeepEquals, "allocation updated")
	s.mockAPI.CheckCall(c, 0, "CreateAllocation", "name", "100", "model-uuid", []string{"db"})
}

func (s *allocateSuite) TestallocateAPIError(c *gc.C) {
	s.stub.SetErrors(errors.New("something failed"))
	_, err := s.run(c, "name:100", "db")
	c.Assert(err, gc.ErrorMatches, "failed to create allocation: something failed")
	s.mockAPI.CheckCall(c, 0, "CreateAllocation", "name", "100", "model-uuid", []string{"db"})
}

func (s *allocateSuite) TestAllocateErrors(c *gc.C) {
	tests := []struct {
		about         string
		args          []string
		expectedError string
	}{{
		about:         "no args",
		args:          []string{},
		expectedError: "budget and service name required",
	}, {
		about:         "budget without allocation limit",
		args:          []string{"name", "db"},
		expectedError: "invalid budget specification, expecting <budget>:<limit>",
	}, {
		about:         "service not specified",
		args:          []string{"name:100"},
		expectedError: "budget and service name required",
	}}
	for i, test := range tests {
		c.Logf("test %d: %s", i, test.about)
		_, err := s.run(c, test.args...)
		c.Check(err, gc.ErrorMatches, test.expectedError)
		s.mockAPI.CheckNoCalls(c)
	}
}

func newMockAPI(s *testing.Stub) *mockapi {
	return &mockapi{Stub: s}
}

type mockapi struct {
	*testing.Stub
	resp string
}

func (api *mockapi) CreateAllocation(name, limit, modelUUID string, services []string) (string, error) {
	api.MethodCall(api, "CreateAllocation", name, limit, modelUUID, services)
	return api.resp, api.NextErr()
}
