package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
	"time"
)

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
