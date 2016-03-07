// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package budget

import (
	"fmt"
	"net/http"
)

var baseURL = "https://api.jujucharms.com/omnibus/v2"

// CreateBudgetRequest is used in the requests to the budget service
// for creating the specified budget.
type CreateBudgetRequest struct {
	Budget string `json:"budget"`
	Limit  string `json:"limit"`
}

// Method returns the http method used for this request.
func (CreateBudgetRequest) Method() string { return "POST" }

// Body returns the body of the request.
func (c CreateBudgetRequest) Body() interface{} {
	return c
}

// URL returns the URL of the request.
func (CreateBudgetRequest) URL() string {
	return fmt.Sprintf("%s/budget", baseURL)
}

// ListBudgetsRequest defines a request to the budgets service
// to list a user's budgets.
type ListBudgetsRequest struct{}

// Method returns the method of the request.
func (ListBudgetsRequest) Method() string { return "GET" }

// URL returns the URL of the request.
func (ListBudgetsRequest) URL() string {
	return fmt.Sprintf("%s/budget", baseURL)
}

// SetBudgetRequest defines a request that updates the limit of
// a budget.
type SetBudgetRequest struct {
	Budget string `json:"-"`
	Limit  string `json:"limit"`
}

// Method returns the method of the request.
func (SetBudgetRequest) Method() string { return "PUT" }

// Body returns the request body.
func (r SetBudgetRequest) Body() interface{} { return r }

// URL returns the URL for the request.
func (r SetBudgetRequest) URL() string {
	return fmt.Sprintf("%s/budget/%s", baseURL, r.Budget)
}

// GetBudgetRequest defines a request that retrieves a specific budget.
type GetBudgetRequest struct {
	Budget string
}

// URL returns the URL for the request.
func (r GetBudgetRequest) URL() string {
	return fmt.Sprintf("%s/budget/%s", baseURL, r.Budget)
}

// Method returns the method for the request.
func (GetBudgetRequest) Method() string { return "GET" }

// CreateAllocationRequest defines a request to create an allocation in the specified budget.
type CreateAllocationRequest struct {
	Model    string   `json:"model"`
	Services []string `json:"services"`
	Limit    string   `json:"limit"`
	Budget   string   `json:"-"`
}

// URL returns the URL for the request.
func (r CreateAllocationRequest) URL() string {
	return fmt.Sprintf("%s/budget/%s/allocation", baseURL, r.Budget)
}

// Method returns the method for the request.
func (CreateAllocationRequest) Method() string { return "POST" }

// Body returns the request body.
func (r CreateAllocationRequest) Body() interface{} { return r }

// UpdateAllocationRequest defines a request to update an allocation
// associated with a service.
type UpdateAllocationRequest struct {
	Model   string `json:"-"`
	Service string `json:"-"`
	Limit   string `json:"limit"`
}

// URL returns the URL for the request.
func (r UpdateAllocationRequest) URL() string {
	return fmt.Sprintf("%s/environment/%s/service/%s/allocation", baseURL, r.Model, r.Service)
}

// Method returns the method for the request.
func (UpdateAllocationRequest) Method() string { return "PUT" }

// Body returns the request body.
func (r UpdateAllocationRequest) Body() interface{} { return r }

// DeleteAllocationRequwest defines a request that removes an allocation associated
// with a service.
type DeleteAllocationRequest struct {
	Model   string `json:"-"`
	Service string `json:"-"`
}

// URL returns the URL for the request.
func (r DeleteAllocationRequest) URL() string {
	return fmt.Sprintf("%s/environment/%s/service/%s/allocation", baseURL, r.Model, r.Service)
}

// Method returns the method for the request.
func (DeleteAllocationRequest) Method() string { return "DELETE" }

// HttpError represents an error caused by a failed http request.
type HttpError struct {
	StatusCode int
	Message    string
}

func (e HttpError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

// NotAvailError indicates that the service is either unreachable or unavailable.
type NotAvailError struct {
	Resp int
}

func (e NotAvailError) Error() string {
	if e.Resp == http.StatusServiceUnavailable {
		return "service unavailable"
	} else {
		return "service unreachable"
	}
}

// IsNotAvail indicates whether the error is a NotAvailError.
func IsNotAvail(err error) bool {
	_, ok := err.(NotAvailError)
	return ok
}
