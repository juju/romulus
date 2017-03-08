// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// The sla package implements wireformats for the sla service.
package sla

// SLARequest defines the json used to post to sla service.
type SLARequest struct {
	ModelUUID string `json:"model"`
	Level     string `json:"sla"`
	Budget    string `json:"budget"`
}
