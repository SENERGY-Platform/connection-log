package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/connection-log/pkg/api/util"
	"github.com/SENERGY-Platform/connection-log/pkg/controller"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
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
		result, err := ctrl.GetResourcesLogEdge(ids, "gateway", duration)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(result)
	}
}
