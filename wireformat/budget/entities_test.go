// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package budget_test

import (
	"sort"
	"testing"

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
	}, {
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model1",
	}}

	expected := []budget.Allocation{{
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model1",
	}, {
		Owner:    "user",
		Limit:    "40",
		Consumed: "10",
		Usage:    "25%",
		Model:    "model2",
	}}

	sort.Sort(budget.SortedAllocations(allocations))
	c.Assert(allocations, gc.DeepEquals, expected)
}
