package api

import (
	"encoding/json"
	"net/http"

	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/SENERGY-Platform/connection-log/pkg/model"
	deviceRepo "github.com/SENERGY-Platform/device-repository/lib/client"
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
			ctrl.Config().GetLogger().Error("unable to check device rights", "error", err, "ids", ids, "right", "r")
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.CheckDeviceOnlineStates(ids)
		if err != nil {
			ctrl.Config().GetLogger().Error("unable to check device online states", "error", err, "ids", ids)
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
			ctrl.Config().GetLogger().Error("unable to check device rights", "error", err, "ids", ids, "right", "r")
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
			ctrl.Config().GetLogger().Error("unable to decode body", "error", err)
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
			ctrl.Config().GetLogger().Error("unable to check hub rights", "error", err, "ids", ids, "right", "r")
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		result, err := ctrl.CheckGatewayOnlineStates(ids)
		if err != nil {
			ctrl.Config().GetLogger().Error("unable to check gateway online states", "error", err, "ids", ids)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}
