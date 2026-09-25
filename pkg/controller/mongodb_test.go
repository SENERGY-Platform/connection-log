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
	"errors"
	"testing"

	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
)

func TestValidateConfig(t *testing.T) {
	cases := []struct {
		name    string
		config  configuration.Config
		wantErr error
	}{
		{"database and user set", configuration.Config{MongoDatabase: "d", MongoUser: "u", MongoPassword: "p"}, nil},
		{"no user, no password needed", configuration.Config{MongoDatabase: "d"}, nil},
		{"empty database", configuration.Config{MongoUser: "u", MongoPassword: "p"}, errEmptyDatabase},
		{"user without password", configuration.Config{MongoDatabase: "d", MongoUser: "u"}, errMissingPassword},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateConfig(c.config)
			if !errors.Is(err, c.wantErr) {
				t.Errorf("validateConfig() = %v, want %v", err, c.wantErr)
			}
		})
	}
}

func TestClientOptionsAppliesCredentialsOnlyWhenUserSet(t *testing.T) {
	withoutUser := clientOptions(configuration.Config{MongoUrl: "mongodb://localhost:27017", MongoDatabase: "d"})
	if withoutUser.Auth != nil {
		t.Errorf("expected no auth without a user, got %+v", withoutUser.Auth)
	}

	withUser := clientOptions(configuration.Config{
		MongoUrl:        "mongodb://localhost:27017",
		MongoUser:       "u",
		MongoPassword:   "p",
		MongoAuthSource: "admin",
	})
	if withUser.Auth == nil {
		t.Fatal("expected auth to be set")
	}
	if withUser.Auth.Username != "u" || withUser.Auth.Password != "p" || withUser.Auth.AuthSource != "admin" {
		t.Errorf("unexpected auth: %+v", withUser.Auth)
	}
}
