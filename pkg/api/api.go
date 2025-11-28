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

package api

import (
	"log"
	"net/http"

	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	deviceRepo "github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/julienschmidt/httprouter"
)

var routes = []func(ctrl *controller.Controller, dr deviceRepo.Interface) (m, p string, h httprouter.Handle){
	PostCheckDeviceOnlineStates,
	PostInternCheckDeviceOnlineStates,
	PostInternCheckGatewayOnlineStates,
	PostInternGetDevicesHistory,
	PostInternGetGatewaysHistory,
	PostInternGetDevicesLogStart,
	PostInternGetGatewaysLogStart,
	PostInternGetDevicesLogEdge,
	PostInternGetGatewaysLogEdge,
	GetCurrentDeviceState,
	GetCurrentGatewayState,
	PostQueryBaseStatesMap,
	PostQueryWithAttributeFilterMapOriginal,
	PostQueryBaseStatesList,
	GetHistoricalDeviceStates,
	GetHistoricalGatewayStates,
	PostQueryHistoricalStatesMap,
	PostQueryHistoricalStatesMapOriginal,
	PostQueryHistoricalStatesList,
	GetSwaggerDoc,
}

// StartRest
// @title Connection Log API
// @Version {version}
// @description Provides HTTP-API to request current and historical connection log.
// @license.name Apache-2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @BasePath /
func StartRest(config configuration.Config, ctrl *controller.Controller) {
	log.Println("start server on port: ", config.ServerPort)
	router := httprouter.New()
	dr := deviceRepo.NewClient(config.DeviceRepoUrl, nil)
	for _, rf := range routes {
		m, p, hf := rf(ctrl, dr)
		router.Handle(m, p, hf)
		log.Println("added route:", m, p)
	}
	corseHandler := util.NewCors(router)
	logger := util.NewLogger(corseHandler)
	log.Println(http.ListenAndServe(":"+config.ServerPort, logger))
}
