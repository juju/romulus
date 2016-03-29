// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.s

package showbudget_test

import (
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/errors"
	coretesting "github.com/juju/juju/testing"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/cmd/showbudget"
	"github.com/juju/romulus/wireformat/budget"
)

var _ = gc.Suite(&showBudgetSuite{})

type showBudgetSuite struct {
	coretesting.FakeJujuXDGDataHomeSuite
	stub    *testing.Stub
	mockAPI *mockapi
}

func (s *showBudgetSuite) SetUpTest(c *gc.C) {
	s.CleanupSuite.SetUpTest(c)
	s.FakeJujuXDGDataHomeSuite.SetUpTest(c)
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
		output string
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
		output: "" +
			"MODEL      \tSERVICES \tSPENT\tALLOCATED\tBY       \tUSAGE\n" +
			"model.joe  \tmysql    \t200  \t1200     \tuser.joe \t42%  \n" +
			"           \twordpress\t300  \t         \t         \n" +
			"model.jess \tlandscape\t600  \t1000     \tuser.jess\t60%  \n" +
			"           \t         \t     \t         \t         \n" +
			"TOTAL      \t         \t1100 \t2200     \t         \t50%  \n" +
			"BUDGET     \t         \t     \t4000     \t         \n" +
			"UNALLOCATED\t         \t     \t1800     \t         \n",
	}}

	for i, test := range tests {
		c.Logf("running test %d: %v", i, test.about)
		s.mockAPI.ResetCalls()

		if test.apierr != "" {
			s.mockAPI.SetErrors(errors.New(test.apierr))
		}

		showBudget := showbudget.NewShowBudgetCommand()

		ctx, err := cmdtesting.RunCommand(c, showBudget, test.args...)
		if test.err == "" {
			c.Assert(err, jc.ErrorIsNil)
			s.stub.CheckCalls(c, []testing.StubCall{{"GetBudget", []interface{}{test.budget}}})
			output := cmdtesting.Stdout(ctx)
			c.Assert(output, gc.Equals, test.output)
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
		Limit: "4000",
		Total: budget.BudgetTotals{
			Allocated:   "2200",
			Unallocated: "1800",
			Available:   "1100",
			Consumed:    "1100",
			Usage:       "50%",
		},
		Allocations: []budget.Allocation{{
			Owner:    "user.joe",
			Limit:    "1200",
			Consumed: "500",
			Usage:    "42%",
			Model:    "model.joe",
			Services: map[string]budget.ServiceAllocation{
				"wordpress": budget.ServiceAllocation{
					Consumed: "300",
				},
				"mysql": budget.ServiceAllocation{
					Consumed: "200",
				},
			},
		}, {
			Owner:    "user.jess",
			Limit:    "1000",
			Consumed: "600",
			Usage:    "60%",
			Model:    "model.jess",
			Services: map[string]budget.ServiceAllocation{
				"landscape": budget.ServiceAllocation{
					Consumed: "600",
				},
			},
		}}}, nil
}
