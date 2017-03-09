// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package metrics_test

import (
	"encoding/json"

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

func (s *metricsSuite) TestSetStatus(c *gc.C) {
	resp := metrics.EnvironmentResponses{}
	c.Assert(resp, gc.HasLen, 0)

	modelUUID := "model-uuid"
	modelUUID2 := "model-uuid2"
	unitName := "some-unit/0"
	unitName2 := "some-unit/1"

	resp.SetStatus(modelUUID, unitName, "GREEN", "")
	c.Assert(resp, gc.HasLen, 1)
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Status, gc.Equals, "GREEN")
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Info, gc.Equals, "")

	resp.SetStatus(modelUUID, unitName2, "RED", "Unit unresponsive.")
	c.Assert(resp, gc.HasLen, 1)
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Status, gc.Equals, "GREEN")
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Info, gc.Equals, "")
	c.Assert(resp[modelUUID].UnitStatuses[unitName2].Status, gc.Equals, "RED")
	c.Assert(resp[modelUUID].UnitStatuses[unitName2].Info, gc.Equals, "Unit unresponsive.")

	resp.SetStatus(modelUUID2, unitName, "UNKNOWN", "")
	c.Assert(resp, gc.HasLen, 2)
	c.Assert(resp[modelUUID2].UnitStatuses[unitName].Status, gc.Equals, "UNKNOWN")
	c.Assert(resp[modelUUID2].UnitStatuses[unitName].Info, gc.Equals, "")

	resp.SetStatus(modelUUID, unitName, "RED", "Invalid data received.")
	c.Assert(resp, gc.HasLen, 2)
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Status, gc.Equals, "RED")
	c.Assert(resp[modelUUID].UnitStatuses[unitName].Info, gc.Equals, "Invalid data received.")
}

func (t *metricsSuite) TestUnmarshalEnvUUID(c *gc.C) {
	data := []byte(`{
	"uuid": "some batch",
	"env-uuid": "some env",
	"unit-name": "some unit",
	"charm-url": "some charm"
}`)
	var mb metrics.MetricBatch
	err := json.Unmarshal(data, &mb)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(mb.ModelUUID, gc.Equals, "some env")
}

func (t *metricsSuite) TestUnmarshalModelUUID(c *gc.C) {
	data := []byte(`{
	"uuid": "some batch",
	"model-uuid": "some model",
	"unit-name": "some unit",
	"charm-url": "some charm"
}`)
	var mb metrics.MetricBatch
	err := json.Unmarshal(data, &mb)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(mb.ModelUUID, gc.Equals, "some model")
}

func (t *metricsSuite) TestUnmarshalResponseService(c *gc.C) {
	data := []byte(`{
	"uuid": "some uuid",
	"env-responses": {
		"one": {
			"acks": ["a", "b", "c"],
			"unit-statuses": {
				"foo": {
					"status": "good",
					"info": "times"
				}
			}
		}
	}
}`)
	var r metrics.Response
	err := json.Unmarshal(data, &r)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(r.EnvResponses, gc.NotNil)
	c.Assert(r.EnvResponses["one"], jc.DeepEquals, metrics.EnvResponse{
		AcknowledgedBatches: []string{"a", "b", "c"},
		UnitStatuses: map[string]metrics.UnitStatus{
			"foo": metrics.UnitStatus{Status: "good", Info: "times"},
		},
	})
}

func (t *metricsSuite) TestUnmarshalResponseApplication(c *gc.C) {
	data := []byte(`{
	"uuid": "some uuid",
	"model-responses": {
		"two": {
			"acks": ["d", "e", "f"],
			"unit-statuses": {
				"bar": {"status": "none"}
			}
		}
	}
}`)
	var r metrics.Response
	err := json.Unmarshal(data, &r)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(r.EnvResponses, gc.NotNil)
	c.Assert(r.EnvResponses["two"], jc.DeepEquals, metrics.EnvResponse{
		AcknowledgedBatches: []string{"d", "e", "f"},
		UnitStatuses: map[string]metrics.UnitStatus{
			"bar": metrics.UnitStatus{Status: "none"},
		},
	})
}
