/*
 * Copyright 2026 InfAI (CC SES)
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
	"sort"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrGatewayHealthNotFound = errors.New("no health report for gateway")

// gatewayHealthDoc is the document connection-log-worker writes, one per hub.
type gatewayHealthDoc struct {
	HubID      string    `bson:"hub_id"`
	State      int       `bson:"state"`
	Reasons    []string  `bson:"reasons"`
	Report     bson.Raw  `bson:"report"`
	ReportTime string    `bson:"report_time"`
	Received   time.Time `bson:"received"`
}

func (this *Controller) getGatewayHealthCollection() *mongo.Collection {
	name := this.config.GatewayHealthCollection
	if name == "" {
		name = "gatewayhealth"
	}
	return this.mongo.Database(this.config.MongoDatabase).Collection(name)
}

// isStale reports whether a report received at the given time is older than maxAge.
func isStale(received time.Time, now time.Time, maxAge time.Duration) bool {
	return now.Sub(received) > maxAge
}

// toGatewayHealth converts the stored document; the report becomes plain JSON again instead of
// BSON key/value pairs.
func toGatewayHealth(doc gatewayHealthDoc, now time.Time, maxAge time.Duration) (model.GatewayHealth, error) {
	report := []byte("null")
	if len(doc.Report) > 0 {
		var err error
		report, err = bson.MarshalExtJSON(doc.Report, false, false)
		if err != nil {
			return model.GatewayHealth{}, fmt.Errorf("report of %s: %w", doc.HubID, err)
		}
	}
	reasons := doc.Reasons
	if reasons == nil {
		reasons = []string{}
	}
	return model.GatewayHealth{
		HubID:      doc.HubID,
		State:      doc.State,
		Reasons:    reasons,
		Report:     report,
		ReportTime: doc.ReportTime,
		Received:   doc.Received.UTC(),
		Stale:      isStale(doc.Received, now, maxAge),
	}, nil
}

// GetGatewayHealth returns ErrGatewayHealthNotFound if the hub has not sent a report.
func (this *Controller) GetGatewayHealth(ctx context.Context, hubID string) (model.GatewayHealth, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(this.config.MongodbTimeout)*time.Second)
	defer cancel()
	var doc gatewayHealthDoc
	err := this.getGatewayHealthCollection().FindOne(ctx, bson.M{"hub_id": hubID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.GatewayHealth{}, ErrGatewayHealthNotFound
	}
	if err != nil {
		return model.GatewayHealth{}, err
	}
	return toGatewayHealth(doc, time.Now(), this.gatewayHealthStaleAge)
}

// GetGatewayHealthList returns the reports of the given hubs, sorted by hub id; hubs without a report are omitted.
func (this *Controller) GetGatewayHealthList(ctx context.Context, hubIDs []string) ([]model.GatewayHealth, error) {
	result := []model.GatewayHealth{}
	if len(hubIDs) == 0 {
		return result, nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(this.config.MongodbTimeout)*time.Second)
	defer cancel()
	cursor, err := this.getGatewayHealthCollection().Find(ctx, bson.M{"hub_id": bson.M{"$in": hubIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	now := time.Now()
	for cursor.Next(ctx) {
		var doc gatewayHealthDoc
		if err = cursor.Decode(&doc); err != nil {
			return nil, err
		}
		item, err := toGatewayHealth(doc, now, this.gatewayHealthStaleAge)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	sort.Slice(result, func(i, j int) bool { return result[i].HubID < result[j].HubID })
	return result, nil
}
