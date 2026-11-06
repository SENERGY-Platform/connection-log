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
	"testing"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
)

// The cases in this file pin the semantics of the history queries independent of the storage behind them.
// A backend supplies a Controller and a way to seed rows. The cases were validated against the InfluxDB
// implementation on 2026-11-06 before it was removed; they run against TimescaleDB now.

// point is one stored state. Offsets are whole seconds relative to an anchor, so every backend can store them.
type point struct {
	offset    int // seconds relative to the anchor
	connected bool
}

type historyBackend struct {
	ctrl *Controller
	// parameterized is false for a backend that builds queries by string concatenation (InfluxQL):
	// ids containing query syntax break it there, which is the reason for the switch to bound parameters.
	parameterized bool
	// seed stores the points of one resource: kind is model.DeviceKind or model.GatewayKind,
	// anchor+offset is the timestamp. Sub-second anchors are only used by backends that support them.
	seed func(t *testing.T, kind, id string, anchor time.Time, pts []point)
}

var behaviorAnchor = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

const (
	devA = "urn:infai:ses:device:behavior-a"
	devB = "urn:infai:ses:device:behavior-b"
	devC = "urn:infai:ses:device:behavior-c-without-data"
	devD = "urn:infai:ses:device:behavior-d"
	gwA  = "urn:infai:ses:hub:behavior-a"
	gwC  = "urn:infai:ses:hub:behavior-c-without-data"
)

// expectation of one resource in a result. wantStates == nil means the result must hold nil States.
type expectation struct {
	prev, next *point
	states     []point
}

func p(offset int, connected bool) *point { return &point{offset: offset, connected: connected} }

func seedBehaviorData(t *testing.T, b historyBackend) {
	t.Helper()
	b.seed(t, model.DeviceKind, devA, behaviorAnchor, []point{{-7200, true}, {-3600, false}, {0, true}, {3600, false}, {7200, true}})
	b.seed(t, model.DeviceKind, devB, behaviorAnchor, []point{{-1800, true}, {1800, false}})
	b.seed(t, model.DeviceKind, devD, behaviorAnchor, []point{{0, false}})
	b.seed(t, model.GatewayKind, gwA, behaviorAnchor, []point{{-3600, true}, {0, false}, {3600, true}})
}

func checkStates(t *testing.T, anchor time.Time, got map[string]model.HistoricalStates, want map[string]expectation) {
	t.Helper()
	for id := range got {
		if _, ok := want[id]; !ok {
			t.Errorf("unexpected id in result: %s", id)
		}
	}
	for id, w := range want {
		g, ok := got[id]
		if !ok {
			t.Errorf("%s: missing in result", id)
			continue
		}
		checkOne(t, id+" prev", anchor, g.PrevState, w.prev)
		checkOne(t, id+" next", anchor, g.NextState, w.next)
		if len(w.states) == 0 {
			if g.States != nil {
				t.Errorf("%s: want nil states, got %v", id, g.States)
			}
		} else if len(g.States) != len(w.states) {
			t.Errorf("%s: want %d states, got %d: %v", id, len(w.states), len(g.States), g.States)
		} else {
			for i := range w.states {
				checkOne(t, id+" states", anchor, &g.States[i], &w.states[i])
			}
		}
	}
}

func checkOne(t *testing.T, what string, anchor time.Time, g *model.State, w *point) {
	t.Helper()
	switch {
	case g == nil && w == nil:
	case g == nil:
		t.Errorf("%s: want %v, got nil", what, anchor.Add(time.Duration(w.offset)*time.Second))
	case w == nil:
		t.Errorf("%s: want nil, got %v", what, *g)
	default:
		wt := anchor.Add(time.Duration(w.offset) * time.Second)
		if g.Time.Location() != time.UTC {
			t.Errorf("%s: want location UTC, got %v", what, g.Time.Location())
		}
		if !g.Time.Equal(wt) || g.Connected != w.connected {
			t.Errorf("%s: want {%v %v}, got {%v %v}", what, wt, w.connected, g.Time, g.Connected)
		}
	}
}

