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
	"flag"
	"github.com/SmartEnergyPlatform/connection-log/pkg/api"
	"github.com/SmartEnergyPlatform/connection-log/pkg/configuration"
	"github.com/SmartEnergyPlatform/connection-log/pkg/controller"
	"log"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()
	conf, err := configuration.Load(*configLocation)
	if err != nil {
		log.Fatal(err)
	}
	ctrl, err := controller.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	api.StartRest(conf, ctrl)
}
