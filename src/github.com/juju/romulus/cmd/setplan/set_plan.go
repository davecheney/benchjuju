// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// The setplan package contains the implementation of the juju set-plan
// command.
package setplan

import (
	"encoding/json"
	"net/url"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/juju/api/service"
	"github.com/juju/juju/cmd/modelcmd"
	"github.com/juju/names"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"gopkg.in/macaroon.v1"
	"launchpad.net/gnuflag"

	api "github.com/juju/romulus/api/plan"
	rcmd "github.com/juju/romulus/cmd"
)

// authorizationClient defines the interface of an api client that
// the comand uses to create an authorization macaroon.
type authorizationClient interface {
	// Authorize returns the authorization macaroon for the specified environment,
	// charm url, service name and plan.
	Authorize(environmentUUID, charmURL, serviceName, plan string, visitWebPage func(*url.URL) error) (*macaroon.Macaroon, error)
}

var newAuthorizationClient = func(options ...api.ClientOption) (authorizationClient, error) {
	return api.NewAuthorizationClient(options...)
}

// NewSetPlanCommand returns a new command that is used to set metric credentials for a
// deployed service.
func NewSetPlanCommand() cmd.Command {
	return modelcmd.Wrap(&setPlanCommand{})
}

// setPlanCommand is a command-line tool for setting
// Service.MetricCredential for development & demonstration purposes.
type setPlanCommand struct {
	modelcmd.ModelCommandBase
	rcmd.HttpCommand

	Service string
	Plan    string
}

// Info implements cmd.Command.
func (c *setPlanCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "set-plan",
		Args:    "<service name> <plan>",
		Purpose: "set the plan for a service",
		Doc: `
Set the plan for the deployed service, effective immediately.

The specified plan name must be a valid plan that is offered for this particular charm. Use "juju list-plans <charm>" for more information.
	
Usage:

 juju set-plan [options] <service name> <plan name>

Example:

 juju set-plan myapp example/uptime
`,
	}
}

// SetFlags implements cmd.Command.
func (c *setPlanCommand) SetFlags(f *gnuflag.FlagSet) {
	c.ModelCommandBase.SetFlags(f)
}

// Init implements cmd.Command.
func (c *setPlanCommand) Init(args []string) error {
	if len(args) < 2 {
		return errors.New("need to specify plan uuid and service name")
	}

	serviceName := args[0]
	if !names.IsValidService(serviceName) {
		return errors.Errorf("invalid service name %q", serviceName)
	}

	c.Plan = args[1]
	c.Service = serviceName

	return c.ModelCommandBase.Init(args[2:])
}

// IsSuperCommand implements cmd.Command.
// Defined here because of ambiguity between HttpCommand and ModelCommandBase.
func (c *setPlanCommand) IsSuperCommand() bool { return false }

// AllowInterspersed implements cmd.Command.
// Defined here because of ambiguity between HttpCommand and ModelCommandBase.
func (c *setPlanCommand) AllowInterspersedFlags() bool { return true }

func (c *setPlanCommand) requestMetricCredentials() ([]byte, error) {
	root, err := c.NewAPIRoot()
	if err != nil {
		return nil, errors.Trace(err)
	}
	jclient := service.NewClient(root)
	envUUID := jclient.ModelUUID()
	charmURL, err := jclient.GetCharmURL(c.Service)
	if err != nil {
		return nil, errors.Trace(err)
	}

	hc, err := c.NewClient()
	if err != nil {
		return nil, errors.Trace(err)
	}
	client, err := newAuthorizationClient(api.HTTPClient(hc))
	if err != nil {
		return nil, errors.Trace(err)
	}
	m, err := client.Authorize(envUUID, charmURL.String(), c.Service, c.Plan, httpbakery.OpenWebBrowser)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ms := macaroon.Slice{m}
	return json.Marshal(ms)
}

// Run implements cmd.Command.
func (c *setPlanCommand) Run(ctx *cmd.Context) error {
	credentials, err := c.requestMetricCredentials()
	if err != nil {
		return errors.Trace(err)
	}

	root, err := c.NewAPIRoot()
	if err != nil {
		return errors.Trace(err)
	}
	api := service.NewClient(root)

	err = api.SetMetricCredentials(c.Service, credentials)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
