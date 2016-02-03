// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package allocate_test

import (
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/errors"
	"github.com/juju/juju/environs/configstore"
	coretesting "github.com/juju/juju/testing"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/cmd/allocate"
)

var _ = gc.Suite(&allocateSuite{})

type allocateSuite struct {
	coretesting.FakeJujuHomeSuite
	stub    *testing.Stub
	mockAPI *mockapi
}

func (s *allocateSuite) SetUpTest(c *gc.C) {
	s.FakeJujuHomeSuite.SetUpTest(c)
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
	s.PatchValue(allocate.NewAPIClient, allocate.APIClientFnc(s.mockAPI))
}

func (s *allocateSuite) TestAllocate(c *gc.C) {
	s.mockAPI.resp = "allocation updated"
	alloc := allocate.NewAllocateCommand()
	ctx, err := cmdtesting.RunCommand(c, alloc, "name:100", "db")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cmdtesting.Stdout(ctx), jc.DeepEquals, "allocation updated")
	s.mockAPI.CheckCall(c, 0, "CreateAllocation", "name", "100", "env-uuid", []string{"db"})
}

func (s *allocateSuite) TestallocateAPIError(c *gc.C) {
	s.stub.SetErrors(errors.New("something failed"))
	set := allocate.NewAllocateCommand()
	_, err := cmdtesting.RunCommand(c, set, "name:100", "db")
	c.Assert(err, gc.ErrorMatches, "failed to create allocation: something failed")
	s.mockAPI.CheckCall(c, 0, "CreateAllocation", "name", "100", "env-uuid", []string{"db"})
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
		set := allocate.NewAllocateCommand()
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

func (api *mockapi) CreateAllocation(name, limit, modelUUID string, services []string) (string, error) {
	api.MethodCall(api, "CreateAllocation", name, limit, modelUUID, services)
	return api.resp, api.NextErr()
}
