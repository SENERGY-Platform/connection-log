package controller

import (
	"context"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (this *Controller) GetCurrentState(ctx context.Context, id, kind string) (model.ResourceCurrentState, error) {
	collection, err := this.getMongoDBCollection(kind)
	if err != nil {
		return model.ResourceCurrentState{}, err
	}
	ctxWt, cf := context.WithTimeout(ctx, time.Duration(this.config.MongodbTimeout)*time.Second)
	defer cf()
	res := collection.FindOne(ctxWt, bson.M{kind: id})
	if err = res.Err(); err != nil {
		return model.ResourceCurrentState{}, err
	}
	var item DeviceState
	if err = res.Decode(&item); err != nil {
		return model.ResourceCurrentState{}, err
	}
	return model.ResourceCurrentState{
		ID:        item.Device,
		Connected: item.Online,
	}, nil
}

func (this *Controller) QueryCurrentStatesSlice(ctx context.Context, query model.QueryCurrent) ([]model.ResourceCurrentState, error) {
	resMap, err := this.QueryCurrentStatesMap(ctx, query)
	if err != nil {
		return nil, err
	}
	sl := make([]model.ResourceCurrentState, 0, len(resMap))
	for id, state := range resMap {
		sl = append(sl, model.ResourceCurrentState{
			ID:        id,
			Connected: state,
		})
	}
	return sl, nil
}

func (this *Controller) QueryCurrentStatesMap(ctx context.Context, query model.QueryCurrent) (map[string]bool, error) {
	collection, err := this.getMongoDBCollection(query.Kind)
	if err != nil {
		return nil, err
	}
	ctxWt, cf := context.WithTimeout(ctx, time.Duration(this.config.MongodbTimeout)*time.Second)
	defer cf()
	cursor, err := collection.Find(ctxWt, bson.M{query.Kind: bson.M{"$in": query.IDs}})
	if err != nil {
		return nil, err
	}
	states := make(map[string]bool)
	for cursor.Next(ctx) {
		var item DeviceState
		if err = cursor.Decode(&item); err != nil {
			return nil, err
		}
		states[item.Device] = item.Online
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return states, nil
}
