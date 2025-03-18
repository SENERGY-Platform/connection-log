package model

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	DeviceKind  = "device"
	GatewayKind = "gateway"
)

type ResourceCurrentState struct {
	ID        string `json:"id"`
	Connected bool   `json:"connected"`
}

type ResourceHistoricalStates struct {
	ID string `json:"id"`
	HistoricalStates
}

type HistoricalStates struct {
	PrevState *State  `json:"prev_state"`
	States    []State `json:"states"`
	NextState *State  `json:"next_state"`
}

type State struct {
	Time      time.Time `json:"time"`
	Connected bool      `json:"connected"`
}

type QueryBase struct {
	Kind string   `json:"kind"`
	IDs  []string `json:"ids"`
}

type QueryCurrent = QueryBase

type QueryHistorical struct {
	QueryBase
	Range Duration  `json:"range"` // Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Since time.Time `json:"since"` // The time must be a quoted string in the RFC 3339 format.
	Until time.Time `json:"until"` // The time must be a quoted string in the RFC 3339 format.
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*d = Duration(time.Duration(val))
		return nil
	case string:
		tmp, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return fmt.Errorf("invalid format: %v", val)
	}
}
