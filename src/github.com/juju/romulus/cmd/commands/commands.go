// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package commands provides functionality for registering all the romulus commands.
package commands

import (
	"github.com/juju/cmd"

	"github.com/juju/romulus/cmd/agree"
	"github.com/juju/romulus/cmd/allocate"
	"github.com/juju/romulus/cmd/createbudget"
	"github.com/juju/romulus/cmd/listbudgets"
	"github.com/juju/romulus/cmd/listplans"
	"github.com/juju/romulus/cmd/setbudget"
	"github.com/juju/romulus/cmd/setplan"
	"github.com/juju/romulus/cmd/showbudget"
	"github.com/juju/romulus/cmd/updateallocation"
)

type commandRegister interface {
	Register(cmd.Command)
}

// RegisterAll registers all romulus commands with the
// provided command registry.
func RegisterAll(r commandRegister) {
	r.Register(agree.NewAgreeCommand())
	r.Register(allocate.NewAllocateCommand())
	r.Register(createbudget.NewCreateBudgetCommand())
	r.Register(listbudgets.NewListBudgetsCommand())
	r.Register(listplans.NewListPlansCommand())
	r.Register(setbudget.NewSetBudgetCommand())
	r.Register(setplan.NewSetPlanCommand())
	r.Register(showbudget.NewShowBudgetCommand())
	r.Register(updateallocation.NewUpdateAllocationCommand())
}
