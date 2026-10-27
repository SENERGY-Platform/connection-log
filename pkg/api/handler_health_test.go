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
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	deviceRepo "github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	hubA      = "urn:infai:ses:hub:a"
	hubB      = "urn:infai:ses:hub:b"
	hubDenied = "urn:infai:ses:hub:denied"
	hubNoDoc  = "urn:infai:ses:hub:nodoc"
	deviceID  = "urn:infai:ses:device:d"
)

type route func(*controller.Controller, deviceRepo.Interface) (string, string, httprouter.Handle)

func do(ctrl *controller.Controller, r route, method, target, body string) *httptest.ResponseRecorder {
	m, p, h := r(ctrl, nil)
	router := httprouter.New()
	router.Handle(m, p, h)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// These requests are rejected before the controller is used, so it can be nil.
func TestGatewayHealthRejectsNonGatewayIds(t *testing.T) {
	cases := []struct {
		name   string
		route  route
		method string
		target string
		body   string
	}{
		{"get with device id", GetCurrentGatewayHealth, http.MethodGet, "/current/gateways/" + deviceID + "/health", ""},
		{"get with unknown prefix", GetCurrentGatewayHealth, http.MethodGet, "/current/gateways/foo/health", ""},
		{"query with device id", PostQueryGatewayHealth, http.MethodPost, "/current/query/health", `{"ids":["` + hubA + `","` + deviceID + `"]}`},
		{"query with empty id", PostQueryGatewayHealth, http.MethodPost, "/current/query/health", `{"ids":[""]}`},
		{"query with invalid body", PostQueryGatewayHealth, http.MethodPost, "/current/query/health", `{"ids":`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if rec := do(nil, c.route, c.method, c.target, c.body); rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body: %s", rec.Code, rec.Body)
			}
		})
	}
}

// The tests below need a MongoDB without access control in MONGO_TEST_URL (throwaway database)
// and use a fake permissions-v2 that grants everything except hubDenied.
func newHealthAPITestController(t *testing.T) (*controller.Controller, *mongo.Collection, *[]string) {
	t.Helper()
	url := os.Getenv("MONGO_TEST_URL")
	if testing.Short() || url == "" {
		t.Skip("needs MONGO_TEST_URL, not in -short")
	}
	var calls []string
	perms := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/accessible/hubs":
			_ = json.NewEncoder(w).Encode([]string{hubB, hubA, hubNoDoc})
		case r.URL.Path == "/check/hubs":
			res := map[string]bool{}
			for _, id := range strings.Split(r.URL.Query().Get("ids"), ",") {
				res[id] = id != hubDenied
			}
			_ = json.NewEncoder(w).Encode(res)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(perms.Close)

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	db := "connectionlog_health_api_test_" + hex.EncodeToString(b)
	ctrl, err := controller.New(configuration.Config{
		MongoUrl:                url,
		MongoDatabase:           db,
		MongodbTimeout:          10,
		GatewayHealthCollection: "gatewayhealth",
		GatewayHealthStaleAge:   "15m",
		PermissionsV2Url:        perms.URL,
		InfluxdbUrl:             "http://localhost:1",
	})
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
	col := seed.Database(db).Collection("gatewayhealth")
	for hub, received := range map[string]time.Time{hubA: time.Now(), hubB: time.Now().Add(-time.Hour), hubDenied: time.Now()} {
		_, err := col.InsertOne(context.Background(), bson.M{
			"hub_id": hub, "state": 1, "reasons": bson.A{}, "report": bson.M{"version": int32(1), "core": bson.M{"ok": true}},
			"report_time": "2026-10-27T14:00:00Z", "received": received,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return ctrl, col, &calls
}

func decodeList(t *testing.T, body []byte) []model.GatewayHealth {
	t.Helper()
	var res []model.GatewayHealth
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("response is not a health list: %v: %s", err, body)
	}
	return res
}

func hubIDsOf(list []model.GatewayHealth) []string {
	res := []string{}
	for _, g := range list {
		res = append(res, g.HubID)
	}
	return res
}

func TestGetCurrentGatewayHealth(t *testing.T) {
	ctrl, _, _ := newHealthAPITestController(t)

	t.Run("report as JSON with the original report fields", func(t *testing.T) {
		rec := do(ctrl, GetCurrentGatewayHealth, http.MethodGet, "/current/gateways/"+hubA+"/health", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", rec.Code, rec.Body)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"hub_id", "state", "reasons", "report", "report_time", "received", "stale"} {
			if _, ok := raw[key]; !ok {
				t.Errorf("response lacks %q: %s", key, rec.Body)
			}
		}
		var report map[string]any
		if err := json.Unmarshal(raw["report"], &report); err != nil || report["version"] != float64(1) {
			t.Errorf("report = %s, %v", raw["report"], err)
		}
		if string(raw["stale"]) != "false" || string(raw["reasons"]) != "[]" {
			t.Errorf("stale = %s, reasons = %s", raw["stale"], raw["reasons"])
		}
	})
	t.Run("stale", func(t *testing.T) {
		rec := do(ctrl, GetCurrentGatewayHealth, http.MethodGet, "/current/gateways/"+hubB+"/health", "")
		var res model.GatewayHealth
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil || !res.Stale {
			t.Errorf("status %d, body %s", rec.Code, rec.Body)
		}
	})
	t.Run("no report gives 404", func(t *testing.T) {
		if rec := do(ctrl, GetCurrentGatewayHealth, http.MethodGet, "/current/gateways/"+hubNoDoc+"/health", ""); rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})
	t.Run("no access gives 401", func(t *testing.T) {
		if rec := do(ctrl, GetCurrentGatewayHealth, http.MethodGet, "/current/gateways/"+hubDenied+"/health", ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})
}

func TestPostQueryGatewayHealth(t *testing.T) {
	ctrl, _, calls := newHealthAPITestController(t)

	t.Run("empty ids list the accessible hubs, sorted, without hubs lacking a report", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"ids":[]}`, `{"ids":null}`} {
			*calls = nil
			rec := do(ctrl, PostQueryGatewayHealth, http.MethodPost, "/current/query/health", body)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s: status = %d: %s", body, rec.Code, rec.Body)
			}
			if got := hubIDsOf(decodeList(t, rec.Body.Bytes())); !reflect.DeepEqual(got, []string{hubA, hubB}) {
				t.Errorf("%s: ids = %v", body, got)
			}
			if !reflect.DeepEqual(*calls, []string{"/accessible/hubs"}) {
				t.Errorf("%s: permission calls = %v, want only the accessible-hubs list", body, *calls)
			}
		}
	})
	t.Run("given ids are filtered by permission, not rejected", func(t *testing.T) {
		rec := do(ctrl, PostQueryGatewayHealth, http.MethodPost, "/current/query/health", `{"ids":["`+hubDenied+`","`+hubB+`","`+hubNoDoc+`"]}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", rec.Code, rec.Body)
		}
		if got := hubIDsOf(decodeList(t, rec.Body.Bytes())); !reflect.DeepEqual(got, []string{hubB}) {
			t.Errorf("ids = %v, want only %s", got, hubB)
		}
	})
	t.Run("nothing found is an empty array", func(t *testing.T) {
		rec := do(ctrl, PostQueryGatewayHealth, http.MethodPost, "/current/query/health", `{"ids":["`+hubDenied+`"]}`)
		if strings.TrimSpace(rec.Body.String()) != "[]" {
			t.Errorf("body = %s", rec.Body)
		}
	})
}
