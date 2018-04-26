package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"godep.lzd.co/service/admin/wshub"
	"godep.lzd.co/service/interfaces"
	"godep.lzd.co/service/logger"
)

const updateInterval = time.Second

type StatusHTTPHandler struct {
	*http.ServeMux
}

func NewStatusHTTPHandler(serviceID string, resources []interfaces.IResource) StatusHTTPHandler {
	handler := StatusHTTPHandler{http.NewServeMux()}
	handler.ownStatus()
	handler.resources(serviceID, resources)
	return handler
}

func (handler StatusHTTPHandler) ownStatus() {
	statusWSHub := wshub.NewWSHub()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	handler.HandleFunc("/ws-status", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error(nil, err.Error())
			http.Error(w, "could not open websocket connection", http.StatusBadRequest)
			return
		}

		statusWSHub.ProcessWSConnection(conn, nil)
	})
}

func (handler StatusHTTPHandler) resources(serviceID string, resources []interfaces.IResource) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	handler.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		names := []string{}
		for _, resource := range resources {
			names = append(names, resource.Caption())
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"caption":   serviceID,
			"resources": names,
		})
	})

	handler.HandleFunc("/resources.json", func(w http.ResponseWriter, r *http.Request) {
		names := []string{}
		for _, resource := range resources {
			names = append(names, resource.Caption())
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"caption":   serviceID,
			"resources": names,
		})
	})

	resourcesWSHubs := make(map[string]*wshub.WSHub, len(resources))
	resourceMap := make(map[string]interfaces.IResource)
	for _, resource := range resources {
		resourcesWSHubs[resource.Caption()] = wshub.NewWSHub()
		resourceMap[resource.Caption()] = resource
	}

	handler.HandleFunc("/ws-resource", func(w http.ResponseWriter, r *http.Request) {
		resourceName := r.FormValue("resource")
		if resourceName == "" {
			http.Error(w, "resource name required", http.StatusBadRequest)
			return
		}

		if wsHub, ok := resourcesWSHubs[resourceName]; ok {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logger.Warning(nil, err.Error())
				http.Error(w, "could not open websocket connection", http.StatusBadRequest)
				return
			}

			wsHub.ProcessWSConnection(conn, nil)
			return
		}

		logger.Warning(nil, "no info available for resource %q", resourceName)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	})

	go func() {
		for {
			for _, resource := range resources {
				if wsHub, ok := resourcesWSHubs[resource.Caption()]; ok {
					if data, err := json.Marshal(resource.Status()); err != nil {
						logger.Warning(nil, err.Error())
					} else {
						wsHub.Broadcast <- data
					}
				}
			}

			time.Sleep(updateInterval)
		}
	}()

	handler.HandleFunc("/resource.json", func(w http.ResponseWriter, r *http.Request) {
		resourceName := r.FormValue("name")
		if resourceName == "" {
			http.Error(w, "resource name required", http.StatusBadRequest)
			return
		}

		if resource, ok := resourceMap[resourceName]; ok {
			if data, err := json.Marshal(resource.Status()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}
		} else {
			http.Error(w, "resource wasn't found", http.StatusBadRequest)
			return
		}
	})
}
