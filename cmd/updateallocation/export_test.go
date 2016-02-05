// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package updateallocation

import (
	"gopkg.in/macaroon-bakery.v1/httpbakery"
)

var (
	NewAPIClient = &newAPIClient
)

// APIClientFnc allow patching of the apiClient
func APIClientFnc(api apiClient) func(*httpbakery.Client) (apiClient, error) {
	return func(_ *httpbakery.Client) (apiClient, error) {
		return api, nil
	}
}