func runHistoryBehavior(t *testing.T, b historyBackend) {
	seedBehaviorData(t, b)
	at := func(offset int) time.Time { return behaviorAnchor.Add(time.Duration(offset) * time.Second) }
	query := func(q model.QueryHistorical) map[string]model.HistoricalStates {
		t.Helper()
		res, err := b.ctrl.QueryHistoricalStatesMap(context.Background(), q)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	ids := func(ids ...string) model.QueryBase { return model.QueryBase{IDs: ids} }

	t.Run("since and until", func(t *testing.T) {
		// both bounds are inclusive for states, prev is strictly before since, next strictly after until
		got := query(model.QueryHistorical{QueryBase: ids(devA, devB, devD), Since: at(-3600), Until: at(3600)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {prev: p(-7200, true), states: []point{{-3600, false}, {0, true}, {3600, false}}, next: p(7200, true)},
			devB: {states: []point{{-1800, true}, {1800, false}}},
			devD: {states: []point{{0, false}}},
		})
	})
	t.Run("since equals until on a point", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devA, devB, devD), Since: at(0), Until: at(0)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {prev: p(-3600, false), states: []point{{0, true}}, next: p(3600, false)},
			devB: {prev: p(-1800, true), next: p(1800, false)},
			devD: {states: []point{{0, false}}},
		})
	})
	t.Run("since after the last point", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devA), Since: at(7201), Until: at(9000)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {prev: p(7200, true)},
		})
	})
	t.Run("range and until", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devA), Range: model.Duration(time.Hour), Until: at(3600)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {prev: p(-3600, false), states: []point{{0, true}, {3600, false}}, next: p(7200, true)},
		})
	})
	t.Run("range and since", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devA), Range: model.Duration(time.Hour), Since: at(-3600)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {prev: p(-7200, true), states: []point{{-3600, false}, {0, true}}, next: p(3600, false)},
		})
	})
	t.Run("since only", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devA, devB), Since: at(0)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {prev: p(-3600, false), states: []point{{0, true}, {3600, false}, {7200, true}}},
			devB: {prev: p(-1800, true), states: []point{{1800, false}}},
		})
	})
	t.Run("until only", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devA, devB), Until: at(0)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {states: []point{{-7200, true}, {-3600, false}, {0, true}}, next: p(3600, false)},
			devB: {states: []point{{-1800, true}}, next: p(1800, false)},
		})
	})
	t.Run("gateway", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(gwA), Since: at(-3600), Until: at(0)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			gwA: {states: []point{{-3600, true}, {0, false}}, next: p(3600, true)},
		})
	})
	t.Run("devices and gateways in one query", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devA, gwA), Since: at(0), Until: at(0)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{
			devA: {prev: p(-3600, false), states: []point{{0, true}}, next: p(3600, false)},
			gwA:  {prev: p(-3600, true), states: []point{{0, false}}, next: p(3600, true)},
		})
	})
	t.Run("id without data is absent", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids(devC, gwC), Since: at(-3600), Until: at(3600)})
		checkStates(t, behaviorAnchor, got, map[string]expectation{})
		got = query(model.QueryHistorical{QueryBase: ids(devA, devC), Since: at(-3600), Until: at(3600)})
		if _, ok := got[devC]; ok {
			t.Errorf("%s must not be in the result", devC)
		}
		if _, ok := got[devA]; !ok {
			t.Errorf("%s missing", devA)
		}
	})
	t.Run("GetHistoricalStates of an id without data is not found", func(t *testing.T) {
		if _, err := b.ctrl.GetHistoricalStates(context.Background(), devC, model.DeviceKind, 0, at(-3600), at(3600)); err == nil {
			t.Error("expected error")
		}
	})
	t.Run("id of unknown kind is rejected", func(t *testing.T) {
		if _, err := b.ctrl.QueryHistoricalStatesMap(context.Background(), model.QueryHistorical{QueryBase: ids("urn:infai:ses:location:x")}); err == nil {
			t.Error("expected error")
		}
	})
	t.Run("ids with quotes and sql syntax are treated as data", func(t *testing.T) {
		if !b.parameterized {
			t.Skip("InfluxQL is built by string concatenation, an id with quotes is a syntax error there")
		}
		// ids reach the query only through GetIdsByKind, which requires the device prefix
		evil := "urn:infai:ses:device:x'; DROP TABLE device; --"
		got := query(model.QueryHistorical{QueryBase: ids(evil, devA), Since: at(-3600), Until: at(3600)})
		if _, ok := got[evil]; ok {
			t.Error("evil id must not match")
		}
		if _, ok := got[devA]; !ok {
			t.Error("devA missing after evil id")
		}
	})

	// range only and the default depend on the current time, so they get their own relative data
	now := time.Now().UTC().Truncate(time.Second)
	b.seed(t, model.DeviceKind, "urn:infai:ses:device:behavior-now", now, []point{{-7200, true}, {-1800, false}, {-60, true}, {3600, false}})
	t.Run("range only has no upper bound", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids("urn:infai:ses:device:behavior-now"), Range: model.Duration(time.Hour)})
		checkStates(t, now, got, map[string]expectation{
			"urn:infai:ses:device:behavior-now": {prev: p(-7200, true), states: []point{{-1800, false}, {-60, true}, {3600, false}}},
		})
	})
	t.Run("nothing set returns everything up to now", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: ids("urn:infai:ses:device:behavior-now")})
		checkStates(t, now, got, map[string]expectation{
			"urn:infai:ses:device:behavior-now": {states: []point{{-7200, true}, {-1800, false}, {-60, true}}},
		})
	})

	t.Run("offline since", func(t *testing.T) {
		// last state false -> listed, last state true -> not, no data -> not; sorted ascending by time
		oa, ob, oc, od := "urn:infai:ses:device:off-a", "urn:infai:ses:device:off-b", "urn:infai:ses:device:off-c", "urn:infai:ses:device:off-d"
		b.seed(t, model.DeviceKind, oa, behaviorAnchor, []point{{-100, true}, {-50, false}})
		b.seed(t, model.DeviceKind, ob, behaviorAnchor, []point{{-200, false}, {-10, true}})
		b.seed(t, model.DeviceKind, od, behaviorAnchor, []point{{-500, true}, {-400, false}})
		res, err := b.ctrl.GetOfflineSince(context.Background(), []string{oa, ob, oc, od}, model.DeviceKind)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 2 {
			t.Fatalf("want 2 entries, got %v", res)
		}
		if res[0].ID != od || !res[0].OfflineSince.Equal(at(-400)) {
			t.Errorf("first: want %s since %v, got %v", od, at(-400), res[0])
		}
		if res[1].ID != oa || !res[1].OfflineSince.Equal(at(-50)) {
			t.Errorf("second: want %s since %v, got %v", oa, at(-50), res[1])
		}
	})
	t.Run("offline since for gateways", func(t *testing.T) {
		res, err := b.ctrl.GetOfflineSince(context.Background(), []string{gwA, gwC}, model.GatewayKind)
		if err != nil {
			t.Fatal(err)
		}
		// gwA ends with a connect at +3600
		if len(res) != 0 {
			t.Errorf("want none, got %v", res)
		}
		b.seed(t, model.GatewayKind, "urn:infai:ses:hub:off-x", behaviorAnchor, []point{{-5, false}})
		res, err = b.ctrl.GetOfflineSince(context.Background(), []string{gwA, gwC, "urn:infai:ses:hub:off-x"}, model.GatewayKind)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 || res[0].ID != "urn:infai:ses:hub:off-x" || !res[0].OfflineSince.Equal(at(-5)) {
			t.Errorf("want off-x since %v, got %v", at(-5), res)
		}
	})
	t.Run("offline since does not mix kinds", func(t *testing.T) {
		// a device id asked for in the gateway table has no rows there
		res, err := b.ctrl.GetOfflineSince(context.Background(), []string{"urn:infai:ses:device:off-a"}, model.GatewayKind)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 0 {
			t.Errorf("want none, got %v", res)
		}
	})
}
