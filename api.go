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

package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SmartEnergyPlatform/util/http/response"

	"github.com/SmartEnergyPlatform/util/http/cors"
	"github.com/SmartEnergyPlatform/util/http/logger"

	"github.com/SmartEnergyPlatform/jwt-http-router"
)

func StartRest() {
	log.Println("start server on port: ", Config.ServerPort)
	httpHandler := getRoutes()
	corseHandler := cors.New(httpHandler)
	logger := logger.New(corseHandler, Config.LogLevel)
	log.Println(http.ListenAndServe(":"+Config.ServerPort, logger))
}

func getRoutes() (router *jwt_http_router.Router) {
	router = jwt_http_router.New(jwt_http_router.JwtConfig{ForceAuth: Config.ForceAuth == "true", ForceUser: Config.ForceUser == "true"})

	router.POST("/state/device/check", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := CheckRightList(jwt.Impersonate, "deviceinstance", ids, "r")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := checkDeviceOnlineStates(ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/state/device/check", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := checkDeviceOnlineStates(ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/state/gateway/check", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := checkGatewayOnlineStates(ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/history/device/:duration", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		duration := ps.ByName("duration")
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := getResourcesHistory(ids, "device", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/history/gateway/:duration", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		duration := ps.ByName("duration")
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := getResourcesHistory(ids, "gateway", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/logstarts/device", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := getResourcesLogstart(ids, "device")
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/logstarts/gateway", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := getResourcesLogstart(ids, "gateway")
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/logedge/device/:duration", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		duration := ps.ByName("duration")
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := getResourcesLogEdge(ids, "device", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	router.POST("/intern/logedge/gateway/:duration", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		duration := ps.ByName("duration")
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := getResourcesLogEdge(ids, "gateway", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(result)
	})

	return
}
