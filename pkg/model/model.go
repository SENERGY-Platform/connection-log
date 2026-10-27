package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/SENERGY-Platform/models/go/models"
)

const (
	DeviceKind          = "device"
	GatewayKind         = "gateway"
	PermDeviceKind      = "devices"
	PermGatewayKind     = "hubs"
	PermDeviceGroupKind = "device-groups"
	PermLocationsKind   = "locations"
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

type HistoricalStatesWithId struct {
	HistoricalStates
	Id string `json:"id"`
}

type State struct {
	Time      time.Time `json:"time"` // Timestamp in RFC 3339 format.
	Connected bool      `json:"connected"`
}

type QueryBase struct {
	IDs []string `json:"ids"` // IDs for witch states are to be retrieved.
}

type QueryWithAttributeFilter struct {
	QueryBase
	DeviceAttributeBlacklist []models.Attribute `json:"device_attribute_blacklist"`
}

type QueryHistorical struct {
	QueryBase
	Range Duration  `json:"range"` // Time range e.g. 24h, valid units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Since time.Time `json:"since"` // Timestamp in RFC 3339 format, can be combined with 'range' or 'until'.
	Until time.Time `json:"until"` // Timestamp in RFC 3339 format, can be combined with 'range' or 'since'.
}

type OfflineSinceResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name,omitempty"`
	OfflineSince time.Time `json:"offline_since"`
}

// GatewayHealth is the latest health report of a gateway (MGW) as received by the platform.
type GatewayHealth struct {
	HubID      string          `json:"hub_id"`
	State      int             `json:"state"`                       // 0 undetermined, 1 healthy, 2 unhealthy.
	Reasons    []string        `json:"reasons"`                     // Short English descriptions of what is unhealthy, empty when healthy.
	Report     json.RawMessage `json:"report" swaggertype:"object"` // The report as sent by the gateway.
	ReportTime string          `json:"report_time"`                 // Time of the snapshot on the gateway, as sent by it (RFC 3339).
	Received   time.Time       `json:"received"`                    // Time the platform received the report (RFC 3339).
	Stale      bool            `json:"stale"`                       // True if the report was received longer ago than the configured stale age.
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
