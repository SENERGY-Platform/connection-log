package model

import "time"

type States struct {
	ResourceID string  `json:"resource_id"`
	InitState  *State  `json:"init_state"`
	PrevState  *State  `json:"prev_state"`
	States     []State `json:"states"`
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
	TimeRange time.Duration `json:"time_range"`
}
