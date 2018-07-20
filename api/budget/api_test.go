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
	"github.com/juju/romulus/wireformat/common"
)

type httpErr struct {
	Error string `json:"error"`
}

func Test(t *testing.T) {
	gc.TestingT(t)
}

type TSuite struct{}

var _ = gc.Suite(&TSuite{})

func (t *TSuite) TestCreateWallet(c *gc.C) {
	expected := "Wallet created successfully"
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateWallet("personal", "200")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{
					"limit":  "200",
					"wallet": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateWalletAPIRoot(c *gc.C) {
	expected := "Wallet created successfully"
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient), budget.APIRoot("http://httpbin.org"))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateWallet("personal", "200")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"http://httpbin.org/wallet",
				map[string]interface{}{
					"limit":  "200",
					"wallet": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateWalletServerError(c *gc.C) {
	respBody, err := json.Marshal(httpErr{Error: "wallet already exists"})
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateWallet("personal", "200")
	c.Assert(err, gc.ErrorMatches, "wallet already exists")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{
					"limit":  "200",
					"wallet": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateWalletRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateWallet("personal", "200")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{
					"limit":  "200",
					"wallet": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateWalletUnavail(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusServiceUnavailable,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateWallet("personal", "200")
	c.Assert(common.IsNotAvail(err), jc.IsTrue)
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{
					"limit":  "200",
					"wallet": "personal",
				},
			}}})
}

func (t *TSuite) TestCreateWalletConnRefused(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusOK,
	}
	httpClient.SetErrors(errors.New("Connection refused"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateWallet("personal", "200")
	c.Assert(common.IsNotAvail(err), jc.IsTrue)
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{
					"limit":  "200",
					"wallet": "personal",
				},
			}}})
}

func (t *TSuite) TestListWallets(c *gc.C) {
	expected := &wireformat.ListWalletsResponse{
		Wallets: wireformat.WalletSummaries{
			wireformat.WalletSummary{
				Owner:       "bob",
				Wallet:      "personal",
				Limit:       "50",
				Budgeted:    "30",
				Unallocated: "20",
				Available:   "45",
				Consumed:    "5",
				Default:     true,
			},
			wireformat.WalletSummary{
				Owner:       "bob",
				Wallet:      "work",
				Limit:       "200",
				Budgeted:    "100",
				Unallocated: "100",
				Available:   "150",
				Consumed:    "50",
				Default:     false,
			},
			wireformat.WalletSummary{
				Owner:       "bob",
				Wallet:      "team",
				Limit:       "50",
				Budgeted:    "10",
				Unallocated: "40",
				Available:   "40",
				Consumed:    "10",
				Default:     false,
			},
		},
		Total: wireformat.WalletTotals{
			Limit:       "300",
			Budgeted:    "140",
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
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.ListWallets()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.DeepEquals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestListWalletsServerError(c *gc.C) {
	respBody, err := json.Marshal(httpErr{Error: "wallet already exists"})
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusNotFound,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.ListWallets()
	c.Assert(err, gc.ErrorMatches, "wallet already exists")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestListWalletsRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.ListWallets()
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"",
				"https://api.jujucharms.com/omnibus/v3/wallet",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestSetWallet(c *gc.C) {
	expected := "Wallet updated successfully"
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.SetWallet("personal", "200")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PATCH",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal",
				map[string]interface{}{
					"update": map[string]interface{}{
						"limit": "200",
					},
				},
			}}})
}

func (t *TSuite) TestSetWalletServerError(c *gc.C) {
	respBody, err := json.Marshal(httpErr{Error: "cannot update wallet"})
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.SetWallet("personal", "200")
	c.Assert(err, gc.ErrorMatches, "cannot update wallet")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PATCH",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal",
				map[string]interface{}{
					"update": map[string]interface{}{
						"limit": "200",
					},
				},
			}}})
}

func (t *TSuite) TestSetWalletRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.SetWallet("personal", "200")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PATCH",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal",
				map[string]interface{}{
					"update": map[string]interface{}{
						"limit": "200",
					},
				},
			}}})
}

