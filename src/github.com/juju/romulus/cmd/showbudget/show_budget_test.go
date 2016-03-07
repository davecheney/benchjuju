// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.s

package showbudget_test

import (
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/cmd/showbudget"
	"github.com/juju/romulus/wireformat/budget"
)

var _ = gc.Suite(&showBudgetSuite{})

type showBudgetSuite struct {
	testing.CleanupSuite
	stub    *testing.Stub
	mockAPI *mockapi
}

func (s *showBudgetSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}
	s.mockAPI = &mockapi{s.stub}
	s.PatchValue(showbudget.NewAPIClient, showbudget.APIClientFnc(s.mockAPI))
}

func (s *showBudgetSuite) TestShowBudgetCommand(c *gc.C) {
	tests := []struct {
		about  string
		args   []string
		err    string
		budget string
		apierr string
	}{{
		about: "missing argument",
		err:   `missing arguments`,
	}, {
		about: "unknown arguments",
		args:  []string{"my-special-budget", "extra", "arguments"},
		err:   `unrecognized args: \["extra" "arguments"\]`,
	}, {
		about:  "api error",
		args:   []string{"personal"},
		apierr: "well, this is embarrassing",
		err:    "failed to retrieve the budget: well, this is embarrassing",
	}, {
		about:  "all ok",
		args:   []string{"personal"},
		budget: "personal",
	},
	}

	for i, test := range tests {
		c.Logf("running test %d: %v", i, test.about)
		s.mockAPI.ResetCalls()

		if test.apierr != "" {
			s.mockAPI.SetErrors(errors.New(test.apierr))
		}

		showBudget := showbudget.NewShowBudgetCommand()

		_, err := cmdtesting.RunCommand(c, showBudget, test.args...)
		if test.err == "" {
			c.Assert(err, jc.ErrorIsNil)
			s.stub.CheckCalls(c, []testing.StubCall{{"GetBudget", []interface{}{test.budget}}})
		} else {
			c.Assert(err, gc.ErrorMatches, test.err)
		}
	}
}

type mockapi struct {
	*testing.Stub
}

func (api *mockapi) GetBudget(name string) (*budget.BudgetWithAllocations, error) {
	api.AddCall("GetBudget", name)
	if err := api.NextErr(); err != nil {
		return nil, err
	}
	return &budget.BudgetWithAllocations{
		Limit: "4000.00",
		Total: budget.BudgetTotals{
			Allocated:   "2200.00",
			Unallocated: "1800.00",
			Available:   "1100,00",
			Consumed:    "1100.0",
			Usage:       "50%",
		},
		Allocations: []budget.Allocation{{
			Owner:    "user.joe",
			Limit:    "1200.00",
			Consumed: "500.00",
			Usage:    "42%",
			Model:    "model.joe",
			Services: map[string]budget.ServiceAllocation{
				"wordpress": budget.ServiceAllocation{
					Consumed: "300.00",
				},
				"mysql": budget.ServiceAllocation{
					Consumed: "200.00",
				},
			},
		}, {
			Owner:    "user.jess",
			Limit:    "1000.00",
			Consumed: "600.00",
			Usage:    "60%",
			Model:    "model.jess",
			Services: map[string]budget.ServiceAllocation{
				"landscape": budget.ServiceAllocation{
					Consumed: "600.00",
				},
			},
		},
		},
	}, nil
}
