package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"motify_core_api/godep_libs/service/admin/wshub"
	"motify_core_api/godep_libs/service/dconfig"
	"motify_core_api/godep_libs/service/logger"
)

type DConfigHTTPHandler struct {
	mux *http.ServeMux
}

func NewDSettingsHTTPHandler(manager *dconfig.Manager) *DConfigHTTPHandler {
	handler := &DConfigHTTPHandler{
		mux: http.NewServeMux(),
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	wsHub := wshub.NewWSHub()

	manager.Subscribe(func(setting *dconfig.Setting) {
		data, err := json.Marshal(map[string]interface{}{
			"type":    "SETTING_CHANGE",
			"setting": setting,
		})
		if err == nil {
			wsHub.Broadcast <- data
		} else {
			logger.Error(nil, err.Error())
		}
	})

	handler.mux.HandleFunc("/edit", requireAuth(func(w http.ResponseWriter, r *http.Request, email string) {
		w.Header().Set("Content-Type", "application/json")
		if err := manager.EditSetting(context.Background(), email, r.FormValue("key"), r.FormValue("value")); err == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "OK",
			})
		} else {
			logger.Error(nil, err.Error())
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "ERR",
				"error":  err.Error(),
			})
		}
	}))

	handler.mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error(nil, err.Error())
			http.Error(w, "could not open websocket connection", http.StatusBadRequest)
			return
		}

		var data []byte
		if settings, err := manager.GetSettings(context.Background()); err == nil {
			data, err = json.Marshal(map[string]interface{}{
				"type":     "SETTINGS_LIST",
				"settings": settings,
			})
		}
		if err != nil {
			logger.Error(nil, err.Error())
		}

		wsHub.ProcessWSConnection(conn, data)
	})

	return handler
}

func (h *DConfigHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
