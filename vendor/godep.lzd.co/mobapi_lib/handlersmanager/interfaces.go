package handlersmanager

import "github.com/sergei-svistunov/gorpc"

type IHandlersManager interface {
	RegisterHandler(h gorpc.IHandler) error
	GetGoRPCHandlersManager() *gorpc.HandlersManager
}