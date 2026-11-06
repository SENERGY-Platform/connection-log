package controller

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
)

func TestResolveWindow(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	since := now.Add(-3 * time.Hour)
	until := now.Add(-1 * time.Hour)
	rng := model.Duration(30 * time.Minute)
	tp := func(v time.Time) *time.Time { return &v }
	cases := []struct {
		name  string
		query model.QueryHistorical
		want  historyWindow
	}{
		{"since and until", model.QueryHistorical{Since: since, Until: until}, historyWindow{tp(since), tp(since), tp(until), tp(until)}},
		{"since and until win over range", model.QueryHistorical{Since: since, Until: until, Range: rng}, historyWindow{tp(since), tp(since), tp(until), tp(until)}},
		{"range and until", model.QueryHistorical{Range: rng, Until: until}, historyWindow{tp(until.Add(-30 * time.Minute)), tp(until.Add(-30 * time.Minute)), tp(until), tp(until)}},
		{"range and since", model.QueryHistorical{Range: rng, Since: since}, historyWindow{tp(since), tp(since), tp(since.Add(30 * time.Minute)), tp(since.Add(30 * time.Minute))}},
		{"range", model.QueryHistorical{Range: rng}, historyWindow{prevBefore: tp(now.Add(-30 * time.Minute)), from: tp(now.Add(-30 * time.Minute))}},
		{"until", model.QueryHistorical{Until: until}, historyWindow{to: tp(until), nextAfter: tp(until)}},
		{"since", model.QueryHistorical{Since: since}, historyWindow{prevBefore: tp(since), from: tp(since)}},
		{"nothing", model.QueryHistorical{}, historyWindow{to: tp(now)}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveWindow(c.query, now); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestStatementsUseOnlyParametersForValues(t *testing.T) {
	ids := []string{"urn:infai:ses:device:x'; DROP TABLE device; --"}
	now := time.Now()
	w := historyWindow{prevBefore: &now, from: &now, to: &now, nextAfter: &now}
	stmts := map[string]string{}
	var args = map[string][]any{}
	stmts["series"], args["series"] = seriesStatement("device", ids, w)
	stmts["prev"], args["prev"] = prevStatement("device", ids, now)
	stmts["next"], args["next"] = nextStatement("device", ids, now)
	stmts["offline"], args["offline"] = offlineSinceStatement("device", ids)
	for name, stmt := range stmts {
		if strings.Contains(stmt, "DROP") || strings.Contains(stmt, "x'") {
			t.Errorf("%s: id reached the statement text: %s", name, stmt)
		}
		if len(args[name]) == 0 || !reflect.DeepEqual(args[name][0], ids) {
			t.Errorf("%s: ids are not the first argument", name)
		}
	}
	if !strings.Contains(stmts["series"], "time >= $2") || !strings.Contains(stmts["series"], "time <= $3") {
		t.Errorf("series: %s", stmts["series"])
	}
}

func TestHistoryTableIsFixed(t *testing.T) {
	for kind, want := range map[string]string{model.DeviceKind: "device", model.GatewayKind: "gateway"} {
		if got, err := historyTable(kind); err != nil || got != want {
			t.Errorf("%s: got %q %v", kind, got, err)
		}
	}
	if _, err := historyTable("device\"; DROP TABLE x; --"); err == nil {
		t.Error("expected error")
	}
}
