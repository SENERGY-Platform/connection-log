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

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestIsStale(t *testing.T) {
	now := time.Date(2026, 10, 27, 14, 0, 0, 0, time.UTC)
	age := 15 * time.Minute
	cases := []struct {
		name     string
		received time.Time
		want     bool
	}{
		{"just received", now, false},
		{"within age", now.Add(-14 * time.Minute), false},
		{"exactly at age", now.Add(-age), false},
		{"one nanosecond past age", now.Add(-age - time.Nanosecond), true},
		{"long ago", now.Add(-24 * time.Hour), true},
		{"in the future", now.Add(time.Hour), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isStale(c.received, now, age); got != c.want {
				t.Errorf("isStale() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestGatewayHealthStaleAgeConfig(t *testing.T) {
	cases := []struct {
		value   string
		want    time.Duration
		wantErr bool
	}{
		{"", 15 * time.Minute, false},
		{"15m", 15 * time.Minute, false},
		{"90s", 90 * time.Second, false},
		{"0s", 0, true},
		{"-5m", 0, true},
		{"soon", 0, true},
		{"15", 0, true},
	}
	for _, c := range cases {
		t.Run(c.value, func(t *testing.T) {
			conf := configuration.Config{MongoDatabase: "d", GatewayHealthStaleAge: c.value}
			got, err := gatewayHealthStaleAge(conf)
			if c.wantErr {
				if !errors.Is(err, errInvalidStaleAge) {
					t.Fatalf("err = %v, want errInvalidStaleAge", err)
				}
				if !errors.Is(validateConfig(conf), errInvalidStaleAge) {
					t.Errorf("validateConfig() did not reject %q", c.value)
				}
				return
			}
			if err != nil || got != c.want {
				t.Errorf("gatewayHealthStaleAge() = %v, %v; want %v", got, err, c.want)
			}
		})
	}
}

// The tests below read what connection-log-worker writes; they need a MongoDB without access
// control in MONGO_TEST_URL and use a throwaway database.
func newHealthTestController(t *testing.T) (*Controller, *mongo.Collection) {
	t.Helper()
	url := os.Getenv("MONGO_TEST_URL")
	if testing.Short() || url == "" {
		t.Skip("needs MONGO_TEST_URL, not in -short")
	}
	db := "connectionlog_health_test_" + randomHex(t)
	conf := configuration.Config{
		MongoUrl:                url,
		MongoDatabase:           db,
		MongodbTimeout:          10,
		GatewayHealthCollection: "gatewayhealth",
		GatewayHealthStaleAge:   "15m",
		InfluxdbUrl:             "http://localhost:1",
	}
	ctrl, err := New(conf)
	if err != nil {
		t.Fatal(err)
	}
	seed, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = seed.Database(db).Drop(ctx)
		_ = seed.Disconnect(ctx)
		ctrl.Close()
	})
	return ctrl, seed.Database(db).Collection("gatewayhealth")
}

func seedHealth(t *testing.T, col *mongo.Collection, hubID string, received time.Time, report bson.M) {
	t.Helper()
	_, err := col.InsertOne(context.Background(), bson.M{
		"hub_id":      hubID,
		"state":       2,
		"reasons":     bson.A{"module a: unhealthy"},
		"report":      report,
		"report_time": "2026-10-27T14:00:00Z",
		"received":    received,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetGatewayHealth(t *testing.T) {
	ctrl, col := newHealthTestController(t)
	ctx := context.Background()
	hubFresh, hubOld, hubMissing := "urn:infai:ses:hub:fresh", "urn:infai:ses:hub:old", "urn:infai:ses:hub:none"
	report := bson.M{
		"version":       int32(1),
		"time":          "2026-10-27T14:00:00Z",
		"state":         int32(2),
		"reasons":       bson.A{"module a: unhealthy"},
		"core":          bson.M{"status": "ok", "uptime": 12.5, "big": int64(9007199254740993), "nested": bson.M{"list": bson.A{int32(1), "x", nil, true}}},
		"core_error":    "",
		"modules":       nil,
		"modules_error": "timeout",
	}
	seedHealth(t, col, hubFresh, time.Now().Add(-time.Minute), report)
	seedHealth(t, col, hubOld, time.Now().Add(-time.Hour), bson.M{"version": int32(1)})

	t.Run("fresh report keeps the original JSON shape", func(t *testing.T) {
		got, err := ctrl.GetGatewayHealth(ctx, hubFresh)
		if err != nil {
			t.Fatal(err)
		}
		if got.HubID != hubFresh || got.State != 2 || got.ReportTime != "2026-10-27T14:00:00Z" || got.Stale {
			t.Errorf("unexpected result: %+v", got)
		}
		if !reflect.DeepEqual(got.Reasons, []string{"module a: unhealthy"}) {
			t.Errorf("reasons = %v", got.Reasons)
		}
		if got.Received.Location() != time.UTC {
			t.Errorf("received is not UTC: %v", got.Received)
		}
		var reportJSON map[string]any
		if err := json.Unmarshal(got.Report, &reportJSON); err != nil {
			t.Fatalf("report is not JSON: %v: %s", err, got.Report)
		}
		core, _ := reportJSON["core"].(map[string]any)
		nested, _ := core["nested"].(map[string]any)
		if reportJSON["version"] != float64(1) || reportJSON["core_error"] != "" || reportJSON["modules_error"] != "timeout" ||
			core["status"] != "ok" || core["uptime"] != 12.5 || !reflect.DeepEqual(nested["list"], []any{float64(1), "x", nil, true}) {
			t.Errorf("report lost its shape: %s", got.Report)
		}
		if v, present := reportJSON["modules"]; !present || v != nil {
			t.Errorf("null field not kept: %s", got.Report)
		}
		if _, err := json.Marshal(got); err != nil {
			t.Errorf("response does not marshal: %v", err)
		}
	})
	t.Run("old report is stale", func(t *testing.T) {
		got, err := ctrl.GetGatewayHealth(ctx, hubOld)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Stale {
			t.Errorf("stale = false for a report received an hour ago")
		}
	})
	t.Run("missing report", func(t *testing.T) {
		if _, err := ctrl.GetGatewayHealth(ctx, hubMissing); !errors.Is(err, ErrGatewayHealthNotFound) {
			t.Errorf("err = %v, want ErrGatewayHealthNotFound", err)
		}
	})
}

func TestGetGatewayHealthList(t *testing.T) {
	ctrl, col := newHealthTestController(t)
	ctx := context.Background()
	now := time.Now()
	seedHealth(t, col, "urn:infai:ses:hub:c", now, bson.M{"version": int32(1)})
	seedHealth(t, col, "urn:infai:ses:hub:a", now, bson.M{"version": int32(1)})
	seedHealth(t, col, "urn:infai:ses:hub:b", now.Add(-time.Hour), bson.M{"version": int32(1)})
	seedHealth(t, col, "urn:infai:ses:hub:other", now, bson.M{"version": int32(1)})

	hubIDs := func(t *testing.T, requested []string) ([]string, []bool) {
		t.Helper()
		got, err := ctrl.GetGatewayHealthList(ctx, requested)
		if err != nil {
			t.Fatal(err)
		}
		var res []string
		var stale []bool
		for _, g := range got {
			res = append(res, g.HubID)
			stale = append(stale, g.Stale)
		}
		return res, stale
	}

	t.Run("sorted, omits hubs without report and other hubs", func(t *testing.T) {
		got, stale := hubIDs(t, []string{"urn:infai:ses:hub:c", "urn:infai:ses:hub:none", "urn:infai:ses:hub:b", "urn:infai:ses:hub:a"})
		want := []string{"urn:infai:ses:hub:a", "urn:infai:ses:hub:b", "urn:infai:ses:hub:c"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ids = %v, want %v", got, want)
		}
		if !reflect.DeepEqual(stale, []bool{false, true, false}) {
			t.Errorf("stale = %v", stale)
		}
	})
	t.Run("empty request gives an empty array, not null", func(t *testing.T) {
		got, err := ctrl.GetGatewayHealthList(ctx, nil)
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("got %v, %v", got, err)
		}
	})
	t.Run("no match gives an empty array, not null", func(t *testing.T) {
		got, err := ctrl.GetGatewayHealthList(ctx, []string{"urn:infai:ses:hub:none"})
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("got %v, %v", got, err)
		}
	})
}
