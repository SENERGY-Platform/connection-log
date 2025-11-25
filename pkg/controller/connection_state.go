package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (this *Controller) GetCurrentState(ctx context.Context, id, kind string) (model.ResourceCurrentState, error) {
	if err := validateKind(kind); err != nil {
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

func (this *Controller) QueryBaseStatesSlice(ctx context.Context, query model.QueryBase) ([]model.ResourceCurrentState, error) {
	resMap, err := this.QueryBaseStatesMap(ctx, query)
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

func (this *Controller) QueryBaseStatesMap(ctx context.Context, query model.QueryBase) (map[string]bool, error) {
	idsBykind, err := GetIdsByKind(query.IDs, false)
	if err != nil {
		return nil, err
	}
	states := map[string]bool{}
	for kind, ids := range idsBykind {
		query.IDs = ids

		if err := validateKind(kind); err != nil {
			return nil, err
		}
		ctxWt, cf := context.WithTimeout(ctx, time.Duration(this.config.MongodbTimeout)*time.Second)
		defer cf()
		cursor, err := this.getMongoDBCollection(kind).Find(ctxWt, bson.M{kind: bson.M{"$in": query.IDs}})
		if err != nil {
			return nil, err
		}
		for cursor.Next(ctx) {
			var item State
			if err = cursor.Decode(&item); err != nil {
				return nil, err
			}
			if kind == model.GatewayKind {
				states[item.GatewayID] = item.Online
			} else {
				states[item.DeviceID] = item.Online
			}
		}
		if err = cursor.Err(); err != nil {
			return nil, err
		}
	}
	return states, nil
}

var permKindMap = map[string]string{
	model.DeviceKind:  model.PermDeviceKind,
	model.GatewayKind: model.PermGatewayKind,
}

func validateKind(kind string) error {
	if kind == model.DeviceKind || kind == model.GatewayKind {
		return nil
	}
	return fmt.Errorf("invalid kind '%s'", kind)
}

func GetKindFromId(id string, perm bool) (kind string, err error) {
	if strings.HasPrefix(id, models.DEVICE_PREFIX) {
		if perm {
			return permKindMap[model.DeviceKind], nil
		}
		return model.DeviceKind, nil
	}
	if strings.HasPrefix(id, models.HUB_PREFIX) {
		if perm {
			return permKindMap[model.GatewayKind], nil
		}
		return model.GatewayKind, nil
	}
	if perm {
		if strings.HasPrefix(id, models.DEVICE_GROUP_PREFIX) {
			return model.PermDeviceGroupKind, nil
		}
		if strings.HasPrefix(id, models.LOCATION_PREFIX) {
			return model.PermLocationsKind, nil
		}
	}

	return "", fmt.Errorf("unsupported kind")
}

func GetIdsByKind(ids []string, perm bool) (idsByKind map[string][]string, err error) {
	idsByKind = map[string][]string{}
	for _, id := range ids {
		kind, err := GetKindFromId(id, perm)
		if err != nil {
			return nil, err
		}
		arr, ok := idsByKind[kind]
		if !ok {
			arr = []string{}
		}
		arr = append(arr, id)
		idsByKind[kind] = arr
	}
	return idsByKind, nil
}
