package budget_test

import (
	"encoding/json"
	"sort"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/wireformat/budget"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type BudgetSuite struct{}

var _ = gc.Suite(&BudgetSuite{})

func (t *BudgetSuite) TestAllocationSorting(c *gc.C) {
	allocations := []budget.Allocation{{
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model2",
		Services: map[string]budget.ServiceAllocation{
			"mongo": budget.ServiceAllocation{
				Consumed: "10",
			},
		},
	}, {
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model1",
		Services: map[string]budget.ServiceAllocation{
			"mysql": budget.ServiceAllocation{
				Consumed: "10",
			},
			"abc": budget.ServiceAllocation{
				Consumed: "10",
			},
		},
	}, {
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model1",
		Services: map[string]budget.ServiceAllocation{
			"mongo": budget.ServiceAllocation{
				Consumed: "10",
			},
			"apache": budget.ServiceAllocation{
				Consumed: "10",
			},
		},
	}}

	expected := []budget.Allocation{{
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model1",
		Services: map[string]budget.ServiceAllocation{
			"mysql": budget.ServiceAllocation{
				Consumed: "10",
			},
			"abc": budget.ServiceAllocation{
				Consumed: "10",
			},
		},
	}, {
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model1",
		Services: map[string]budget.ServiceAllocation{
			"mongo": budget.ServiceAllocation{
				Consumed: "10",
			},
			"apache": budget.ServiceAllocation{
				Consumed: "10",
			},
		},
	}, {
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model2",
		Services: map[string]budget.ServiceAllocation{
			"mongo": budget.ServiceAllocation{
				Consumed: "10",
			},
		},
	}}

	sort.Sort(budget.SortedAllocations(allocations))
	c.Assert(allocations, gc.DeepEquals, expected)
}

func (t *BudgetSuite) TestUnmarshalServices(c *gc.C) {
	data := []byte(`{
	"owner": "bob",
	"limit": "sky",
	"consumed": "much",
	"usage": "angry",
	"model": "citizen",
	"services": {
		"practical": {
			"consumed": "vivaciously"
		}
	}
}`)
	var alloc budget.Allocation
	err := json.Unmarshal(data, &alloc)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(alloc.Model, gc.Equals, "citizen")
	c.Assert(alloc.Services, gc.HasLen, 1)
	c.Assert(alloc.Services["practical"].Consumed, gc.Equals, "vivaciously")
}

func (t *BudgetSuite) TestUnmarshalApplications(c *gc.C) {
	data := []byte(`{
	"owner": "bob",
	"limit": "sky",
	"consumed": "much",
	"usage": "angry",
	"model": "citizen",
	"applications": {
		"theoretical": {
			"consumed": "voraciously"
		}
	}
}`)
	var alloc budget.Allocation
	err := json.Unmarshal(data, &alloc)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(alloc.Model, gc.Equals, "citizen")
	c.Assert(alloc.Services, gc.HasLen, 1)
	c.Assert(alloc.Services["theoretical"].Consumed, gc.Equals, "voraciously")
}
