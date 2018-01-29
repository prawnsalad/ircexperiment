package master

import (
	"sync"

	"../common"
)

type RpcClientInterface struct {
	queueLock  sync.Mutex
	queue      []interface{}
	MasterProc *MasterProcess
}

func (s *RpcClientInterface) Write(e common.RpcEventClientData, resp *int) error {
	//println("[master] Writing data to client", e.ClientID, e.Data)
	client, exists := s.MasterProc.Clients[e.ClientID]
	if exists {
		client.Write(e.Data)
	}
	return nil
}

func (s *RpcClientInterface) Close(e common.RpcEventClientState, resp *int) error {
	// client := getClientFromSomewhere(e.ClientID)
	// client.Close()
	println("Closing client", e.ClientID)
	return nil
}
