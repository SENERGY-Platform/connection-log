package controller

import (
	"fmt"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
)

// historyTables maps a kind to its table in the connectionlog database. The table name is never taken from a
// request: it is the value of this map, selected by a kind that GetIdsByKind derived from an id prefix.
// The tables are created and written by connection-log-worker.
var historyTables = map[string]string{
	model.DeviceKind:  "device",
	model.GatewayKind: "gateway",
}

func historyTable(kind string) (string, error) {
	if err := validateKind(kind); err != nil {
		return "", err
	}
	table, ok := historyTables[kind]
	if !ok {
		return "", fmt.Errorf("no history table for kind '%s'", kind)
	}
	return table, nil
}

// historyWindow selects what a history query returns. A nil bound is not applied.
type historyWindow struct {
	prevBefore *time.Time // last state strictly before this timestamp
	from       *time.Time // states with time >= from
	to         *time.Time // states with time <= to
	nextAfter  *time.Time // first state strictly after this timestamp
}

// resolveWindow turns the combinations of range, since and until into a window.
// The precedence is the one of the InfluxDB implementation this replaces.
// Bounds are compared exactly, without truncation to seconds (the InfluxDB code formatted them with time.RFC3339).
func resolveWindow(query model.QueryHistorical, now time.Time) historyWindow {
	hasRange := query.Range > 0
	hasSince := !query.Since.IsZero()
	hasUntil := !query.Until.IsZero()
	rng := time.Duration(query.Range)
	between := func(since, until time.Time) historyWindow {
		return historyWindow{prevBefore: &since, from: &since, to: &until, nextAfter: &until}
	}
	switch {
	case hasSince && hasUntil:
		return between(query.Since, query.Until)
	case hasRange && hasUntil:
		return between(query.Until.Add(-rng), query.Until)
	case hasRange && hasSince:
		return between(query.Since, query.Since.Add(rng))
	case hasRange:
		since := now.Add(-rng)
		return historyWindow{prevBefore: &since, from: &since}
	case hasUntil:
		until := query.Until
		return historyWindow{to: &until, nextAfter: &until}
	case hasSince:
		since := query.Since
		return historyWindow{prevBefore: &since, from: &since}
	default:
		return historyWindow{to: &now}
	}
}

// seriesStatement selects all states of the given ids inside the window, oldest first per id.
func seriesStatement(table string, ids []string, w historyWindow) (string, []any) {
	args := []any{ids}
	stmt := "SELECT id, time, connected FROM " + table + " WHERE id = ANY($1)"
	if w.from != nil {
		args = append(args, *w.from)
		stmt += fmt.Sprintf(" AND time >= $%d", len(args))
	}
	if w.to != nil {
		args = append(args, *w.to)
		stmt += fmt.Sprintf(" AND time <= $%d", len(args))
	}
	return stmt + " ORDER BY id, time", args
}

// prevStatement selects the last state of each id strictly before the timestamp.
func prevStatement(table string, ids []string, before time.Time) (string, []any) {
	return "SELECT DISTINCT ON (id) id, time, connected FROM " + table + " WHERE id = ANY($1) AND time < $2 ORDER BY id, time DESC",
		[]any{ids, before}
}

// nextStatement selects the first state of each id strictly after the timestamp.
func nextStatement(table string, ids []string, after time.Time) (string, []any) {
	return "SELECT DISTINCT ON (id) id, time, connected FROM " + table + " WHERE id = ANY($1) AND time > $2 ORDER BY id, time ASC",
		[]any{ids, after}
}

// offlineSinceStatement selects the ids whose last state is disconnected, with the time of that state, oldest first.
func offlineSinceStatement(table string, ids []string) (string, []any) {
	return "SELECT id, time FROM (SELECT DISTINCT ON (id) id, time, connected FROM " + table +
			" WHERE id = ANY($1) ORDER BY id, time DESC) AS last_state WHERE NOT connected ORDER BY time, id",
		[]any{ids}
}
