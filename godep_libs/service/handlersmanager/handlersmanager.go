package handlersmanager

import (
	"reflect"

	"github.com/sergei-svistunov/gorpc"
)

type HandlersManager struct {
	*gorpc.HandlersManager
}

func New(handlersPath string) *HandlersManager {
	hm := &HandlersManager{
		HandlersManager: gorpc.NewHandlersManager(handlersPath, gorpc.HandlersManagerCallbacks{
			OnHandlerRegistration: func(path string, method reflect.Method) interface{} {
				return method
			},
		}),
	}

	return hm
}

func (hm *HandlersManager) GetGoRPCHandlersManager() *gorpc.HandlersManager {
	return hm.HandlersManager
}
