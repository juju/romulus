// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package metrics_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/wireformat/metrics"
)

type metricsSuite struct {
}

var _ = gc.Suite(&metricsSuite{})

func (s *metricsSuite) TestAck(c *gc.C) {
	resp := metrics.EnvironmentResponses{}
	c.Assert(resp, gc.HasLen, 0)

	modelUUID := "model-uuid"
	modelUUID2 := "model-uuid2"
	batchUUID := "batch-uuid"
	batchUUID2 := "batch-uuid2"

	resp.Ack(modelUUID, batchUUID)
	resp.Ack(modelUUID, batchUUID2)
	resp.Ack(modelUUID2, batchUUID)
	c.Assert(resp, gc.HasLen, 2)

	c.Assert(resp[modelUUID].AcknowledgedBatches, jc.SameContents, []string{batchUUID, batchUUID2})
	c.Assert(resp[modelUUID2].AcknowledgedBatches, jc.SameContents, []string{batchUUID})
}

func (s *metricsSuite) TestSetUnitStatus(c *gc.C) {
	resp := metrics.EnvironmentResponses{}
	c.Assert(resp, gc.HasLen, 0)

	modelUUID := "model-uuid"
	modelUUID2 := "model-uuid2"
	unitName := "some-unit/0"
	unitName2 := "some-unit/1"

	resp.SetUnitStatus(modelUUID, unitName, "GREEN", "")
	c.Assert(resp, gc.HasLen, 1)
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Status, gc.Equals, "GREEN")
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Info, gc.Equals, "")

	resp.SetUnitStatus(modelUUID, unitName2, "RED", "Unit unresponsive.")
	c.Assert(resp, gc.HasLen, 1)
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Status, gc.Equals, "GREEN")
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Info, gc.Equals, "")
	c.Assert(resp[modelUUID].UnitStatuses[unitName2].Status, gc.Equals, "RED")
	c.Assert(resp[modelUUID].UnitStatuses[unitName2].Info, gc.Equals, "Unit unresponsive.")

	resp.SetUnitStatus(modelUUID2, unitName, "UNKNOWN", "")
	c.Assert(resp, gc.HasLen, 2)
	c.Assert(resp[modelUUID2].UnitStatuses[unitName].Status, gc.Equals, "UNKNOWN")
	c.Assert(resp[modelUUID2].UnitStatuses[unitName].Info, gc.Equals, "")

	resp.SetUnitStatus(modelUUID, unitName, "RED", "Invalid data received.")
	c.Assert(resp, gc.HasLen, 2)
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Status, gc.Equals, "RED")
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Info, gc.Equals, "Invalid data received.")
}

func (s *metricsSuite) TestSetModelStatus(c *gc.C) {
	resp := metrics.EnvironmentResponses{}
	c.Assert(resp, gc.HasLen, 0)

	for i, test := range []struct {
		status, info string
	}{{
		"GREEN", "it's good",
	}, {
		"AMBER", "outlook unclear",
	}, {
		"RED", "oh no",
	}} {
		c.Logf("test#%d", i)
		resp.SetModelStatus("model-uuid", test.status, test.info)
		c.Check(resp["model-uuid"].ModelStatus.Status, gc.Equals, test.status)
		c.Check(resp["model-uuid"].ModelStatus.Info, gc.Equals, test.info)
		c.Check(resp, gc.HasLen, 1)
	}

	resp.SetModelStatus("model-uuid2", "GREEN", "good")
	c.Assert(resp, gc.HasLen, 2)
}
