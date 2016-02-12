// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package updateallocation_test

import (
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/errors"
	"github.com/juju/juju/environs/configstore"
	coretesting "github.com/juju/juju/testing"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/cmd/updateallocation"
)

var _ = gc.Suite(&updateAllocationSuite{})

type updateAllocationSuite struct {
	coretesting.FakeJujuXDGDataHomeSuite
	stub    *testing.Stub
	mockAPI *mockapi
}

func (s *updateAllocationSuite) SetUpTest(c *gc.C) {
	s.FakeJujuXDGDataHomeSuite.SetUpTest(c)
	c.Log(coretesting.SingleEnvConfig)
	store, err := configstore.Default()
	c.Assert(err, jc.ErrorIsNil)
	info := store.CreateInfo(coretesting.SampleModelName)
	apiEndpoint := configstore.APIEndpoint{
		ModelUUID: "env-uuid",
	}
	info.SetAPIEndpoint(apiEndpoint)
	err = info.Write()
	c.Assert(err, jc.ErrorIsNil)
	s.stub = &testing.Stub{}
	s.mockAPI = newMockAPI(s.stub)
	s.PatchValue(updateallocation.NewAPIClient, updateallocation.APIClientFnc(s.mockAPI))
}

func (s *updateAllocationSuite) TestUpdateAllocation(c *gc.C) {
	s.mockAPI.resp = "name budget set to 5"
	set := updateallocation.NewUpdateAllocationCommand()
	ctx, err := cmdtesting.RunCommand(c, set, "name", "5")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cmdtesting.Stdout(ctx), jc.DeepEquals, "name budget set to 5")
	s.mockAPI.CheckCall(c, 0, "UpdateAllocation", "env-uuid", "name", "5")
}

func (s *updateAllocationSuite) TestUpdateAllocationAPIError(c *gc.C) {
	s.stub.SetErrors(errors.New("something failed"))
	set := updateallocation.NewUpdateAllocationCommand()
	_, err := cmdtesting.RunCommand(c, set, "name", "5")
	c.Assert(err, gc.ErrorMatches, "failed to update the allocation: something failed")
	s.mockAPI.CheckCall(c, 0, "UpdateAllocation", "env-uuid", "name", "5")
}

func (s *updateAllocationSuite) TestUpdateAllocationErrors(c *gc.C) {
	tests := []struct {
		about         string
		args          []string
		expectedError string
	}{
		{
			about:         "value needs to be a number",
			args:          []string{"name", "badvalue"},
			expectedError: "value needs to be a whole number",
		},
		{
			about:         "value is missing",
			args:          []string{"name"},
			expectedError: "service and value required",
		},
		{
			about:         "no args",
			args:          []string{},
			expectedError: "service and value required",
		},
	}
	for i, test := range tests {
		s.mockAPI.ResetCalls()
		c.Logf("test %d: %s", i, test.about)
		set := updateallocation.NewUpdateAllocationCommand()
		_, err := cmdtesting.RunCommand(c, set, test.args...)
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

func (api *mockapi) UpdateAllocation(modelUUID, name, value string) (string, error) {
	api.MethodCall(api, "UpdateAllocation", modelUUID, name, value)
	return api.resp, api.NextErr()
}
