// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// The plan package contains wireformat structs intended for
// the plan management API.
package plan

import (
	"encoding/json"

	"github.com/juju/errors"
	"github.com/juju/utils"
	"gopkg.in/juju/names.v2"
)

// Plan structure is used as a wire format to store information on ISV-created
// rating plan and charm URLs for which the plan is valid (a subscription
// using this plan can be created).
type Plan struct {
	URL        string `json:"url"`        // Name of the rating plan
	Definition string `json:"plan"`       // The rating plan
	CreatedOn  string `json:"created-on"` // When the plan was created - RFC3339 encoded timestamp
}

// AuthorizationRequest defines the struct used to request a plan authorization.
type AuthorizationRequest struct {
	EnvironmentUUID string `json:"env-uuid"`
	CharmURL        string `json:"charm-url"`
	ServiceName     string `json:"service-name"`
	PlanURL         string `json:"plan-url"`
}

type authorizationRequestV1 AuthorizationRequest

// UnmarshalJSON implements a transitional json.Unmarshaler to allow
// forward-compatible processing of fields renamed in Juju 2.0.
func (ar *AuthorizationRequest) UnmarshalJSON(data []byte) error {
	v := struct {
		authorizationRequestV1
		ModelUUID       string `json:"model-uuid"`
		ApplicationName string `json:"application-name"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*ar = AuthorizationRequest(v.authorizationRequestV1)
	if ar.EnvironmentUUID == "" {
		ar.EnvironmentUUID = v.ModelUUID
	}
	if ar.ServiceName == "" {
		ar.ServiceName = v.ApplicationName
	}
	return nil
}

// Validate checks the AuthorizationRequest for errors.
func (s AuthorizationRequest) Validate() error {
	if !utils.IsValidUUIDString(s.EnvironmentUUID) {
		return errors.Errorf("invalid environment UUID: %q", s.EnvironmentUUID)
	}
	if s.ServiceName == "" {
		return errors.New("undefined service name")
	}
	if !names.IsValidApplication(s.ServiceName) {
		return errors.Errorf("invalid service name: %q", s.ServiceName)
	}
	if s.CharmURL == "" {
		return errors.New("undefined charm url")
	}
	if !names.IsValidCharm(s.CharmURL) {
		return errors.Errorf("invalid charm url: %q", s.CharmURL)
	}
	if s.PlanURL == "" {
		return errors.Errorf("undefined plan url")
	}
	return nil
}
