package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "github.com/SENERGY-Platform/connection-log/docs"
	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	"github.com/julienschmidt/httprouter"
	"github.com/swaggo/swag"
)

const (
	pathParamID     = "id"
	queryParamRange = "range"
	queryParamSince = "since"
	queryParamUntil = "until"
)

var permKindMap = map[string]string{
	model.DeviceKind:  model.PermDeviceKind,
	model.GatewayKind: model.PermGatewayKind,
}

// GetCurrentDeviceState godoc
// @Summary Get current device state
// @Description Get the current state of a device.
// @Tags Current states
// @Produce	json
// @Security Bearer
// @Param id path string true "device id"
// @Success	200 {object} model.ResourceCurrentState "device state"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/devices/{id} [get]
func GetCurrentDeviceState(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/current/devices/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), model.PermDeviceKind, []string{id}, "r")
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

// GetCurrentGatewayState godoc
// @Summary Get current gateway state
// @Description Get the current state of a gateway.
// @Tags Current states
// @Produce	json
// @Security Bearer
// @Param id path string true "gateway id"
// @Success	200 {object} model.ResourceCurrentState "gateway state"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/gateways/{id} [get]
func GetCurrentGatewayState(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/current/gateways/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), model.PermGatewayKind, []string{id}, "r")
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

// PostQueryCurrentStatesMap godoc
// @Summary Query current states
// @Description Query current states for multiple IDs by resource kind (device, gateway).
// @Tags Current states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryCurrent true "query object"
// @Success	200 {object} map[string]bool "current states mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/query/map [post]
func PostQueryCurrentStatesMap(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/current/query/map", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryCurrent
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err = controller.ValidateKind(query.Kind); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), permKindMap[query.Kind], query.IDs)
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

// PostQueryCurrentStatesList godoc
// @Summary Query current states
// @Description Query current states for multiple IDs by resource kind (device, gateway).
// @Tags Current states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryCurrent true "query object"
// @Success	200 {array} model.ResourceCurrentState "current states"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/query/list [post]
func PostQueryCurrentStatesList(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/current/query/list", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryCurrent
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err = controller.ValidateKind(query.Kind); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), permKindMap[query.Kind], query.IDs)
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

// GetHistoricalDeviceStates godoc
// @Summary Get historical device states
// @Description Get the historical states of a device.
// @Tags Historical states
// @Produce	json
// @Security Bearer
// @Param id path string true "device id"
// @Param range query string false "time range e.g. 24h, valid units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'"
// @Param since query string false "timestamp in RFC 3339 format, can be combined with 'range' or 'until'"
// @Param until query string false "timestamp in RFC 3339 format, can be combined with 'range' or 'since'"
// @Success	200 {object} model.ResourceHistoricalStates "device state"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /historical/devices/{id} [get]
func GetHistoricalDeviceStates(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/historical/devices/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), model.PermDeviceKind, []string{id}, "r")
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

// GetHistoricalGatewayStates godoc
// @Summary Get historical gateway states
// @Description Get the historical states of a gateway.
// @Tags Historical states
// @Produce	json
// @Security Bearer
// @Param id path string true "gateway id"
// @Param range query string false "time range e.g. 24h, valid units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'"
// @Param since query string false "timestamp in RFC 3339 format, can be combined with 'range' or 'until'"
// @Param until query string false "timestamp in RFC 3339 format, can be combined with 'range' or 'since'"
// @Success	200 {object} model.ResourceHistoricalStates "gateway states"
// @Failure	400 {string} string "error message"
// @Failure	401 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /historical/gateways/{id} [get]
func GetHistoricalGatewayStates(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/historical/gateways/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		ok, err := ctrl.CheckRightList(util.GetAuthToken(request), model.PermGatewayKind, []string{id}, "r")
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

// PostQueryHistoricalStatesMap godoc
// @Summary Query historical states
// @Description Query current historical states for multiple IDs by resource kind (device, gateway).
// @Tags Historical states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryHistorical true "query object"
// @Success	200 {object} map[string]model.HistoricalStates "historical states mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /historical/query/map [post]
func PostQueryHistoricalStatesMap(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/historical/query/map", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryHistorical
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err = controller.ValidateKind(query.Kind); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), permKindMap[query.Kind], query.IDs)
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

// PostQueryHistoricalStatesList godoc
// @Summary Query historical states
// @Description Query current historical states for multiple IDs by resource kind (device, gateway).
// @Tags Historical states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryHistorical true "query object"
// @Success	200 {array} model.ResourceHistoricalStates "historical states"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /historical/query/list [post]
func PostQueryHistoricalStatesList(ctrl *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodPost, "/historical/query/list", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryHistorical
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err = controller.ValidateKind(query.Kind); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), permKindMap[query.Kind], query.IDs)
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

//go:generate go install github.com/swaggo/swag/cmd/swag@latest
//go:generate swag init -o ../../docs --parseDependency -d .. -g api/api.go
func GetSwaggerDoc(_ *controller.Controller) (string, string, httprouter.Handle) {
	return http.MethodGet, "/doc", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		doc, err := swag.ReadDoc()
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		//remove empty host to enable developer-swagger-api service to replace it; can not use cleaner delete on json object, because developer-swagger-api is sensible to formatting; better alternative is refactoring of developer-swagger-api/apis/db/db.py
		doc = strings.Replace(doc, `"host": "",`, "", 1)
		_, _ = writer.Write([]byte(doc))
	}
}

func parseHistoricalStatesQuery(query url.Values) (rng time.Duration, since time.Time, until time.Time, err error) {
	if rngStr := query.Get(queryParamRange); rngStr != "" {
		rng, err = time.ParseDuration(rngStr)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
	}
	if sinceStr := query.Get(queryParamSince); sinceStr != "" {
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
	}
	if untilStr := query.Get(queryParamUntil); untilStr != "" {
		since, err = time.Parse(time.RFC3339, untilStr)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
	}
	return
}
