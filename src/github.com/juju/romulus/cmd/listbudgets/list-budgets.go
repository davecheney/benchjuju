// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package listbudgets

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

// NewListBudgetsCommand returns a new command that is used
// to list budgets a user has access to.
func NewListBudgetsCommand() cmd.Command {
	return &listBudgetsCommand{}
}

type listBudgetsCommand struct {
	rcmd.HttpCommand

	out cmd.Output
}

const listBudgetsDoc = `
List the available budgets.

Example:
 juju list-budgets
`

// Info implements cmd.Command.Info.
func (c *listBudgetsCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "list-budgets",
		Purpose: "list budgets",
		Doc:     listBudgetsDoc,
	}
}

// SetFlags implements cmd.Command.SetFlags.
func (c *listBudgetsCommand) SetFlags(f *gnuflag.FlagSet) {
	c.out.AddFlags(f, "tabular", map[string]cmd.Formatter{
		"tabular": formatTabular,
	})
}

func (c *listBudgetsCommand) Run(ctx *cmd.Context) error {
	defer c.Close()
	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	api, err := newAPIClient(client)
	if err != nil {
		return errors.Annotate(err, "failed to create an api client")
	}
	budgets, err := api.ListBudgets()
	if err != nil {
		return errors.Annotate(err, "failed to retrieve budgets")
	}
	if budgets == nil {
		return errors.New("no budget information available")
	}
	err = c.out.Write(ctx, budgets)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// formatTabular returns a tabular view of available budgets.
func formatTabular(value interface{}) ([]byte, error) {
	b, ok := value.(*wireformat.ListBudgetsResponse)
	if !ok {
		return nil, errors.Errorf("expected value of type %T, got %T", b, value)
	}
	sort.Sort(b.Budgets)

	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true

	table.AddRow("BUDGET", "MONTHLY", "ALLOCATED", "AVAILABLE", "SPENT")
	for _, budgetEntry := range b.Budgets {
		table.AddRow(budgetEntry.Budget, budgetEntry.Limit, budgetEntry.Allocated, budgetEntry.Available, budgetEntry.Consumed)
	}
	table.AddRow("TOTAL", b.Total.Limit, b.Total.Allocated, b.Total.Available, b.Total.Consumed)
	table.AddRow("", "", "", "", "")
	table.AddRow("Credit limit:", b.Credit, "", "", "")
	return []byte(table.String()), nil
}

var newAPIClient = newAPIClientImpl

func newAPIClientImpl(c *httpbakery.Client) (apiClient, error) {
	client := api.NewClient(c)
	return client, nil
}

type apiClient interface {
	// ListBudgets returns a list of budgets a user has access to.
	ListBudgets() (*wireformat.ListBudgetsResponse, error)
}
