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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	deviceRepo "github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/julienschmidt/httprouter"
)

// GetCurrentGatewayHealth godoc
// @Summary Get current gateway health
// @Description Get the latest health report of a gateway. The report is marked stale if it was received longer ago than the configured stale age.
// @Tags Gateway health
// @Produce	json
// @Security Bearer
// @Param id path string true "gateway id"
// @Success	200 {object} model.GatewayHealth "gateway health"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	404 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/gateways/{id}/health [get]
func GetCurrentGatewayHealth(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodGet, "/current/gateways/:" + pathParamID + "/health", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if err := validateGatewayIds([]string{id}); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), []string{id}, "r")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(writer, "access denied", http.StatusUnauthorized)
			return
		}
		res, err := ctrl.GetGatewayHealth(request.Context(), id)
		if errors.Is(err, controller.ErrGatewayHealthNotFound) {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err = json.NewEncoder(writer).Encode(res); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// PostQueryGatewayHealth godoc
// @Summary Query current gateway health
// @Description Query the latest health reports for multiple gateway IDs. If no IDs are provided, all accessible gateways are queried. Gateways the caller may not access are omitted, as are gateways without a report. The result is sorted by gateway ID.
// @Tags Gateway health
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryBase true "query object"
// @Success	200 {array} model.GatewayHealth "gateway health"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/query/health [post]
func PostQueryGatewayHealth(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/current/query/health", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryBase
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err = validateGatewayIds(query.IDs); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		ids := query.IDs
		if len(ids) == 0 {
			ids, err = ctrl.ListIds(token, model.PermGatewayKind)
		} else {
			ids, err = ctrl.PermissionsFilterIDs(token, ids)
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := ctrl.GetGatewayHealthList(request.Context(), ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err = json.NewEncoder(writer).Encode(res); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func validateGatewayIds(ids []string) error {
	for _, id := range ids {
		if id == "" {
			return errors.New("missing id parameter")
		}
		kind, err := controller.GetKindFromId(id, false)
		if err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
		if kind != model.GatewayKind {
			return fmt.Errorf("%s: health endpoints only handle gateways", id)
		}
	}
	return nil
}
