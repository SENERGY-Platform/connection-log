package model

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	DeviceKind      = "device"
	GatewayKind     = "gateway"
	PermDeviceKind  = "devices"
	PermGatewayKind = "hubs"
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
	PrevState *State  `json:"prev_state"` // Last state preceding the selected time frame.
	States    []State `json:"states"`     // All states within the selected time frame.
	NextState *State  `json:"next_state"` // First state succeeding the selected time frame.
}

type State struct {
	Time      time.Time `json:"time"` // Timestamp in RFC 3339 format.
	Connected bool      `json:"connected"`
}

type QueryBase struct {
	IDs []string `json:"ids"` // IDs for witch states are to be retrieved.
}

type QueryHistorical struct {
	QueryBase
	Range Duration  `json:"range"` // Time range e.g. 24h, valid units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Since time.Time `json:"since"` // Timestamp in RFC 3339 format, can be combined with 'range' or 'until'.
	Until time.Time `json:"until"` // Timestamp in RFC 3339 format, can be combined with 'range' or 'since'.
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
