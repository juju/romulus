// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package allocate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/juju/cmd/envcmd"
	"github.com/juju/juju/environs/configstore"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"launchpad.net/gnuflag"

	api "github.com/juju/romulus/api/budget"
	rcmd "github.com/juju/romulus/cmd"
)

var budgetWithLimitRe = regexp.MustCompile(`^[a-zA-Z0-9\-]+:[1-9][0-9]*$`)

type allocateCommand struct {
	envcmd.EnvCommandBase
	rcmd.HttpCommand
	Budget   string
	Model    string
	Services []string
	Limit    string
}

// NewAllocateCommand returns a new allocateCommand
func NewAllocateCommand() cmd.Command {
	return envcmd.Wrap(&allocateCommand{})
}

const doc = `
Allocate budget for the specified services, replacing any prior allocations
made for the specified services.

Usage:

 juju allocate <budget>:<value> <service> [<service2> ...]

Example:

 juju allocate somebudget:42 db
     Assigns service "db" to an allocation on budget "somebudget" with the limit "42".
`

// Info implements cmd.Command.Info.
func (c *allocateCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "allocate",
		Purpose: "allocate budget to services",
		Doc:     doc,
	}
}

// SetFlags implements cmd.Command.
func (c *allocateCommand) SetFlags(f *gnuflag.FlagSet) {
	c.EnvCommandBase.SetFlags(f)
}

// AllowInterspersedFlags implements cmd.Command.
func (c *allocateCommand) AllowInterspersedFlags() bool { return true }

// IsSuperCommand implements cmd.Command.
func (c *allocateCommand) IsSuperCommand() bool { return false }

// Init implements cmd.Command.Init.
func (c *allocateCommand) Init(args []string) error {
	if len(args) < 2 {
		return errors.New("budget and service name required")
	}
	budgetWithLimit := args[0]
	var err error
	c.Budget, c.Limit, err = parseBudgetWithLimit(budgetWithLimit)
	if err != nil {
		return err
	}
	c.Model, err = c.modelUUID()
	if err != nil {
		return err
	}

	c.Services = args[1:]
	return nil
}

// Run implements cmd.Command.Run and has most of the logic for the run command.
func (c *allocateCommand) Run(ctx *cmd.Context) error {
	defer c.Close()
	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	api, err := newAPIClient(client)
	if err != nil {
		return errors.Annotate(err, "failed to create an api client")
	}
	resp, err := api.CreateAllocation(c.Budget, c.Limit, c.Model, c.Services)
	if err != nil {
		return errors.Annotate(err, "failed to create allocation")
	}
	fmt.Fprintf(ctx.Stdout, resp)
	return nil
}

func (c *allocateCommand) modelUUID() (string, error) {
	store, err := configstore.Default()
	if err != nil {
		return "", errors.Trace(err)
	}
	envInfo, err := store.ReadInfo(c.EnvName())
	if err != nil {
		return "", errors.Trace(err)
	}
	return envInfo.APIEndpoint().EnvironUUID, nil
}

func parseBudgetWithLimit(bl string) (string, string, error) {
	if !budgetWithLimitRe.MatchString(bl) {
		return "", "", errors.New("invalid budget specification, expecting <budget>:<limit>")
	}
	parts := strings.Split(bl, ":")
	return parts[0], parts[1], nil
}

var newAPIClient = newAPIClientImpl

func newAPIClientImpl(c *httpbakery.Client) (apiClient, error) {
	client := api.NewClient(c)
	return client, nil
}

type apiClient interface {
	CreateAllocation(string, string, string, []string) (string, error)
}
