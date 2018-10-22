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
	"github.com/SmartEnergyPlatform/jwt-http-router"
)

func CheckRightList(jwthttp jwt_http_router.JwtImpersonate, kind string, ids []string, right string) (ok bool, err error) {
	oks := map[string]bool{}
	err = jwthttp.PostJSON(Config.PermissionsUrl+"/ids/check/"+kind+"/"+right, ids, &oks)
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
