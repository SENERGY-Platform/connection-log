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
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	errEmptyDatabase   = errors.New("mongo database name must not be empty")
	errMissingPassword = errors.New("mongo password must not be empty when a mongo user is set")
	errInvalidStaleAge = errors.New("gateway health stale age must be a positive duration")
)

func validateConfig(config configuration.Config) error {
	if config.MongoDatabase == "" {
		return errEmptyDatabase
	}
	if config.MongoUser != "" && config.MongoPassword == "" {
		return errMissingPassword
	}
	if _, err := gatewayHealthStaleAge(config); err != nil {
		return err
	}
	return nil
}

// gatewayHealthStaleAge parses the configured stale age; an empty value means the default.
func gatewayHealthStaleAge(config configuration.Config) (time.Duration, error) {
	value := config.GatewayHealthStaleAge
	if value == "" {
		value = configuration.DefaultGatewayHealthStaleAge
	}
	age, err := time.ParseDuration(value)
	if err != nil || age <= 0 {
		return 0, fmt.Errorf("%w: %q", errInvalidStaleAge, value)
	}
	return age, nil
}

// clientOptions applies the credentials after the URI so they replace any given in MongoUrl.
func clientOptions(config configuration.Config) *options.ClientOptions {
	opts := options.Client().ApplyURI(config.MongoUrl)
	if config.MongoUser != "" {
		opts.SetAuth(options.Credential{
			Username:   config.MongoUser,
			Password:   config.MongoPassword,
			AuthSource: config.MongoAuthSource,
		})
	}
	return opts
}

// connect runs listCollections on the service's database because Connect is lazy and ping needs no
// authentication; unreachable servers and wrong or missing credentials then fail at startup.
func connect(ctx context.Context, opts *options.ClientOptions, database string, timeout time.Duration) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	listOpts := options.ListCollections().SetNameOnly(true).SetAuthorizedCollections(true)
	if _, err = client.Database(database).ListCollectionNames(ctx, bson.D{}, listOpts); err != nil {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), timeout)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		return nil, fmt.Errorf("mongo startup check failed: %w", err)
	}
	return client, nil
}

func (this *Controller) getMongoDBCollection(kind string) *mongo.Collection {
	if kind == model.GatewayKind {
		return this.mongo.Database(this.config.MongoDatabase).Collection(this.config.GatewayStateCollection)
	}
	return this.mongo.Database(this.config.MongoDatabase).Collection(this.config.DeviceStateCollection)
}

func (this *Controller) getDeviceStateCollection() *mongo.Collection {
	return this.mongo.Database(this.config.MongoDatabase).Collection(this.config.DeviceStateCollection)
}

func (this *Controller) getGatewayStateCollection() *mongo.Collection {
	return this.mongo.Database(this.config.MongoDatabase).Collection(this.config.GatewayStateCollection)
}

type State struct {
	DeviceID  string `json:"device,omitempty" bson:"device,omitempty"`
	GatewayID string `json:"gateway,omitempty" bson:"gateway,omitempty"`
	Online    bool   `json:"online" bson:"online"`
}

type DeviceState struct {
	Device string `json:"device,omitempty" bson:"device,omitempty"`
	Online bool   `json:"online" bson:"online"`
}

type GatewayState struct {
	Gateway string `json:"gateway,omitempty" bson:"gateway,omitempty"`
	Online  bool   `json:"online" bson:"online"`
}
