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
	"fmt"
)

func ExampleHistory() {
	fmt.Println(LoadConfig("config.json"))
	Config.InfluxdbUrl = "http://localhost:8004"

	result, err := getResourceHistory("iot#bc5eb8b8-f575-43b9-9637-5f7d988652b8", "device", "1h")
	fmt.Println(err)
	resultJson, err := json.Marshal(result)
	fmt.Println(string(resultJson), err)

	result, err = getResourceKindHistory("device", "1h")
	fmt.Println(err)
	resultJson, err = json.Marshal(result)
	fmt.Println(string(resultJson), err)

	result, err = getResourcesHistory([]string{"iot#bc5eb8b8-f575-43b9-9637-5f7d988652b8"}, "device", "7d")
	fmt.Println(err)
	resultJson, err = json.Marshal(result)
	fmt.Println(string(resultJson), err)

	result, err = getResourcesLogstart([]string{"iot#bc5eb8b8-f575-43b9-9637-5f7d988652b8"}, "device")
	fmt.Println(err)
	resultJson, err = json.Marshal(result)
	fmt.Println(string(resultJson), err)

	result, err = getResourcesLogEdge([]string{"iot#bc5eb8b8-f575-43b9-9637-5f7d988652b8"}, "device", "5h")
	fmt.Println(err)
	resultJson, err = json.Marshal(result)
	fmt.Println(string(resultJson), err)

	//Output:
	//
}
