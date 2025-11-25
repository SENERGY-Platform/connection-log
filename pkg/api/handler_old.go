package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	deviceRepo "github.com/SENERGY-Platform/device-repository/lib/client"
	_ "github.com/influxdata/influxdb1-client/v2"
	"github.com/julienschmidt/httprouter"
)

// PostCheckDeviceOnlineStates godoc
// @Summary Check device online states
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param ids body []string true "list of IDs"
// @Success	200 {object} map[string]bool "states mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /state/device/check [post]
func PostCheckDeviceOnlineStates(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/state/device/check", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.DeviceKind {
				http.Error(res, "devices endpoint only handles devices", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.CheckDeviceOnlineStates(ids)
		if err != nil {
			log.Println("ERROR: while checking online states", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternCheckDeviceOnlineStates godoc
// @Summary Intern check device online states
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param ids body []string true "list of IDs"
// @Success	200 {object} map[string]bool "states mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/state/device/check [post]
func PostInternCheckDeviceOnlineStates(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/state/device/check", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.DeviceKind {
				http.Error(res, "devices endpoint only handles devices", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.CheckDeviceOnlineStates(ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternCheckGatewayOnlineStates godoc
// @Summary Intern check gateway online states
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param ids body []string true "list of IDs"
// @Success	200 {object} map[string]bool "states mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/state/gateway/check [post]
func PostInternCheckGatewayOnlineStates(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/state/gateway/check", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.GatewayKind {
				http.Error(res, "gateways endpoint only handles gateways", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.CheckGatewayOnlineStates(ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternGetDevicesHistory godoc
// @Summary Intern get devices history
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param ids body []string true "list of IDs"
// @Param duration path string true "duration in influxdb format https://docs.influxdata.com/influxdb/v1.5/query_language/spec/#durations"
// @Success	200 {array} client.Result "result"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/history/device/{duration} [post]
func PostInternGetDevicesHistory(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/history/device/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		duration := ps.ByName("duration")
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.DeviceKind {
				http.Error(res, "devices endpoint only handles devices", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.GetResourcesHistory(ids, "device", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternGetGatewaysHistory godoc
// @Summary Intern get gateways history
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param ids body []string true "list of IDs"
// @Param duration path string true "duration in influxdb format https://docs.influxdata.com/influxdb/v1.5/query_language/spec/#durations"
// @Success	200 {array} client.Result "result"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/history/gateway/{duration} [post]
func PostInternGetGatewaysHistory(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/history/gateway/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		duration := ps.ByName("duration")
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.GatewayKind {
				http.Error(res, "gateways endpoint only handles gateways", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.GetResourcesHistory(ids, "gateway", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternGetDevicesLogStart godoc
// @Summary Intern get devices log start
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param ids body []string true "list of IDs"
// @Success	200 {object} map[string]float64 "unix timestamps mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/logstarts/device [post]
func PostInternGetDevicesLogStart(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logstarts/device", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.DeviceKind {
				http.Error(res, "devices endpoint only handles devices", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.GetResourcesLogstart(ids, "device")
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternGetGatewaysLogStart godoc
// @Summary Intern get gateways log start
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param ids body []string true "list of IDs"
// @Success	200 {object} map[string]float64 "unix timestamps mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/logstarts/gateway [post]
func PostInternGetGatewaysLogStart(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logstarts/gateway", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.GatewayKind {
				http.Error(res, "gateways endpoint only handles gateways", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.GetResourcesLogstart(ids, "gateway")
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternGetDevicesLogEdge godoc
// @Summary Intern get devices log edge
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param duration path string true "duration in influxdb format https://docs.influxdata.com/influxdb/v1.5/query_language/spec/#durations"
// @Param ids body []string true "list of IDs"
// @Success	200 {object} map[string][][]any ""
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/logedge/device/{duration} [post]
func PostInternGetDevicesLogEdge(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logedge/device/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		duration := ps.ByName("duration")
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.DeviceKind {
				http.Error(res, "devices endpoint only handles devices", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.GetResourcesLogEdge(ids, "device", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}

// PostInternGetGatewaysLogEdge godoc
// @Summary Intern get gateways log edge
// @Tags Old api
// @Accept json
// @Produce	json
// @Security Bearer
// @Param duration path string true "duration in influxdb format https://docs.influxdata.com/influxdb/v1.5/query_language/spec/#durations"
// @Param ids body []string true "list of IDs"
// @Success	200 {object} map[string][][]any ""
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /intern/logedge/gateway/{duration} [post]
func PostInternGetGatewaysLogEdge(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logedge/gateway/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		duration := ps.ByName("duration")
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		for _, id := range ids {
			kind, err := controller.GetKindFromId(id, false)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if kind != model.GatewayKind {
				http.Error(res, "gateways endpoint only handles gateways", http.StatusBadRequest)
				return
			}
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), ids, "r")
		if err != nil {
			log.Println("ERROR: while checking rights", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.GetResourcesLogEdge(ids, "gateway", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}
