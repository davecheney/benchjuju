// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package budget_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/juju/errors"
	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/api/budget"
	wireformat "github.com/juju/romulus/wireformat/budget"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type TSuite struct{}

var _ = gc.Suite(&TSuite{})

func (t *TSuite) TestCreateBudget(c *gc.C) {
	expected := "Budget created successfully"
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.CreateBudget("personal", "200")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{
					"limit":  "200",
					"budget": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateBudgetServerError(c *gc.C) {
	respBody, err := json.Marshal("budget already exists")
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.CreateBudget("personal", "200")
	c.Assert(err, gc.ErrorMatches, "400: budget already exists")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{
					"limit":  "200",
					"budget": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateBudgetRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client := budget.NewClient(httpClient)
	response, err := client.CreateBudget("personal", "200")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{
					"limit":  "200",
					"budget": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateBudgetUnavail(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusServiceUnavailable,
	}
	client := budget.NewClient(httpClient)
	response, err := client.CreateBudget("personal", "200")
	c.Assert(wireformat.IsNotAvail(err), jc.IsTrue)
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{
					"limit":  "200",
					"budget": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateBudgetConnRefused(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusOK,
	}
	httpClient.SetErrors(errors.New("Connection refused"))
	client := budget.NewClient(httpClient)
	response, err := client.CreateBudget("personal", "200")
	c.Assert(wireformat.IsNotAvail(err), jc.IsTrue)
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{
					"limit":  "200",
					"budget": "personal",
				},
			}}})
}

func (t *TSuite) TestListBudgets(c *gc.C) {
	expected := &wireformat.ListBudgetsResponse{
		Budgets: wireformat.BudgetSummaries{
			wireformat.BudgetSummary{
				Owner:       "bob",
				Budget:      "personal",
				Limit:       "50",
				Allocated:   "30",
				Unallocated: "20",
				Available:   "45",
				Consumed:    "5",
			},
			wireformat.BudgetSummary{
				Owner:       "bob",
				Budget:      "work",
				Limit:       "200",
				Allocated:   "100",
				Unallocated: "100",
				Available:   "150",
				Consumed:    "50",
			},
			wireformat.BudgetSummary{
				Owner:       "bob",
				Budget:      "team",
				Limit:       "50",
				Allocated:   "10",
				Unallocated: "40",
				Available:   "40",
				Consumed:    "10",
			},
		},
		Total: wireformat.BudgetTotals{
			Limit:       "300",
			Allocated:   "140",
			Available:   "235",
			Unallocated: "160",
			Consumed:    "65",
		},
		Credit: "400",
	}
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.ListBudgets()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.DeepEquals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestListBudgetsServerError(c *gc.C) {
	respBody, err := json.Marshal("budget already exists")
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.ListBudgets()
	c.Assert(err, gc.ErrorMatches, "400: budget already exists")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestListBudgetsRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client := budget.NewClient(httpClient)
	response, err := client.ListBudgets()
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"https://api.jujucharms.com/omnibus/v2/budget",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestSetBudget(c *gc.C) {
	expected := "Budget updated successfully"
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.SetBudget("personal", "200")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PUT",
				"https://api.jujucharms.com/omnibus/v2/budget/personal",
				map[string]interface{}{
					"limit": "200",
				},
			}}})
}

func (t *TSuite) TestSetBudgetServerError(c *gc.C) {
	respBody, err := json.Marshal("cannot update budget")
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.SetBudget("personal", "200")
	c.Assert(err, gc.ErrorMatches, "400: cannot update budget")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PUT",
				"https://api.jujucharms.com/omnibus/v2/budget/personal",
				map[string]interface{}{
					"limit": "200",
				},
			}}})
}

func (t *TSuite) TestSetBudgetRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client := budget.NewClient(httpClient)
	response, err := client.SetBudget("personal", "200")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PUT",
				"https://api.jujucharms.com/omnibus/v2/budget/personal",
				map[string]interface{}{
					"limit": "200",
				},
			}}})
}

