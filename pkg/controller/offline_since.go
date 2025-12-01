package controller

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
	influx "github.com/influxdata/influxdb1-client/v2"
)

func (this *Controller) GetOfflineSince(ctx context.Context, ids []string, kind string) ([]model.OfflineSinceResponse, error) {
	query, err := this.queries.OfflineSinceQuery(ids, kind)
	if err != nil {
		return nil, err
	}

	resp, err := this.influx.Query(influx.NewQuery(query, this.config.InfluxdbDb, "s"))
	if err != nil {
		return nil, err
	}

	err = resp.Error()
	if err != nil {
		return nil, err
	}

	result := []model.OfflineSinceResponse{}

	for _, res := range resp.Results {
		for _, series := range res.Series {
			for _, row := range series.Values {
				timestamp, ok := row[0].(json.Number)
				if !ok {
					continue
				}
				i64, err := timestamp.Int64()
				if err != nil {
					continue
				}
				parsedTime := time.Unix(i64, 0)
				id, ok := row[1].(string)
				if !ok {
					continue
				}
				result = append(result, model.OfflineSinceResponse{ID: id, OfflineSince: parsedTime})
			}
		}
	}
	slices.SortFunc(result, func(a, b model.OfflineSinceResponse) int {
		return a.OfflineSince.Compare(b.OfflineSince)
	})
	return result, nil

}
