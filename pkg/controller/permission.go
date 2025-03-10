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
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"log"
)

func (this *Controller) CheckRightList(token string, kind string, ids []string, right string) (ok bool, err error) {
	oks, err := CheckAccess(this.config.PermissionsV2Url, token, kind, ids)
	if err != nil {
		return false, err
	}
	for _, element := range oks {
		if !element {
			return false, err
		}
	}
	return true, err
}

func (this *Controller) PermissionsFilterIDs(token string, kind string, IDs []string) ([]string, error) {
	result, err := CheckAccess(this.config.PermissionsV2Url, token, kind, IDs)
	if err != nil {
		return nil, err
	}
	var okIDs []string
	var nOkIDs []string
	for id, ok := range result {
		if ok {
			okIDs = append(okIDs, id)
		} else {
			nOkIDs = append(nOkIDs, id)
		}
	}
	if len(nOkIDs) > 0 {
		log.Printf("access denied for IDs: %v", nOkIDs)
	}
	return okIDs, nil
}

func CheckAccess(permV2Url string, token string, kind string, ids []string) (result map[string]bool, err error) {
	result, err, _ = client.New(permV2Url).CheckMultiplePermissions(token, kind, ids, client.Execute)
	if err != nil {
		return result, err
	}
	return result, nil
}

type QueryMessage struct {
	Resource string         `json:"resource"`
	Find     *QueryFind     `json:"find"`
	ListIds  *QueryListIds  `json:"list_ids"`
	CheckIds *QueryCheckIds `json:"check_ids"`
}
type QueryFind struct {
	QueryListCommons
	Search string     `json:"search"`
	Filter *Selection `json:"filter"`
}

type QueryListIds struct {
	QueryListCommons
	Ids []string `json:"ids"`
}

type QueryCheckIds struct {
	Ids    []string `json:"ids"`
	Rights string   `json:"rights"`
}

type QueryListCommons struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Rights   string `json:"rights"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
}

type QueryOperationType string

const (
	QueryEqualOperation             QueryOperationType = "=="
	QueryUnequalOperation           QueryOperationType = "!="
	QueryAnyValueInFeatureOperation QueryOperationType = "any_value_in_feature"
)

type ConditionConfig struct {
	Feature   string             `json:"feature"`
	Operation QueryOperationType `json:"operation"`
	Value     interface{}        `json:"value"`
	Ref       string             `json:"ref"`
}

type Selection struct {
	And       []Selection     `json:"and"`
	Or        []Selection     `json:"or"`
	Not       *Selection      `json:"not"`
	Condition ConditionConfig `json:"condition"`
}