func (t *TSuite) TestGetBudget(c *gc.C) {
	expected := &wireformat.BudgetWithAllocations{
		Limit: "4000.00",
		Total: wireformat.BudgetTotals{
			Allocated:   "2200.00",
			Unallocated: "1800.00",
			Available:   "1100,00",
			Consumed:    "1100.0",
			Usage:       "50%",
		},
		Allocations: []wireformat.Allocation{{
			Owner:    "user.joe",
			Limit:    "1200.00",
			Consumed: "500.00",
			Usage:    "42%",
			Model:    "model.joe",
			Services: map[string]wireformat.ServiceAllocation{
				"wordpress": wireformat.ServiceAllocation{
					Consumed: "300.00",
				},
				"mysql": wireformat.ServiceAllocation{
					Consumed: "200.00",
				},
			},
		}, {
			Owner:    "user.jess",
			Limit:    "1000.00",
			Consumed: "600.00",
			Usage:    "60%",
			Model:    "model.jess",
			Services: map[string]wireformat.ServiceAllocation{
				"landscape": wireformat.ServiceAllocation{
					Consumed: "600.00",
				},
			},
		},
		},
	}
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.GetBudget("personal")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.DeepEquals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"https://api.jujucharms.com/omnibus/v2/budget/personal",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestGetBudgetServerError(c *gc.C) {
	respBody, err := json.Marshal("budget not found")
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusNotFound,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.GetBudget("personal")
	c.Assert(err, gc.ErrorMatches, "404: budget not found")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"https://api.jujucharms.com/omnibus/v2/budget/personal",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestGetBudgetRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client := budget.NewClient(httpClient)
	response, err := client.GetBudget("personal")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"https://api.jujucharms.com/omnibus/v2/budget/personal",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestCreateAllocation(c *gc.C) {
	expected := "Allocation created successfully"
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.CreateAllocation("personal", "200", "model", []string{"db"})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget/personal/allocation",
				map[string]interface{}{
					"limit":    "200",
					"model":    "model",
					"services": []interface{}{"db"},
				},
			}}})
}

func (t *TSuite) TestCreateAllocationServerError(c *gc.C) {
	respBody, err := json.Marshal("cannot create allocation")
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.CreateAllocation("personal", "200", "model", []string{"db"})
	c.Assert(err, gc.ErrorMatches, "400: cannot create allocation")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget/personal/allocation",
				map[string]interface{}{
					"limit":    "200",
					"model":    "model",
					"services": []interface{}{"db"},
				},
			}}})
}

func (t *TSuite) TestCreateAllocationRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client := budget.NewClient(httpClient)
	response, err := client.CreateAllocation("personal", "200", "model", []string{"db"})
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"https://api.jujucharms.com/omnibus/v2/budget/personal/allocation",
				map[string]interface{}{
					"limit":    "200",
					"model":    "model",
					"services": []interface{}{"db"},
				},
			}}})
}

func (t *TSuite) TestUpdateAllocation(c *gc.C) {
	expected := "Allocation updated."
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.UpdateAllocation("model", "db", "200")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PUT",
				"https://api.jujucharms.com/omnibus/v2/environment/model/service/db/allocation",
				map[string]interface{}{
					"limit": "200",
				},
			}}})
}

func (t *TSuite) TestUpdateAllocationServerError(c *gc.C) {
	respBody, err := json.Marshal("cannot update allocation")
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.UpdateAllocation("model", "db", "200")
	c.Assert(err, gc.ErrorMatches, "400: cannot update allocation")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PUT",
				"https://api.jujucharms.com/omnibus/v2/environment/model/service/db/allocation",
				map[string]interface{}{
					"limit": "200",
				},
			}}})
}

func (t *TSuite) TestUpdateAllocationRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client := budget.NewClient(httpClient)
	response, err := client.UpdateAllocation("model", "db", "200")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PUT",
				"https://api.jujucharms.com/omnibus/v2/environment/model/service/db/allocation",
				map[string]interface{}{
					"limit": "200",
				},
			}}})
}

func (t *TSuite) TestDeleteAllocation(c *gc.C) {
	expected := "Allocation deleted."
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.DeleteAllocation("model", "db")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"DELETE",
				"https://api.jujucharms.com/omnibus/v2/environment/model/service/db/allocation",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestDeleteAllocationServerError(c *gc.C) {
	respBody, err := json.Marshal("cannot delete allocation")
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client := budget.NewClient(httpClient)
	response, err := client.DeleteAllocation("model", "db")
	c.Assert(err, gc.ErrorMatches, "400: cannot delete allocation")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"DELETE",
				"https://api.jujucharms.com/omnibus/v2/environment/model/service/db/allocation",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestDeleteAllocationRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client := budget.NewClient(httpClient)
	response, err := client.DeleteAllocation("model", "db")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"DELETE",
				"https://api.jujucharms.com/omnibus/v2/environment/model/service/db/allocation",
				map[string]interface{}{},
			}}})
}

type mockClient struct {
	jujutesting.Stub

	RespCode int
	RespBody []byte
}

func (c *mockClient) DoWithBody(req *http.Request, body io.ReadSeeker) (*http.Response, error) {
	requestData := map[string]interface{}{}
	if body != nil {
		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bodyBytes, &requestData)
		if err != nil {
			panic(err)
		}
	}
	c.Stub.MethodCall(c, "DoWithBody", req.Method, req.URL.String(), requestData)

	resp := &http.Response{
		StatusCode: c.RespCode,
		Body:       ioutil.NopCloser(bytes.NewReader(c.RespBody)),
	}
	return resp, c.Stub.NextErr()
}
