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
