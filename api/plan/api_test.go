// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package plan_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon.v1"

	api "github.com/juju/romulus/api/plan"
	wireformat "github.com/juju/romulus/wireformat/plan"
)

const (
	testPlan = `
metrics:
  pings:
    unit:
      transform: max
      period: hour
      gaps: zero
`
)

type clientSuite struct {
	httpClient *mockHttpClient

	client api.Client
}

var _ = gc.Suite(&clientSuite{})

func (s *clientSuite) SetUpTest(c *gc.C) {
	s.httpClient = &mockHttpClient{}

	client, err := api.NewClient(api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	s.client = client

}

func (s *clientSuite) TestGet(c *gc.C) {
	plans := []wireformat.Plan{{URL: "bob/uptime", Definition: testPlan}}
	jsonPlans, err := json.Marshal(plans)
	c.Assert(err, jc.ErrorIsNil)

	tests := []struct {
		about   string
		planURL string
		err     string
		status  int
		body    []byte
	}{{
		about:   "not found",
		planURL: "bob/uptime",
		status:  http.StatusNotFound,
		err:     `failed to retrieve associated plans: received http response:  - code "Not Found"`,
	}, {
		about:   "internal server error",
		planURL: "bob/uptime",
		status:  http.StatusInternalServerError,
		err:     `failed to retrieve associated plans: received http response:  - code "Internal Server Error"`,
	}, {
		about:   "wrong response format",
		planURL: "bob/uptime",
		status:  http.StatusOK,
		body:    []byte("wrong response format"),
		err:     `failed to unmarshal response: invalid character 'w' looking for beginning of value`,
	}, {
		about:   "all is well",
		planURL: "bob/uptime",
		status:  http.StatusOK,
		body:    jsonPlans,
	}}

	for _, t := range tests {
		s.httpClient.status = t.status
		s.httpClient.body = t.body
		plans, err := s.client.GetAssociatedPlans(t.planURL)
		if t.err == "" {
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(plans, jc.DeepEquals, plans)
		} else {
			c.Assert(err, gc.ErrorMatches, t.err)
		}
	}
}

func (s *clientSuite) TestAuthorize(c *gc.C) {
	envUUID := utils.MustNewUUID()
	charmURL := "cs:trusty/test-carm-0"
	service := "test-charm"
	plan := utils.MustNewUUID()

	m, err := macaroon.New(nil, "", "")
	c.Assert(err, jc.ErrorIsNil)
	data, err := json.Marshal(m)
	c.Assert(err, jc.ErrorIsNil)

	httpClient := &mockHttpClient{}
	httpClient.status = http.StatusOK
	httpClient.body = data
	authClient, err := api.NewAuthorizationClient(api.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	_, err = authClient.Authorize(envUUID.String(), charmURL, service, plan.String(), nil)
	c.Assert(err, jc.ErrorIsNil)
}

type mockHttpClient struct {
	status int
	body   []byte
}

func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(m.body)),
	}, nil
}

func (m *mockHttpClient) DoWithBody(req *http.Request, body io.ReadSeeker) (*http.Response, error) {
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(m.body)),
	}, nil
}
