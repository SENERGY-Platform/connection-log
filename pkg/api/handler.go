package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	_ "github.com/SENERGY-Platform/connection-log/docs"
	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	deviceRepo "github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/julienschmidt/httprouter"
	"github.com/swaggo/swag"
)

const (
	pathParamID     = "id"
	queryParamRange = "range"
	queryParamSince = "since"
	queryParamUntil = "until"
)

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
func GetCurrentDeviceState(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodGet, "/current/devices/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		kind, err := controller.GetKindFromId(id, false)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if kind != model.DeviceKind {
			http.Error(writer, "devices endpoint only handles devices", http.StatusBadRequest)
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
func GetCurrentGatewayState(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodGet, "/current/gateways/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		kind, err := controller.GetKindFromId(id, false)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if kind != model.GatewayKind {
			http.Error(writer, "gateways endpoint only handles gateways", http.StatusBadRequest)
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

// PostQueryBaseStatesMap godoc
// @Summary Query current states
// @Description Query current states for multiple IDs (supported: devices, gateways/hubs, device-groups, locations).
// @Tags Current states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryWithAttributeFilter true "query object, attribute value and origin will only be checked if set, otherwise all values or origins will be blacklisted"
// @Success	200 {object} map[string]bool "current states mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/query/map [post]
func PostQueryBaseStatesMap(ctrl *controller.Controller, dr deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/current/query/map", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryWithAttributeFilter
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		query.IDs, err = ctrl.PermissionsFilterIDs(token, query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		query.IDs, err = resolveDeviceIds(dr, token, query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		query.IDs, err = filterDevices(dr, token, query.IDs, query.DeviceAttributeBlacklist)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := ctrl.QueryBaseStatesMap(request.Context(), query.QueryBase)
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

// PostQueryBaseStatesList godoc
// @Summary Query current states
// @Description Query current states for multiple IDs (supported: devices, gateways/hubs, device-groups, locations).
// @Tags Current states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryWithAttributeFilter true "query object, attribute value and origin will only be checked if set, otherwise all values or origins will be blacklisted"
// @Success	200 {array} model.ResourceCurrentState "current states"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /current/query/list [post]
func PostQueryBaseStatesList(ctrl *controller.Controller, dr deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/current/query/list", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryWithAttributeFilter
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		query.IDs, err = ctrl.PermissionsFilterIDs(token, query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		query.IDs, err = resolveDeviceIds(dr, token, query.IDs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		query.IDs, err = filterDevices(dr, token, query.IDs, query.DeviceAttributeBlacklist)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := ctrl.QueryBaseStatesSlice(request.Context(), query.QueryBase)
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
func GetHistoricalDeviceStates(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodGet, "/historical/devices/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		kind, err := controller.GetKindFromId(id, false)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if kind != model.DeviceKind {
			http.Error(writer, "devices endpoint only handles devices", http.StatusBadRequest)
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
func GetHistoricalGatewayStates(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodGet, "/historical/gateways/:" + pathParamID, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName(pathParamID)
		if id == "" {
			http.Error(writer, "missing id parameter", http.StatusBadRequest)
			return
		}
		kind, err := controller.GetKindFromId(id, false)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if kind != model.GatewayKind {
			http.Error(writer, "gateways endpoint only handles gateways", http.StatusBadRequest)
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
// @Description Query current historical states for multiple IDs (supported: devices, gateways/hubs).
// @Tags Historical states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryHistorical true "query object"
// @Success	200 {object} map[string]model.HistoricalStates "historical states mapped to IDs"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /historical/query/map [post]
func PostQueryHistoricalStatesMap(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/historical/query/map", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryHistorical
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), query.IDs)
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
// @Description Query current historical states for multiple IDs (supported: devices, gateways/hubs).
// @Tags Historical states
// @Accept json
// @Produce	json
// @Security Bearer
// @Param query body model.QueryHistorical true "query object"
// @Success	200 {array} model.ResourceHistoricalStates "historical states"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /historical/query/list [post]
func PostQueryHistoricalStatesList(ctrl *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
	return http.MethodPost, "/historical/query/list", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var query model.QueryHistorical
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		query.IDs, err = ctrl.PermissionsFilterIDs(util.GetAuthToken(request), query.IDs)
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
func GetSwaggerDoc(_ *controller.Controller, _ deviceRepo.Interface) (string, string, httprouter.Handle) {
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

func resolveDeviceIds(deviceRepoClient deviceRepo.Interface, token string, originalIds []string) (result []string, err error) {
	ids := slices.Clone(originalIds)
	result = []string{}
	for _, id := range ids {
		if strings.HasPrefix(id, models.DEVICE_GROUP_PREFIX) {
			deviceGroup, err, _ := deviceRepoClient.ReadDeviceGroup(id, token, false)
			if err != nil {
				return nil, err
			}
			result = append(result, deviceGroup.DeviceIds...)
		} else if strings.HasPrefix(id, models.LOCATION_PREFIX) {
			location, err, _ := deviceRepoClient.GetLocation(id, token)
			if err != nil {
				return nil, err
			}
			result = append(result, location.DeviceIds...)
			for _, deviceGroupId := range location.DeviceGroupIds {
				deviceGroup, err, _ := deviceRepoClient.ReadDeviceGroup(deviceGroupId, token, false)
				if err != nil {
					return nil, err
				}
				result = append(result, deviceGroup.DeviceIds...)
			}
		} else {
			result = append(result, id)
		}
	}
	return result, nil
}

func filterDevices(deviceRepoClient deviceRepo.Interface, token string, ids []string, deviceAttributeBlacklist []models.Attribute) (filteredIds []string, err error) {
	if len(deviceAttributeBlacklist) == 0 {
		return ids, nil
	}
	filteredIds = []string{}
	deviceIds := []string{}
	for _, id := range ids {
		if !strings.HasPrefix(id, models.DEVICE_PREFIX) {
			// never filtered
			filteredIds = append(filteredIds, id)
			continue
		}
		deviceIds = append(deviceIds, id)
	}
	if len(deviceIds) == 0 {
		return
	}
	devices, err, _ := deviceRepoClient.ListDevices(token, deviceRepo.DeviceListOptions{Ids: deviceIds})
	if err != nil {
		return nil, err
	}
outer:
	for _, device := range devices {
		for _, blackListAttribute := range deviceAttributeBlacklist {
			for _, deviceAttribute := range device.Attributes {
				if blackListAttribute.Key == deviceAttribute.Key &&
					(len(blackListAttribute.Value) == 0 || blackListAttribute.Value == deviceAttribute.Value) &&
					(len(blackListAttribute.Origin) == 0 || blackListAttribute.Origin == deviceAttribute.Origin) {
					continue outer
				}
			}
		}
		filteredIds = append(filteredIds, device.Id)
	}
	return
}
