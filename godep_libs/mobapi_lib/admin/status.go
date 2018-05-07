package admin

import (
	"encoding/json"
	"motify_core_api/godep_libs/mobapi_lib/resources"
	"net/http"
)

type StatusHTTPHandler struct {
	*http.ServeMux
}

func NewStatusHTTPHandler(serviceID string, resourceList []resources.IResource) StatusHTTPHandler {
	handler := StatusHTTPHandler{http.NewServeMux()}
	handler.resources(serviceID, resourceList)
	return handler
}

func (handler StatusHTTPHandler) resources(serviceID string, resourceList []resources.IResource) {
	handler.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		names := []string{}
		for _, resource := range resourceList {
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
		for _, resource := range resourceList {
			names = append(names, resource.Caption())
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"caption":   serviceID,
			"resources": names,
		})
	})

	resourceMap := make(map[string]resources.IResource)
	for _, resource := range resourceList {
		resourceMap[resource.Caption()] = resource
	}

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
