// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package terms defines the terms service API.
package terms_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	stdtesting "testing"
	"time"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/romulus/api/terms"
)

type apiSuite struct {
	client     terms.Client
	httpClient *mockHttpClient
}

func Test(t *stdtesting.T) {
	gc.TestingT(t)
}

var _ = gc.Suite(&apiSuite{})

func (s *apiSuite) SetUpTest(c *gc.C) {
	s.httpClient = &mockHttpClient{}
	var err error
	s.client, err = terms.NewClient(terms.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
}

func (s *apiSuite) TestSignedAgreements(c *gc.C) {
	t := time.Now().UTC()
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []terms.AgreementResponse{
		{
			User:      "test-user",
			Term:      "hello-world-terms",
			Revision:  1,
			CreatedOn: t,
		},
		{
			User:      "test-user",
			Term:      "hello-universe-terms",
			Revision:  42,
			CreatedOn: t,
		},
	})
	signedAgreements, err := s.client.GetUsersAgreements()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(signedAgreements, gc.HasLen, 2)
	c.Assert(signedAgreements[0].User, gc.Equals, "test-user")
	c.Assert(signedAgreements[0].Term, gc.Equals, "hello-world-terms")
	c.Assert(signedAgreements[0].Revision, gc.Equals, 1)
	c.Assert(signedAgreements[0].CreatedOn, gc.DeepEquals, t)
	c.Assert(signedAgreements[1].User, gc.Equals, "test-user")
	c.Assert(signedAgreements[1].Term, gc.Equals, "hello-universe-terms")
	c.Assert(signedAgreements[1].Revision, gc.Equals, 42)
	c.Assert(signedAgreements[1].CreatedOn, gc.DeepEquals, t)
}

func (s *apiSuite) TestUnsignedTerms(c *gc.C) {
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []terms.GetTermsResponse{
		{
			Name:     "hello-world-terms",
			Revision: 1,
			Content:  "terms doc content",
		},
		{
			Name:     "hello-universe-terms",
			Revision: 1,
			Content:  "universal terms doc content",
		},
	})
	missingAgreements, err := s.client.GetUnsignedTerms(&terms.CheckAgreementsRequest{
		Terms: []string{
			"hello-world-terms/1",
			"hello-universe-terms/1",
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(missingAgreements, gc.HasLen, 2)
	c.Assert(missingAgreements[0].Name, gc.Equals, "hello-world-terms")
	c.Assert(missingAgreements[0].Revision, gc.Equals, 1)
	c.Assert(missingAgreements[0].Content, gc.Equals, "terms doc content")
	c.Assert(missingAgreements[1].Name, gc.Equals, "hello-universe-terms")
	c.Assert(missingAgreements[1].Revision, gc.Equals, 1)
	c.Assert(missingAgreements[1].Content, gc.Equals, "universal terms doc content")
	s.httpClient.SetBody(c, terms.SaveAgreementResponses{
		Agreements: []terms.AgreementResponse{{
			User:     "test-user",
			Term:     "hello-world-terms",
			Revision: 1,
		}}})

	p1 := &terms.SaveAgreements{
		Agreements: []terms.SaveAgreement{{
			TermName:     "hello-world-terms",
			TermRevision: 1,
		}}}
	response, err := s.client.SaveAgreement(p1)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response.Agreements, gc.HasLen, 1)
	c.Assert(response.Agreements[0].User, gc.Equals, "test-user")
	c.Assert(response.Agreements[0].Term, gc.Equals, "hello-world-terms")
	c.Assert(response.Agreements[0].Revision, gc.Equals, 1)
}

func (s *apiSuite) TestNoFoundReturnsError(c *gc.C) {
	s.httpClient.status = http.StatusNotFound
	s.httpClient.body = []byte("something failed")
	_, err := s.client.GetUnsignedTerms(&terms.CheckAgreementsRequest{
		Terms: []string{
			"hello-world-terms/1",
			"hello-universe-terms/1",
		},
	})
	c.Assert(err, gc.ErrorMatches, "failed to get unsigned terms: Not Found: something failed")
}

func (s *apiSuite) TestSignedAgreementsEnvTermsURL(c *gc.C) {
	cleanup := testing.PatchEnvironment("JUJU_TERMS", "http://example.com")
	defer cleanup()

	t := time.Now().UTC()
	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []terms.AgreementResponse{
		{
			User:      "test-user",
			Term:      "hello-world-terms",
			Revision:  1,
			CreatedOn: t,
		},
		{
			User:      "test-user",
			Term:      "hello-universe-terms",
			Revision:  42,
			CreatedOn: t,
		},
	})
	_, err := s.client.GetUsersAgreements()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.httpClient.Calls(), gc.HasLen, 1)
	s.httpClient.CheckCall(c, 0, "Do", "http://example.com/agreements")
}

func (s *apiSuite) TestUnsignedTermsEnvTermsURL(c *gc.C) {
	cleanup := testing.PatchEnvironment("JUJU_TERMS", "http://example.com")
	defer cleanup()

	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, []terms.GetTermsResponse{
		{
			Name:     "hello-world-terms",
			Revision: 1,
			Content:  "terms doc content",
		},
		{
			Name:     "hello-universe-terms",
			Revision: 1,
			Content:  "universal terms doc content",
		},
	})
	_, err := s.client.GetUnsignedTerms(&terms.CheckAgreementsRequest{
		Terms: []string{
			"hello-world-terms/1",
			"hello-universe-terms/1",
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.httpClient.Calls(), gc.HasLen, 1)
	s.httpClient.CheckCall(c, 0, "Do", "http://example.com/agreement?Terms=hello-world-terms%2F1&Terms=hello-universe-terms%2F1")
	s.httpClient.ResetCalls()
}

func (s *apiSuite) TestSaveAgreementEnvTermsURL(c *gc.C) {
	cleanup := testing.PatchEnvironment("JUJU_TERMS", "http://example.com")
	defer cleanup()

	s.httpClient.status = http.StatusOK
	s.httpClient.SetBody(c, terms.SaveAgreementResponses{})
	p1 := &terms.SaveAgreements{
		Agreements: []terms.SaveAgreement{{
			TermName:     "hello-world-terms",
			TermRevision: 1,
		}}}
	_, err := s.client.SaveAgreement(p1)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.httpClient.Calls(), gc.HasLen, 1)
	s.httpClient.CheckCall(c, 0, "DoWithBody", "http://example.com/agreement")
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

func (m *mockHttpClient) DoWithBody(req *http.Request, body io.ReadSeeker) (*http.Response, error) {
	m.AddCall("DoWithBody", req.URL.String())
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(m.body)),
	}, nil
}

func (m *mockHttpClient) SetBody(c *gc.C, v interface{}) {
	b, err := json.Marshal(&v)
	c.Assert(err, jc.ErrorIsNil)
	m.body = b
}