func (t *TSuite) TestGetWallet(c *gc.C) {
	expected := &wireformat.WalletWithBudgets{
		Limit: "4000.00",
		Total: wireformat.WalletTotals{
			Budgeted:    "2200.00",
			Unallocated: "1800.00",
			Available:   "1100,00",
			Consumed:    "1100.0",
			Usage:       "50%",
		},
		Budgets: []wireformat.Budget{{
			Owner:    "user.joe",
			Limit:    "1200.00",
			Consumed: "500.00",
			Usage:    "42%",
			Model:    "model.joe",
		}, {
			Owner:    "user.jess",
			Limit:    "1000.00",
			Consumed: "600.00",
			Usage:    "60%",
			Model:    "model.jess",
		},
		},
	}
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.GetWallet("personal")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.DeepEquals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestGetWalletServerError(c *gc.C) {
	respBody, err := json.Marshal(httpErr{Error: "wallet not found"})
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusNotFound,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.GetWallet("personal")
	c.Assert(err, gc.ErrorMatches, "wallet not found")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestGetWalletRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.GetWallet("personal")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.IsNil)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"GET",
				"",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestCreateBudget(c *gc.C) {
	expected := "Budget created successfully"
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateBudget("personal", "200", "model")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal/budget",
				map[string]interface{}{
					"limit": "200",
					"model": "model",
				},
			}}})
}

func (t *TSuite) TestCreateBudgetServerError(c *gc.C) {
	respBody, err := json.Marshal(httpErr{Error: "cannot create budget"})
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateBudget("personal", "200", "model")
	c.Assert(err, gc.ErrorMatches, "cannot create budget")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal/budget",
				map[string]interface{}{
					"limit": "200",
					"model": "model",
				},
			}}})
}

func (t *TSuite) TestCreateBudgetRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.CreateBudget("personal", "200", "model")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"POST",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/wallet/personal/budget",
				map[string]interface{}{
					"limit": "200",
					"model": "model",
				},
			}}})
}

func (t *TSuite) TestUpdateBudget(c *gc.C) {
	expected := "Budget updated."
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.UpdateBudget("model-uuid", "personal", "200")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PATCH",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/model/model-uuid/budget",
				map[string]interface{}{
					"update": map[string]interface{}{
						"limit":  "200",
						"wallet": "personal",
					},
				},
			}}})
}

func (t *TSuite) TestUpdateBudgetServerError(c *gc.C) {
	respBody, err := json.Marshal(httpErr{Error: "cannot update budget"})
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.UpdateBudget("model-uuid", "work", "200")
	c.Assert(err, gc.ErrorMatches, "cannot update budget")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PATCH",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/model/model-uuid/budget",
				map[string]interface{}{
					"update": map[string]interface{}{
						"limit":  "200",
						"wallet": "work",
					},
				},
			}}})
}

func (t *TSuite) TestUpdateBudgetRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.UpdateBudget("model-uuid", "", "200")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"PATCH",
				"application/json",
				"https://api.jujucharms.com/omnibus/v3/model/model-uuid/budget",
				map[string]interface{}{
					"update": map[string]interface{}{
						"limit": "200",
					},
				},
			}}})
}

func (t *TSuite) TestDeleteBudget(c *gc.C) {
	expected := "Budget deleted."
	respBody, err := json.Marshal(expected)
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusOK,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.DeleteBudget("model-uuid")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.Equals, expected)
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"DELETE",
				"",
				"https://api.jujucharms.com/omnibus/v3/model/model-uuid/budget",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestDeleteBudgetServerError(c *gc.C) {
	respBody, err := json.Marshal(httpErr{Error: "cannot delete budget"})
	c.Assert(err, jc.ErrorIsNil)
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
		RespBody: respBody,
	}
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.DeleteBudget("model-uuid")
	c.Assert(err, gc.ErrorMatches, "cannot delete budget")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"DELETE",
				"",
				"https://api.jujucharms.com/omnibus/v3/model/model-uuid/budget",
				map[string]interface{}{},
			}}})
}

func (t *TSuite) TestDeleteBudgetRequestError(c *gc.C) {
	httpClient := &mockClient{
		RespCode: http.StatusBadRequest,
	}
	httpClient.SetErrors(errors.New("bogus error"))
	client, err := budget.NewClient(budget.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	response, err := client.DeleteBudget("model-uuid")
	c.Assert(err, gc.ErrorMatches, ".*bogus error")
	c.Assert(response, gc.Equals, "")
	httpClient.CheckCalls(c,
		[]jujutesting.StubCall{{
			"DoWithBody",
			[]interface{}{"DELETE",
				"",
				"https://api.jujucharms.com/omnibus/v3/model/model-uuid/budget",
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
	c.Stub.MethodCall(c, "DoWithBody", req.Method, req.Header.Get("Content-Type"), req.URL.String(), requestData)

	resp := &http.Response{
		StatusCode: c.RespCode,
		Body:       ioutil.NopCloser(bytes.NewReader(c.RespBody)),
	}
	return resp, c.Stub.NextErr()
}
