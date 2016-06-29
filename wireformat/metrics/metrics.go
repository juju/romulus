// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package metrics defines the format that will be used to send metric
// batches to the collector and receive updates.
package metrics

import (
	"encoding/json"
	"time"
)

// MetricBatch is a batch of metrics that will be sent to
// the metric collector
type MetricBatch struct {
	UUID        string    `json:"uuid"`
	ModelUUID   string    `json:"env-uuid"`
	UnitName    string    `json:"unit-name"`
	CharmUrl    string    `json:"charm-url"`
	Created     time.Time `json:"created"`
	Metrics     []Metric  `json:"metrics"`
	Credentials []byte    `json:"credentials"`
}

type metricBatchV1 MetricBatch

// UnmarshalJSON implements a transitional json.Unmarshaler to allow
// forward-compatible processing of fields renamed in Juju 2.0.
func (mb *MetricBatch) UnmarshalJSON(data []byte) error {
	v := struct {
		metricBatchV1
		ModelUUID string `json:"model-uuid"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*mb = MetricBatch(v.metricBatchV1)
	if mb.ModelUUID == "" {
		mb.ModelUUID = v.ModelUUID
	}
	return nil
}

// Metric represents a single Metric.
type Metric struct {
	Key   string    `json:"key"`
	Value string    `json:"value"`
	Time  time.Time `json:"time"`
}

// Response represents the response from the metrics collector.
type Response struct {
	UUID           string               `json:"uuid"`
	EnvResponses   EnvironmentResponses `json:"env-responses"`
	NewGracePeriod time.Duration        `json:"new-grace-period"`
}

type responseV1 Response

// UnmarshalJSON implements a transitional json.Unmarshaler to allow
// forward-compatible processing of fields renamed in Juju 2.0.
func (r *Response) UnmarshalJSON(data []byte) error {
	v := struct {
		responseV1
		ModelResponses EnvironmentResponses `json:"model-responses"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*r = Response(v.responseV1)
	if r.EnvResponses == nil {
		r.EnvResponses = v.ModelResponses
	}
	return nil
}

type EnvironmentResponses map[string]EnvResponse

// Ack adds the specified the batch UUID to the list of acknowledged batches
// for the specified environment.
func (e EnvironmentResponses) Ack(modelUUID, batchUUID string) {
	env := e[modelUUID]

	env.AcknowledgedBatches = append(env.AcknowledgedBatches, batchUUID)
	e[modelUUID] = env
}

func (e EnvironmentResponses) SetStatus(modelUUID, unitName, status, info string) {
	s := UnitStatus{
		Status: status,
		Info:   info,
	}

	env := e[modelUUID]

	if env.UnitStatuses == nil {
		env.UnitStatuses = map[string]UnitStatus{
			unitName: s,
		}
	} else {
		env.UnitStatuses[unitName] = s
	}
	e[modelUUID] = env
}

// EnvResponse contains the response data relevant to a concrete environment.
type EnvResponse struct {
	AcknowledgedBatches []string              `json:"acks,omitempty"`
	UnitStatuses        map[string]UnitStatus `json:"unit-statuses,omitempty"`
}

type UnitStatus struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}
