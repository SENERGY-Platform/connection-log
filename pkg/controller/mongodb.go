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
	"github.com/SmartEnergyPlatform/connection-log/pkg/configuration"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

func createIndexes(config configuration.Config, db *mongo.Client) error {
	err := createDeviceIndexes(config, db)
	if err != nil {
		return err
	}
	return createGatewayIndexes(config, db)
}

func createDeviceIndexes(config configuration.Config, db *mongo.Client) error {
	collection := db.Database(config.MongoTable).Collection(config.DeviceStateCollection)
	indexname := "device_1"
	indexkey := "device"
	direction := bsonx.Int32(1)
	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bsonx.Doc{{indexkey, direction}},
		Options: options.Index().SetName(indexname),
	})
	return err
}

func createGatewayIndexes(config configuration.Config, db *mongo.Client) error {
	collection := db.Database(config.MongoTable).Collection(config.DeviceStateCollection)
	indexname := "gateway_1"
	indexkey := "gateway"
	direction := bsonx.Int32(1)
	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bsonx.Doc{{indexkey, direction}},
		Options: options.Index().SetName(indexname),
	})
	return err
}

func (this *Controller) getDeviceStateCollection() *mongo.Collection {
	return this.mongo.Database(this.config.MongoTable).Collection(this.config.DeviceStateCollection)
}

func (this *Controller) getGatewayStateCollection()*mongo.Collection {
	return this.mongo.Database(this.config.MongoTable).Collection(this.config.GatewayStateCollection)
}

type DeviceState struct {
	Device string `json:"device,omitempty" bson:"device,omitempty"`
	Online bool   `json:"online" bson:"online"`
}

type GatewayState struct {
	Gateway string `json:"gateway,omitempty" bson:"gateway,omitempty"`
	Online  bool   `json:"online" bson:"online"`
}
