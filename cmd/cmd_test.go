// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cmd_test

import (
	gc "gopkg.in/check.v1"
	stdtesting "testing"

	jujucmd "github.com/juju/cmd"
	"github.com/juju/cmd/cmdtesting"
	coretesting "github.com/juju/juju/testing"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"

	"github.com/juju/romulus/cmd"
)

func TestPackage(t *stdtesting.T) {
	gc.TestingT(t)
}

type httpSuite struct {
	coretesting.FakeJujuXDGDataHomeSuite
	testing.CleanupSuite
	caCert string
}

var _ = gc.Suite(&httpSuite{})

type testCommand struct {
	cmd.HttpCommand
}

func (c *testCommand) Info() *jujucmd.Info {
	return &jujucmd.Info{Name: "test"}
}

func (c *testCommand) Run(ctx *jujucmd.Context) error {
	return nil
}

func (s *httpSuite) SetUpTest(c *gc.C) {
	s.CleanupSuite.SetUpTest(c)
	s.FakeJujuXDGDataHomeSuite.SetUpTest(c)
}

func (s *httpSuite) TearDownTest(c *gc.C) {
	s.FakeJujuXDGDataHomeSuite.TearDownTest(c)
	s.CleanupSuite.TearDownTest(c)
}

func (s *httpSuite) TestNewClient(c *gc.C) {
	basecmd := &testCommand{}
	defer basecmd.Close()

	ctx, err := cmdtesting.RunCommand(c, basecmd)
	c.Assert(err, jc.ErrorIsNil)

	client, err := basecmd.NewClient(ctx)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(client, gc.NotNil)
}
