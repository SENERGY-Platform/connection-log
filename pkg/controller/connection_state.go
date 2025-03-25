package controller

import (
	"context"
	"fmt"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (this *Controller) GetCurrentState(ctx context.Context, id, kind string) (model.ResourceCurrentState, error) {
	if err := ValidateKind(kind); err != nil {
		return model.ResourceCurrentState{}, err
	}
	ctxWt, cf := context.WithTimeout(ctx, time.Duration(this.config.MongodbTimeout)*time.Second)
	defer cf()
	res := this.getMongoDBCollection(kind).FindOne(ctxWt, bson.M{kind: id})
	if err := res.Err(); err != nil {
		return model.ResourceCurrentState{}, err
	}
	var item State
	if err := res.Decode(&item); err != nil {
		return model.ResourceCurrentState{}, err
	}
	return model.ResourceCurrentState{
		ID:        id,
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
	if err := ValidateKind(query.Kind); err != nil {
		return nil, err
	}
	ctxWt, cf := context.WithTimeout(ctx, time.Duration(this.config.MongodbTimeout)*time.Second)
	defer cf()
	cursor, err := this.getMongoDBCollection(query.Kind).Find(ctxWt, bson.M{query.Kind: bson.M{"$in": query.IDs}})
	if err != nil {
		return nil, err
	}
	states := make(map[string]bool)
	for cursor.Next(ctx) {
		var item State
		if err = cursor.Decode(&item); err != nil {
			return nil, err
		}
		if query.Kind == model.GatewayKind {
			states[item.GatewayID] = item.Online
		} else {
			states[item.DeviceID] = item.Online
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return states, nil
}

func ValidateKind(kind string) error {
	if kind == model.DeviceKind || kind == model.GatewayKind {
		return nil
	}
	return fmt.Errorf("invalid kind '%s'", kind)
}
