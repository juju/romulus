package plan_test

import (
	"encoding/json"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/wireformat/plan"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type planSuite struct{}

var _ = gc.Suite(&planSuite{})

func (t *planSuite) TestUnmarshalEnvService(c *gc.C) {
	data := []byte(`{
	"env-uuid": "some env",
	"charm-url": "some charm",
	"service-name": "some service",
	"plan-url": "some plan"
}`)
	var ar plan.AuthorizationRequest
	err := json.Unmarshal(data, &ar)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ar, gc.Equals, plan.AuthorizationRequest{
		EnvironmentUUID: "some env",
		CharmURL:        "some charm",
		ServiceName:     "some service",
		PlanURL:         "some plan",
	})
}

func (t *planSuite) TestUnmarshalModelApplication(c *gc.C) {
	data := []byte(`{
	"model-uuid": "some model",
	"charm-url": "some charm",
	"application-name": "some application",
	"plan-url": "some plan"
}`)
	var ar plan.AuthorizationRequest
	err := json.Unmarshal(data, &ar)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ar, gc.Equals, plan.AuthorizationRequest{
		EnvironmentUUID: "some model",
		CharmURL:        "some charm",
		ServiceName:     "some application",
		PlanURL:         "some plan",
	})
}
