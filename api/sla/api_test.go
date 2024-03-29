// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package sla_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/v3"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon.v2"

	api "github.com/juju/romulus/api/sla"
	"github.com/juju/romulus/wireformat/sla"
)

type clientSuite struct {
	httpClient *mockHttpClient

	client api.AuthClient
}

var _ = gc.Suite(&clientSuite{})

func (s *clientSuite) SetUpTest(c *gc.C) {
	s.httpClient = &mockHttpClient{}

	client, err := api.NewClient(api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	s.client = client

}

func (s *clientSuite) TestAuthorize(c *gc.C) {
	modelUUID := utils.MustNewUUID()
	level := "essential"

	m, err := macaroon.New(nil, nil, "", macaroon.LatestVersion)
	c.Assert(err, jc.ErrorIsNil)
	data, err := json.Marshal(sla.SLAResponse{
		Owner:       "bob",
		Credentials: m,
		Message:     "info",
	})
	c.Assert(err, jc.ErrorIsNil)

	httpClient := &mockHttpClient{}
	httpClient.status = http.StatusOK
	httpClient.body = data
	authClient, err := api.NewClient(api.HTTPClient(httpClient))
	c.Assert(err, jc.ErrorIsNil)
	resp, err := authClient.Authorize(modelUUID.String(), level, "")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(resp.Owner, gc.Equals, "bob")
	c.Assert(resp.Message, gc.Equals, "info")
	c.Assert(resp.Credentials.Signature(), jc.DeepEquals, m.Signature())

}

type mockHttpClient struct {
	testing.Stub

	status int
	body   []byte
}

func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	m.AddCall("Do", req.URL.String())
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(m.body)),
	}, nil
}
