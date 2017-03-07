// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package sla contains the sla service API client.
package sla

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/juju/errors"
	"github.com/juju/romulus/wireformat/sla"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"gopkg.in/macaroon.v1"
)

var DefaultURL = "https://api.jujucharms.com/omnibus/v2"

// AuthClient defines the interface available to clients of the support api.
type AuthClient interface {
	// Authorize returns the sla macaroon for the specified model
	Authorize(modelUUID, supportLevel, budget string) (*macaroon.Macaroon, error)
}

var _ AuthClient = (*client)(nil)

type httpClient interface {
	DoWithBody(req *http.Request, body io.ReadSeeker) (*http.Response, error)
}

// client is the implementation of the Client interface.
type client struct {
	client  httpClient
	baseURL string
}

// ClientOption defines a function which configures a Client.
type ClientOption func(h *client) error

// HTTPClient returns a function that sets the http client used by the API
// (e.g. if we want to use TLS).
func HTTPClient(c httpClient) func(h *client) error {
	return func(h *client) error {
		h.client = c
		return nil
	}
}

// BaseURL sets the base url for the api client.
func BaseURL(url string) func(h *client) error {
	return func(h *client) error {
		h.baseURL = url
		return nil
	}
}

// NewClient returns a new client for the sla api.
func NewClient(options ...ClientOption) (*client, error) {
	c := &client{
		client:  httpbakery.NewClient(),
		baseURL: DefaultURL,
	}

	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return c, nil
}

// Authorize obtains a sla authorization macaroon.
func (c *client) Authorize(modelUUID, supportLevel, budget string) (*macaroon.Macaroon, error) {
	u, err := url.Parse(c.baseURL + "/sla/authorize")
	if err != nil {
		return nil, errors.Trace(err)
	}

	slaRequest := sla.SLARequest{
		ModelUUID: modelUUID,
		Level:     supportLevel,
		Budget:    budget,
	}

	buff := &bytes.Buffer{}
	encoder := json.NewEncoder(buff)
	err = encoder.Encode(slaRequest)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.DoWithBody(req, bytes.NewReader(buff.Bytes()))
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer discardClose(response)

	if response.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			return nil, errors.Errorf("failed to authorize sla: received http response: %v - code %q", string(body), http.StatusText(response.StatusCode))
		}
		return nil, errors.Errorf("failed to authorize sla: http response is %q", http.StatusText(response.StatusCode))
	}

	var m *macaroon.Macaroon
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&m)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal the response")
	}

	return m, nil
}

// discardClose reads any remaining data from the response body and closes it.
func discardClose(response *http.Response) {
	if response == nil || response.Body == nil {
		return
	}
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()
}
