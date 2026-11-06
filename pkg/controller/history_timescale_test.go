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

	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// The tables are created by connection-log-worker; this is the contract this service reads.
const historyDDL = `
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE TABLE IF NOT EXISTS device (
    id        TEXT        NOT NULL,
    time      TIMESTAMPTZ NOT NULL,
    connected BOOLEAN     NOT NULL,
    PRIMARY KEY (id, time)
);
SELECT create_hypertable('device', by_range('time'), if_not_exists => TRUE);
ALTER TABLE device SET (timescaledb.enable_columnstore = true, timescaledb.segmentby = 'id', timescaledb.orderby = '"time" DESC');
CALL add_columnstore_policy('device', after => INTERVAL '30 days', if_not_exists => true);
CREATE TABLE IF NOT EXISTS gateway (
    id        TEXT        NOT NULL,
    time      TIMESTAMPTZ NOT NULL,
    connected BOOLEAN     NOT NULL,
    PRIMARY KEY (id, time)
);
SELECT create_hypertable('gateway', by_range('time'), if_not_exists => TRUE);
ALTER TABLE gateway SET (timescaledb.enable_columnstore = true, timescaledb.segmentby = 'id', timescaledb.orderby = '"time" DESC');
CALL add_columnstore_policy('gateway', after => INTERVAL '30 days', if_not_exists => true);
`

const timescaleImage = "timescale/timescaledb-ha:pg18.6-ts2.29.2"

