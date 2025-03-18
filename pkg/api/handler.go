package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/url"
	"time"
)

func PostCheckDeviceOnlineStates(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/state/device/check", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "deviceinstance", ids, "r")
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

func PostInternCheckDeviceOnlineStates(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/state/device/check", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "deviceinstance", ids, "r")
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

func PostInternCheckGatewayOnlineStates(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/state/gateway/check", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "gatewayinstance", ids, "r")
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

func PostInternGetDevicesHistory(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/history/device/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		duration := ps.ByName("duration")
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "deviceinstance", ids, "r")
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

func PostInternGetGatewaysHistory(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/history/gateway/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		duration := ps.ByName("duration")
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "gatewayinstance", ids, "r")
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

func PostInternGetDevicesLogStart(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logstarts/device", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "deviceinstance", ids, "r")
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

func PostInternGetGatewaysLogStart(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logstarts/gateway", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "gatewayinstance", ids, "r")
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

func PostInternGetDevicesLogEdge(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logedge/device/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		duration := ps.ByName("duration")
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "deviceinstance", ids, "r")
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

func PostInternGetGatewaysLogEdge(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/intern/logedge/gateway/:duration", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		duration := ps.ByName("duration")
		ids := []string{}
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(r), "gatewayinstance", ids, "r")
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

func GetCurrentDeviceState(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/current/devices/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), "deviceinstance", []string{id}, "r")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(writer, "access denied", http.StatusUnauthorized)
			return
		}
		res, err := ctrl.GetCurrentState(request.Context(), id, model.DeviceKind)
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

func GetCurrentGatewayState(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/current/gateways/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), "gatewayinstance", []string{id}, "r")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(writer, "access denied", http.StatusUnauthorized)
			return
		}
		res, err := ctrl.GetCurrentState(request.Context(), id, model.GatewayKind)
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

func PostQueryCurrentStatesMap(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/current/query/map", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryCurrent
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), query.Kind+"instance", query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := ctrl.QueryCurrentStatesMap(request.Context(), query)
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

func PostQueryCurrentStatesList(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/current/query/list", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryCurrent
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), query.Kind+"instance", query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := ctrl.QueryCurrentStatesSlice(request.Context(), query)
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

func GetHistoricalDeviceStates(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/historical/devices/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), "deviceinstance", []string{id}, "r")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(writer, "access denied", http.StatusUnauthorized)
			return
		}
		rng, since, until, err := parseHistoricalStatesQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := ctrl.GetHistoricalStates(request.Context(), id, model.DeviceKind, rng, since, until)
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

func GetHistoricalGatewayStates(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/historical/gateways/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), "gatewayinstance", []string{id}, "r")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(writer, "access denied", http.StatusUnauthorized)
			return
		}
		rng, since, until, err := parseHistoricalStatesQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := ctrl.GetHistoricalStates(request.Context(), id, model.GatewayKind, rng, since, until)
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

func PostQueryHistoricalStatesMap(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/historical/query/map", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryHistorical
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), query.Kind+"instance", query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := ctrl.QueryHistoricalStatesMap(request.Context(), query)
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

func PostQueryHistoricalStatesList(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/historical/query/list", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryHistorical
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), query.Kind+"instance", query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := ctrl.QueryHistoricalStatesSlice(request.Context(), query)
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

func parseHistoricalStatesQuery(query url.Values) (rng time.Duration, since time.Time, until time.Time, err error) {
	if rngStr := query.Get("range"); rngStr != "" {
		rng, err = time.ParseDuration(rngStr)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
	}
	if sinceStr := query.Get("since"); sinceStr != "" {
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
	}
	if untilStr := query.Get("until"); untilStr != "" {
		since, err = time.Parse(time.RFC3339, untilStr)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
	}
	return
}
