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
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestStartAuthenticates needs a throwaway server with access control; MONGO_AUTH_TEST_USER and
// MONGO_AUTH_TEST_PASSWORD are root credentials, used to create and remove the test users. This
// service only ever reads, so its production user gets the `read` role rather than `readWrite`.
func TestStartAuthenticates(t *testing.T) {
	url, rootUser, rootPassword := os.Getenv("MONGO_AUTH_TEST_URL"), os.Getenv("MONGO_AUTH_TEST_USER"), os.Getenv("MONGO_AUTH_TEST_PASSWORD")
	if testing.Short() || url == "" || rootUser == "" || rootPassword == "" {
		t.Skip("needs MONGO_AUTH_TEST_URL, MONGO_AUTH_TEST_USER and MONGO_AUTH_TEST_PASSWORD, not in -short")
	}
	ctx := context.Background()
	root, err := mongo.Connect(ctx, options.Client().ApplyURI(url).SetAuth(options.Credential{Username: rootUser, Password: rootPassword, AuthSource: "admin"}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Disconnect(ctx) })

	suffix := randomHex(t)
	testDB, otherDB := "connectionlog_auth_test_"+suffix, "connectionlog_auth_other_"+suffix
	readUser, readPassword := "connectionlog-read-"+suffix, randomHex(t)
	otherUser, otherPassword := "connectionlog-other-"+suffix, randomHex(t)
	createUser(t, root, readUser, readPassword, "read", testDB)
	createUser(t, root, otherUser, otherPassword, "read", otherDB)
	passwords := []string{readPassword, otherPassword, rootPassword}

	// the service reads what connection-log-worker writes; seed the document as root, standing in for it.
	seedID := "device-" + suffix
	if _, err := root.Database(testDB).Collection("devicestate").InsertOne(ctx, bson.M{"device": seedID, "online": true}); err != nil {
		t.Fatal(err)
	}

	config := func(user, password string) configuration.Config {
		return configuration.Config{
			MongoUrl:               url,
			MongoUser:              user,
			MongoPassword:          password,
			MongoAuthSource:        "admin",
			MongoDatabase:          testDB,
			MongodbTimeout:         10,
			DeviceStateCollection:  "devicestate",
			GatewayStateCollection: "gatewaystate",
			// New also builds an influx client; it must be a well-formed address, though nothing connects to it here.
			InfluxdbUrl: "http://localhost:1",
		}
	}

	t.Run("New with read-only credentials can read", func(t *testing.T) {
		ctrl, err := New(config(readUser, readPassword))
		if err != nil {
			t.Fatal(err)
		}
		defer ctrl.Close()
		state, err := ctrl.GetCurrentState(ctx, seedID, model.DeviceKind)
		if err != nil {
			t.Fatalf("read as the read-only service user: %v", err)
		}
		if !state.Connected {
			t.Errorf("state.Connected = false, want true")
		}
	})

	cases := []struct {
		name, user, password string
		wantErr              string
	}{
		{"read-only user", readUser, readPassword, ""},
		{"no credentials", "", "", "mongo startup check failed: "},
		{"user of another database", otherUser, otherPassword, "mongo startup check failed: "},
		{"wrong password", readUser, readPassword + "-wrong", "mongo startup check failed: "},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := config(c.user, c.password)
			if err := validateConfig(conf); err != nil {
				t.Fatal(err)
			}
			client, err := connect(ctx, clientOptions(conf), conf.MongoDatabase, 10*time.Second)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				_ = client.Disconnect(context.Background())
				return
			}
			if err == nil {
				_ = client.Disconnect(context.Background())
				t.Fatalf("expected an error containing %q", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("unexpected error: %v", err)
			}
			for _, pw := range passwords {
				if strings.Contains(err.Error(), pw) {
					t.Error("error text contains a password")
				}
			}
		})
	}
}

// createUser registers the cleanup first, so a partly failed creation is removed as well.
func createUser(t *testing.T, root *mongo.Client, user, password, role, db string) {
	t.Helper()
	admin := root.Database("admin")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = admin.RunCommand(ctx, bson.D{{Key: "dropUser", Value: user}}).Err()
		_ = root.Database(db).Drop(ctx)
	})
	cmd := bson.D{
		{Key: "createUser", Value: user},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: bson.A{bson.D{{Key: "role", Value: role}, {Key: "db", Value: db}}}},
	}
	if err := admin.RunCommand(context.Background(), cmd).Err(); err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func randomHex(t *testing.T) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}