func newTimescaleBackend(t *testing.T) (historyBackend, *pgxpool.Pool) {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        timescaleImage,
			ExposedPorts: []string{"5432/tcp"},
			Env:          map[string]string{"POSTGRES_PASSWORD": "pw", "POSTGRES_DB": "connectionlog"},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(2*time.Minute),
				wait.ForListeningPort("5432/tcp"),
			),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}
	// the pool is built the way New builds it, so the configuration path is exercised too
	pool, err := newPostgresPool(ctx, configuration.Config{
		PostgresHost: host,
		PostgresPort: int64(port.Int()),
		PostgresUser: "postgres",
		PostgresPw:   "pw",
		PostgresDb:   "connectionlog",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err = pool.Exec(ctx, historyDDL); err != nil {
		t.Fatal(err)
	}
	ctrl := &Controller{pg: pool}
	ctrl.config = configuration.Config{}
	return historyBackend{ctrl: ctrl, parameterized: true, seed: func(t *testing.T, kind, id string, anchor time.Time, pts []point) {
		table, err := historyTable(kind)
		if err != nil {
			t.Fatal(err)
		}
		for _, pt := range pts {
			ts := anchor.Add(time.Duration(pt.offset) * time.Second)
			if _, err := pool.Exec(ctx, "INSERT INTO "+table+" (id, time, connected) VALUES ($1, $2, $3) ON CONFLICT (id, time) DO NOTHING", id, ts, pt.connected); err != nil {
				t.Fatal(err)
			}
		}
	}}, pool
}

func TestHistoryBehaviorTimescale(t *testing.T) {
	b, pool := newTimescaleBackend(t)
	runHistoryBehavior(t, b)

	// The same cases as against InfluxDB are done. What follows needs resolution below one second,
	// which InfluxDB stored with precision s and the worker now writes in full.
	ctx := context.Background()
	anchor := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)
	ms := func(n int) time.Time { return anchor.Add(time.Duration(n) * time.Millisecond) }
	subA := "urn:infai:ses:device:subsecond-a"
	subGw := "urn:infai:ses:hub:subsecond-g"
	insert := func(table, id string, at time.Time, connected bool) {
		if _, err := pool.Exec(ctx, "INSERT INTO "+table+" (id, time, connected) VALUES ($1, $2, $3)", id, at, connected); err != nil {
			t.Fatal(err)
		}
	}
	insert("device", subA, ms(250), true)
	insert("device", subA, ms(500), false)
	insert("device", subA, ms(1500).Add(123*time.Microsecond), true)
	insert("gateway", subGw, ms(100), false)

	query := func(q model.QueryHistorical) map[string]model.HistoricalStates {
		t.Helper()
		res, err := b.ctrl.QueryHistoricalStatesMap(ctx, q)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	ptAt := func(tm time.Time, c bool) *model.State { return &model.State{Time: tm, Connected: c} }

	t.Run("sub-second states are returned without truncation", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{subA}}, Since: ms(0), Until: ms(2000)})[subA]
		want := []model.State{*ptAt(ms(250), true), *ptAt(ms(500), false), *ptAt(ms(1500).Add(123*time.Microsecond), true)}
		if len(got.States) != len(want) {
			t.Fatalf("want %v, got %v", want, got.States)
		}
		for i := range want {
			if !got.States[i].Time.Equal(want[i].Time) || got.States[i].Connected != want[i].Connected {
				t.Errorf("state %d: want %v, got %v", i, want[i], got.States[i])
			}
		}
		if got.PrevState != nil || got.NextState != nil {
			t.Errorf("unexpected prev or next: %v %v", got.PrevState, got.NextState)
		}
	})
	t.Run("sub-second bounds are exact", func(t *testing.T) {
		// since on a stored sub-second point: that point is a state, not prev
		got := query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{subA}}, Since: ms(500), Until: ms(500)})[subA]
		if len(got.States) != 1 || !got.States[0].Time.Equal(ms(500)) {
			t.Errorf("states: %v", got.States)
		}
		if got.PrevState == nil || !got.PrevState.Time.Equal(ms(250)) || !got.PrevState.Connected {
			t.Errorf("prev: %v", got.PrevState)
		}
		if got.NextState == nil || !got.NextState.Time.Equal(ms(1500).Add(123*time.Microsecond)) {
			t.Errorf("next: %v", got.NextState)
		}
		// a window between two points holds no state but both neighbours
		got = query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{subA}}, Since: ms(300), Until: ms(499)})[subA]
		if got.States != nil || got.PrevState == nil || !got.PrevState.Time.Equal(ms(250)) || got.NextState == nil || !got.NextState.Time.Equal(ms(500)) {
			t.Errorf("between: %+v", got)
		}
	})
	t.Run("sub-second range and until", func(t *testing.T) {
		got := query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{subA}}, Range: model.Duration(250 * time.Millisecond), Until: ms(500)})[subA]
		// window [250ms, 500ms], both points are inside
		if len(got.States) != 2 || got.PrevState != nil || got.NextState == nil {
			t.Errorf("%+v", got)
		}
	})
	t.Run("offline since keeps sub-second precision", func(t *testing.T) {
		b.seed(t, model.DeviceKind, "urn:infai:ses:device:subsecond-off", anchor, nil)
		insert("device", "urn:infai:ses:device:subsecond-off", ms(750), true)
		insert("device", "urn:infai:ses:device:subsecond-off", ms(1250), false)
		res, err := b.ctrl.GetOfflineSince(ctx, []string{"urn:infai:ses:device:subsecond-off", subA}, model.DeviceKind)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 || !res[0].OfflineSince.Equal(ms(1250)) {
			t.Errorf("got %v", res)
		}
		res, err = b.ctrl.GetOfflineSince(ctx, []string{subGw}, model.GatewayKind)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 || !res[0].OfflineSince.Equal(ms(100)) {
			t.Errorf("got %v", res)
		}
	})
	t.Run("fractional-second bounds are compared exactly", func(t *testing.T) {
		id := "urn:infai:ses:device:fraction"
		insert("device", id, anchor.Add(time.Hour), true)
		// since lies 700ms after the point: the point is prev, not part of the states
		got := query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{id}}, Since: anchor.Add(time.Hour + 700*time.Millisecond), Until: anchor.Add(2 * time.Hour)})[id]
		if got.PrevState == nil || !got.PrevState.Time.Equal(anchor.Add(time.Hour)) || got.States != nil {
			t.Errorf("got %+v", got)
		}
	})
	t.Run("results are read correctly from compressed chunks", func(t *testing.T) {
		for _, table := range []string{"device", "gateway"} {
			if _, err := pool.Exec(ctx, "SELECT compress_chunk(c) FROM show_chunks('"+table+"') c"); err != nil {
				t.Fatal(err)
			}
			var n int
			if err := pool.QueryRow(ctx, "SELECT count(*) FROM timescaledb_information.chunks WHERE hypertable_name = $1 AND is_compressed", table).Scan(&n); err != nil || n == 0 {
				t.Fatalf("%s: no compressed chunk: %d %v", table, n, err)
			}
		}
		got := query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{subA, devA}}, Since: ms(300), Until: ms(499)})
		if g := got[subA]; g.States != nil || g.PrevState == nil || !g.PrevState.Time.Equal(ms(250)) || g.NextState == nil || !g.NextState.Time.Equal(ms(500)) {
			t.Errorf("prev and next: %+v", g)
		}
		series := query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{subA}}, Since: ms(0), Until: ms(2000)})[subA]
		if len(series.States) != 3 || !series.States[2].Time.Equal(ms(1500).Add(123*time.Microsecond)) {
			t.Errorf("series: %+v", series)
		}
		checkStates(t, behaviorAnchor, query(model.QueryHistorical{QueryBase: model.QueryBase{IDs: []string{devA}}, Since: behaviorAnchor.Add(-time.Hour), Until: behaviorAnchor.Add(time.Hour)}), map[string]expectation{
			devA: {prev: p(-7200, true), states: []point{{-3600, false}, {0, true}, {3600, false}}, next: p(7200, true)},
		})
		res, err := b.ctrl.GetOfflineSince(ctx, []string{"urn:infai:ses:device:subsecond-off", subA, "urn:infai:ses:device:off-a"}, model.DeviceKind)
		if err != nil || len(res) != 2 || res[0].ID != "urn:infai:ses:device:off-a" || res[1].ID != "urn:infai:ses:device:subsecond-off" || !res[1].OfflineSince.Equal(ms(1250)) {
			t.Errorf("offline since: %v %v", res, err)
		}
	})
	t.Run("offline since without ids is empty", func(t *testing.T) {
		res, err := b.ctrl.GetOfflineSince(ctx, nil, model.DeviceKind)
		if err != nil || res == nil || len(res) != 0 {
			t.Errorf("got %v %v", res, err)
		}
	})
	t.Run("injected id is data and the tables survive", func(t *testing.T) {
		var n int
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM device").Scan(&n); err != nil || n == 0 {
			t.Errorf("device table gone or empty: %d %v", n, err)
		}
	})
	t.Run("invalid kind never reaches a table name", func(t *testing.T) {
		for _, kind := range []string{"", "device; DROP TABLE device", "devices", "hubs"} {
			if _, err := b.ctrl.GetOfflineSince(ctx, []string{"x"}, kind); err == nil {
				t.Errorf("kind %q accepted", kind)
			}
		}
	})
}
