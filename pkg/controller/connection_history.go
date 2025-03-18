package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"github.com/influxdata/influxdb1-client/models"
	influx "github.com/influxdata/influxdb1-client/v2"
	"log"
	"time"
)

func (this *Controller) GetHistoricalStates(ctx context.Context, id, kind string, rng time.Duration, since, until time.Time) (model.ResourceHistoricalStates, error) {
	resMap, err := this.QueryHistoricalStatesMap(ctx, model.QueryHistorical{
		QueryBase: model.QueryBase{
			Kind: kind,
			IDs:  []string{id},
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

func (this *Controller) QueryHistoricalStatesMap(_ context.Context, query model.QueryHistorical) (map[string]model.HistoricalStates, error) {
	if len(query.IDs) == 0 {
		return nil, nil
	}
	statement, prevID, seriesID, nextID, err := this.buildStatement(query)
	if err != nil {
		return nil, err
	}
	resp, err := this.influx.Query(influx.NewQuery(statement, this.config.InfluxdbDb, "s"))
	if err != nil {
		return nil, err
	}
	if err = resp.Error(); err != nil {
		return nil, err
	}
	resMap, err := handleResults(resp.Results, query.Kind, prevID, seriesID, nextID)
	if err != nil {
		return nil, err
	}
	return resMap, nil
}

func handleResults(results []influx.Result, kind string, prevID, seriesID, nextID int) (map[string]model.HistoricalStates, error) {
	if len(results) == 0 {
		return nil, errors.New("no results")
	}
	resMap := make(map[string]model.HistoricalStates)
	for _, result := range results {
		if result.Err != "" {
			return nil, errors.New(result.Err)
		}
		switch result.StatementId {
		case seriesID:
			handleSeries(resMap, kind, result.Series, 0)
		case prevID:
			handleSeries(resMap, kind, result.Series, 1)
		case nextID:
			handleSeries(resMap, kind, result.Series, 2)
		default:
			return nil, fmt.Errorf("unknown statement id: %d", result.StatementId)
		}

	}
	return resMap, nil
}

func handleSeries(resMap map[string]model.HistoricalStates, kind string, series []models.Row, resType int) {
	for _, row := range series {
		if len(row.Values) == 0 {
			continue
		}
		key, ok := row.Tags[kind]
		if !ok {
			continue
		}
		if err := handleRow(resMap, row.Values, key, resType); err != nil {
			log.Println("ERROR:", err)
			continue
		}
	}
}

func handleRow(resMap map[string]model.HistoricalStates, rowValues [][]any, key string, resType int) error {
	resource := resMap[key]
	if resType > 0 {
		state, err := rowItemToState(rowValues[0])
		if err != nil {
			return err
		}
		switch resType {
		case 1:
			resource.PrevState = &state
		case 2:
			resource.NextState = &state
		}
	} else {
		for _, item := range rowValues {
			state, err := rowItemToState(item)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
			resource.States = append(resource.States, state)
		}
	}
	resMap[key] = resource
	return nil
}

func rowItemToState(item []any) (model.State, error) {
	if len(item) < 2 {
		return model.State{}, errors.New("invalid length")
	}
	timeVal, ok := item[0].(json.Number)
	if !ok {
		return model.State{}, fmt.Errorf("invalid type: time=%t", item[0])
	}
	timeInt, err := timeVal.Int64()
	if err != nil {
		return model.State{}, fmt.Errorf("time conversion failed: %s", err)
	}
	connected, ok := item[1].(bool)
	if !ok {
		return model.State{}, fmt.Errorf("invalid type: connected=%t", item[1])
	}
	return model.State{
		Time:      time.Unix(timeInt, 0).UTC(),
		Connected: connected,
	}, nil
}

func (this *Controller) buildStatement(query model.QueryHistorical) (string, int, int, int, error) {
	hasRange := query.Range > 0
	hasSince := !query.Since.IsZero()
	hasUntil := !query.Until.IsZero()
	switch {
	case hasSince && hasUntil:
		// Since && Until: time >= timestamp AND time <= timestamp
		// include prev and next
		prevQ, err := this.queries.StatePrevQuery(query.IDs, query.Kind, query.Since)
		if err != nil {
			return "", 0, 0, 0, err
		}
		seriesQ, err := this.queries.StatesTimeGrtEqLesEqQuery(query.IDs, query.Kind, query.Since, query.Until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		nextQ, err := this.queries.StateNextQuery(query.IDs, query.Kind, query.Until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		return prevQ + seriesQ + nextQ, 0, 1, 2, nil
	case hasRange && hasUntil:
		// Range && Until: time >= (timestamp - duration) AND time <= timestamp
		// include prev and next
		since := query.Until.Add(time.Duration(query.Range) * -1)
		prevQ, err := this.queries.StatePrevQuery(query.IDs, query.Kind, since)
		if err != nil {
			return "", 0, 0, 0, err
		}
		seriesQ, err := this.queries.StatesTimeGrtEqLesEqQuery(query.IDs, query.Kind, since, query.Until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		nextQ, err := this.queries.StateNextQuery(query.IDs, query.Kind, query.Until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		return prevQ + seriesQ + nextQ, 0, 1, 2, nil
	case hasRange && hasSince:
		// Range && Since: time >= timestamp AND time <= (timestamp + duration)
		// include prev and next
		until := query.Since.Add(time.Duration(query.Range))
		prevQ, err := this.queries.StatePrevQuery(query.IDs, query.Kind, query.Since)
		if err != nil {
			return "", 0, 0, 0, err
		}
		seriesQ, err := this.queries.StatesTimeGrtEqLesEqQuery(query.IDs, query.Kind, query.Since, until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		nextQ, err := this.queries.StateNextQuery(query.IDs, query.Kind, until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		return prevQ + seriesQ + nextQ, 0, 1, 2, nil
	case hasRange:
		// Range: time >= (now - duration)
		// include prev
		timestamp := getCurrentTime(this.config.InfluxdbUseUTC).Add(time.Duration(query.Range) * -1)
		prevQ, err := this.queries.StatePrevQuery(query.IDs, query.Kind, timestamp)
		if err != nil {
			return "", 0, 0, 0, err
		}
		seriesQ, err := this.queries.StatesTimeGrtEqQuery(query.IDs, query.Kind, timestamp)
		if err != nil {
			return "", 0, 0, 0, err
		}
		return prevQ + seriesQ, 0, 1, -1, nil
	case hasUntil:
		// Until: time <= timestamp
		// include next
		seriesQ, err := this.queries.StatesTimeLesEqQuery(query.IDs, query.Kind, query.Until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		nextQ, err := this.queries.StateNextQuery(query.IDs, query.Kind, query.Until)
		if err != nil {
			return "", 0, 0, 0, err
		}
		return seriesQ + nextQ, -1, 0, 1, nil
	case hasSince:
		// Since: time >= timestamp
		// include prev
		prevQ, err := this.queries.StatePrevQuery(query.IDs, query.Kind, query.Since)
		if err != nil {
			return "", 0, 0, 0, err
		}
		seriesQ, err := this.queries.StatesTimeGrtEqQuery(query.IDs, query.Kind, query.Since)
		if err != nil {
			return "", 0, 0, 0, err
		}
		return prevQ + seriesQ, 0, 1, -1, nil
	default:
		seriesQ, err := this.queries.StatesTimeLesEqQuery(query.IDs, query.Kind, getCurrentTime(this.config.InfluxdbUseUTC))
		if err != nil {
			return "", 0, 0, 0, err
		}
		return seriesQ, -1, 0, -1, nil
	}
}

func getCurrentTime(utc bool) time.Time {
	if utc {
		return time.Now().UTC()
	}
	return time.Now()
}
