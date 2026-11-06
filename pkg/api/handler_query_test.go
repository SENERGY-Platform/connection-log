/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"net/url"
	"testing"
	"time"
)

func TestParseHistoricalStatesQuery(t *testing.T) {
	a := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 11, 6, 6, 0, 0, 0, time.UTC)
	cases := []struct {
		name      string
		query     string
		wantRange time.Duration
		wantSince time.Time
		wantUntil time.Time
		wantErr   bool
	}{
		{name: "nothing", query: ""},
		{name: "range", query: "range=24h", wantRange: 24 * time.Hour},
		{name: "since", query: "since=2026-11-05T00:00:00Z", wantSince: a},
		{name: "until", query: "until=2026-11-06T06:00:00Z", wantUntil: b},
		{name: "since and until", query: "since=2026-11-05T00:00:00Z&until=2026-11-06T06:00:00Z", wantSince: a, wantUntil: b},
		{name: "range and until", query: "range=24h&until=2026-11-06T06:00:00Z", wantRange: 24 * time.Hour, wantUntil: b},
		{name: "invalid range", query: "range=1d", wantErr: true},
		{name: "invalid since", query: "since=2026-11-05", wantErr: true},
		{name: "invalid until", query: "until=yesterday", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			values, err := url.ParseQuery(c.query)
			if err != nil {
				t.Fatal(err)
			}
			rng, since, until, err := parseHistoricalStatesQuery(values)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got range=%v since=%v until=%v", rng, since, until)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if rng != c.wantRange {
				t.Errorf("range: got %v, want %v", rng, c.wantRange)
			}
			if !since.Equal(c.wantSince) {
				t.Errorf("since: got %v, want %v", since, c.wantSince)
			}
			if !until.Equal(c.wantUntil) {
				t.Errorf("until: got %v, want %v", until, c.wantUntil)
			}
		})
	}
}
