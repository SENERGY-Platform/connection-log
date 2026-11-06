package controller

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
)

func (this *Controller) GetHistoricalStates(ctx context.Context, id, kind string, rng time.Duration, since, until time.Time) (model.ResourceHistoricalStates, error) {
	resMap, err := this.QueryHistoricalStatesMap(ctx, model.QueryHistorical{
		QueryBase: model.QueryBase{
			IDs: []string{id},
		},
		Range: model.Duration(rng),
		Since: since,
		Until: until,
	})
	if err != nil {
		return model.ResourceHistoricalStates{}, err
	}
	item, ok := resMap[id]
	if !ok {
		return model.ResourceHistoricalStates{}, errors.New("not found")
	}
	return model.ResourceHistoricalStates{
		ID:               id,
		HistoricalStates: item,
	}, nil
}

func (this *Controller) QueryHistoricalStatesSlice(ctx context.Context, query model.QueryHistorical) ([]model.ResourceHistoricalStates, error) {
	resMap, err := this.QueryHistoricalStatesMap(ctx, query)
	if err != nil {
		return nil, err
	}
	sl := make([]model.ResourceHistoricalStates, 0, len(resMap))
	for id, resource := range resMap {
		sl = append(sl, model.ResourceHistoricalStates{
			ID:               id,
			HistoricalStates: resource,
		})
	}
	return sl, nil
}

// QueryHistoricalStatesMap returns the states of the requested ids. An id without any state inside
// or next to the window is not part of the result.
func (this *Controller) QueryHistoricalStatesMap(ctx context.Context, query model.QueryHistorical) (map[string]model.HistoricalStates, error) {
	idsBykind, err := GetIdsByKind(query.IDs, false)
	if err != nil {
		return nil, err
	}
	window := resolveWindow(query, time.Now())
	resMap := map[string]model.HistoricalStates{}
	for kind, ids := range idsBykind {
		table, err := historyTable(kind)
		if err != nil {
			return nil, err
		}
		subResMap, err := this.queryWindow(ctx, table, ids, window)
		if err != nil {
			return nil, err
		}
		maps.Copy(resMap, subResMap)
	}
	return resMap, nil
}

func (this *Controller) queryWindow(ctx context.Context, table string, ids []string, w historyWindow) (map[string]model.HistoricalStates, error) {
	resMap := map[string]model.HistoricalStates{}
	if w.prevBefore != nil {
		stmt, args := prevStatement(table, ids, *w.prevBefore)
		err := this.queryStates(ctx, stmt, args, func(id string, state model.State) {
			res := resMap[id]
			res.PrevState = &state
			resMap[id] = res
		})
		if err != nil {
			return nil, err
		}
	}
	stmt, args := seriesStatement(table, ids, w)
	err := this.queryStates(ctx, stmt, args, func(id string, state model.State) {
		res := resMap[id]
		res.States = append(res.States, state)
		resMap[id] = res
	})
	if err != nil {
		return nil, err
	}
	if w.nextAfter != nil {
		stmt, args := nextStatement(table, ids, *w.nextAfter)
		err := this.queryStates(ctx, stmt, args, func(id string, state model.State) {
			res := resMap[id]
			res.NextState = &state
			resMap[id] = res
		})
		if err != nil {
			return nil, err
		}
	}
	return resMap, nil
}

func (this *Controller) queryStates(ctx context.Context, stmt string, args []any, handle func(id string, state model.State)) error {
	rows, err := this.pg.Query(ctx, stmt, args...)
	if err != nil {
		return fmt.Errorf("unable to query connection history: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var state model.State
		if err := rows.Scan(&id, &state.Time, &state.Connected); err != nil {
			return fmt.Errorf("unable to read connection history: %w", err)
		}
		state.Time = state.Time.UTC()
		handle(id, state)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("unable to read connection history: %w", err)
	}
	return nil
}
