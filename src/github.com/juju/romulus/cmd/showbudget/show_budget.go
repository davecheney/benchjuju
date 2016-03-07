// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package showbudget

import (
	"sort"

	"github.com/gosuri/uitable"
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"launchpad.net/gnuflag"

	api "github.com/juju/romulus/api/budget"
	rcmd "github.com/juju/romulus/cmd"
	wireformat "github.com/juju/romulus/wireformat/budget"
)

// NewShowBudgetCommand returns a new command that is used
// to show details of the specified wireformat.
func NewShowBudgetCommand() cmd.Command {
	return &showBudgetCommand{}
}

type showBudgetCommand struct {
	rcmd.HttpCommand

	out    cmd.Output
	budget string
}

const showBudgetDoc = `
Display budget usage information.

Example:
 juju show-budget personal
`

// Info implements cmd.Command.Info.
func (c *showBudgetCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "show-budget",
		Purpose: "show budget usage",
		Doc:     showBudgetDoc,
	}
}

// Init implements cmd.Command.Init.
func (c *showBudgetCommand) Init(args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}
	c.budget, args = args[0], args[1:]

	return cmd.CheckEmpty(args)
}

// SetFlags implements cmd.Command.SetFlags.
func (c *showBudgetCommand) SetFlags(f *gnuflag.FlagSet) {
	c.out.AddFlags(f, "tabular", map[string]cmd.Formatter{
		"tabular": formatTabular,
	})
}

func (c *showBudgetCommand) Run(ctx *cmd.Context) error {
	defer c.Close()
	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	api, err := newAPIClient(client)
	if err != nil {
		return errors.Annotate(err, "failed to create an api client")
	}
	budget, err := api.GetBudget(c.budget)
	if err != nil {
		return errors.Annotate(err, "failed to retrieve the budget")
	}
	err = c.out.Write(ctx, budget)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// formatTabular returns a tabular view of available budgets.
func formatTabular(value interface{}) ([]byte, error) {
	b, ok := value.(*wireformat.BudgetWithAllocations)
	if !ok {
		return nil, errors.Errorf("expected value of type %T, got %T", b, value)
	}

	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true

	table.AddRow("MODEL", "SERVICES", "SPENT", "ALLOCATED BY", "USAGE")
	for _, allocation := range b.Allocations {
		firstLine := true
		// We'll sort the service names to avoid nondeterministic
		// command output.
		services := make([]string, 0, len(allocation.Services))
		for serviceName, _ := range allocation.Services {
			services = append(services, serviceName)
		}
		sort.Strings(services)
		for _, serviceName := range services {
			service, _ := allocation.Services[serviceName]
			if firstLine {
				table.AddRow(allocation.Model, serviceName, service.Consumed, allocation.Owner, allocation.Usage)
				firstLine = false
				continue
			}
			table.AddRow("", serviceName, service.Consumed, "", "")
		}

	}
	table.AddRow("", "", "", "", "")
	table.AddRow("TOTAL", "", b.Total.Consumed, b.Total.Allocated, b.Total.Usage)
	table.AddRow("BUDGET", "", "", b.Limit, "")
	table.AddRow("UNALLOCATED", "", "", b.Total.Unallocated, "")
	return []byte(table.String()), nil
}

var newAPIClient = newAPIClientImpl

func newAPIClientImpl(c *httpbakery.Client) (apiClient, error) {
	client := api.NewClient(c)
	return client, nil
}

type apiClient interface {
	GetBudget(string) (*wireformat.BudgetWithAllocations, error)
}
