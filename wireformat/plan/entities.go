// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// The plan package contains wireformat structs intended for
// the plan management API.
package plan

import (
	"regexp"

	"github.com/juju/errors"
	"github.com/juju/utils/v3"
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
	EnvironmentUUID string `json:"env-uuid"` // TODO(cmars): rename to EnvUUID
	CharmURL        string `json:"charm-url"`
	ServiceName     string `json:"service-name"`
	PlanURL         string `json:"plan-url"`
}

const (
	applicationSnippet   = "(?:[a-z][a-z0-9]*(?:-[a-z0-9]*[a-z][a-z0-9]*)*)"
	validUserNameSnippet = "[a-zA-Z0-9][a-zA-Z0-9.+-]*[a-zA-Z0-9]"
	seriesSnippet        = "[a-z]+([a-z0-9]+)?"
	charmNameSnippet     = "[a-z][a-z0-9]*(-[a-z0-9]*[a-z][a-z0-9]*)*"
	localSchemaSnippet   = "local:"
	revisionSnippet      = "(-1|0|[1-9][0-9]*)"
)

var (
	validApplication          = regexp.MustCompile("^" + applicationSnippet + "$")
	v1CharmStoreSchemaSnippet = "cs:(~" + validUserNameSnippet + "/)?"
	validV1CharmRegEx         = regexp.MustCompile("^(" +
		localSchemaSnippet + "|" +
		v1CharmStoreSchemaSnippet + ")?(" +
		seriesSnippet + "/)?" +
		charmNameSnippet + "(-" +
		revisionSnippet + ")?$")
	v3CharmStoreSchemaSnippet = "(cs:)?(" + validUserNameSnippet + "/)?"
	validV3CharmRegEx         = regexp.MustCompile("^(" +
		localSchemaSnippet + "|" +
		v3CharmStoreSchemaSnippet + ")" +
		charmNameSnippet + "(/" +
		seriesSnippet + ")?(/" +
		revisionSnippet + ")?$")
)

func isValidCharm(url string) bool {
	return validV1CharmRegEx.MatchString(url) || validV3CharmRegEx.MatchString(url)
}

// Validate checks the AuthorizationRequest for errors.
func (s AuthorizationRequest) Validate() error {
	if !utils.IsValidUUIDString(s.EnvironmentUUID) {
		return errors.Errorf("invalid environment UUID: %q", s.EnvironmentUUID)
	}
	if s.ServiceName == "" {
		return errors.New("undefined service name")
	}
	if !validApplication.MatchString(s.ServiceName) {
		return errors.Errorf("invalid service name: %q", s.ServiceName)
	}
	if s.CharmURL == "" {
		return errors.New("undefined charm url")
	}
	if !isValidCharm(s.CharmURL) {
		return errors.Errorf("invalid charm url: %q", s.CharmURL)
	}
	if s.PlanURL == "" {
		return errors.Errorf("undefined plan url")
	}
	return nil
}
