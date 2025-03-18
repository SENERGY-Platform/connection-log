/*
 * Copyright 2018 InfAI (CC SES)
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
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (this *Controller) CheckDeviceOnlineStates(ids []string) (result map[string]bool, err error) {
	result = map[string]bool{}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := this.getDeviceStateCollection().Find(ctx, bson.M{"device": bson.M{"$in": ids}})
	if err != nil {
		return result, err
	}
	for cursor.Next(context.Background()) {
		element := DeviceState{}
		err = cursor.Decode(&element)
		if err != nil {
			return nil, err
		}
		result[element.Device] = element.Online
	}
	err = cursor.Err()
	return
}

func (this *Controller) CheckGatewayOnlineStates(ids []string) (result map[string]bool, err error) {
	result = map[string]bool{}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := this.getGatewayStateCollection().Find(ctx, bson.M{"gateway": bson.M{"$in": ids}})
	if err != nil {
		return result, err
	}
	for cursor.Next(context.Background()) {
		element := GatewayState{}
		err = cursor.Decode(&element)
		if err != nil {
			return nil, err
		}
		result[element.Gateway] = element.Online
	}
	err = cursor.Err()
	return
}
