package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"github.com/influxdata/influxdb1-client/models"
	influx "github.com/influxdata/influxdb1-client/v2"
	"log"
	"time"
)

func (this *Controller) GetResourcesStates(ids []string, kind string, duration time.Duration, includeInit, includePrev bool) ([]model.States, error) {
	if len(ids) == 0 {
		return []model.States{}, nil
	}
	timestamp := time.Now().UTC().Add(duration * -1)
	statement, err := this.queries.GetResStatesRangeQuery(ids, kind, timestamp)
	if err != nil {
		return []model.States{}, err
	}
	var initStmtID int
	if includeInit {
		stmt, err := this.queries.GetResStatesInitQuery(ids, kind)
		if err != nil {
			return []model.States{}, err
		}
		statement += stmt
		initStmtID = 1
	}
	var prevStmtID int
	if includePrev {
		stmt, err := this.queries.GetResStatesPrevQuery(ids, kind, timestamp)
		if err != nil {
			return []model.States{}, err
		}
		statement += stmt
		prevStmtID = initStmtID + 1
	}
	resp, err := this.influx.Query(influx.NewQuery(statement, this.config.InfluxdbDb, "s"))
	if err != nil {
		return []model.States{}, err
	}
	if err = resp.Error(); err != nil {
		return []model.States{}, err
	}
	resMap, err := handleResults(resp.Results, kind, initStmtID, prevStmtID)
	if err != nil {
		return []model.States{}, err
	}
	result := make([]model.States, 0, len(resMap))
	for _, resource := range resMap {
		result = append(result, resource)
	}
	return result, nil
}

func handleResults(results []influx.Result, kind string, initStmtID, prevStmtID int) (map[string]model.States, error) {
	if len(results) == 0 {
		return nil, errors.New("no results")
	}
	resMap := make(map[string]model.States)
	for _, result := range results {
		if result.Err != "" {
			return nil, errors.New(result.Err)
		}
		resType := 0
		if result.StatementId > 0 {
			switch result.StatementId {
			case initStmtID:
				resType = 1
			case prevStmtID:
				resType = 2
			}
		}
		handleSeries(resMap, kind, result.Series, resType)
	}
	return resMap, nil
}

func handleSeries(resMap map[string]model.States, kind string, series []models.Row, resType int) {
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

func handleRow(resMap map[string]model.States, rowValues [][]any, key string, resType int) error {
	resource, ok := resMap[key]
	if !ok {
		resource.ResourceID = key
	}
	if resType > 0 {
		state, err := rowItemToState(rowValues[0])
		if err != nil {
			return err
		}
		switch resType {
		case 1:
			resource.InitState = &state
		case 2:
			resource.PrevState = &state
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
