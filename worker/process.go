package worker

import (
	"fmt"
	"net/rpc"

	"../common"
	"github.com/kiwiirc/webircgateway/pkg/irc"
)

type WorkerProces struct {
	Commands  map[string]Command
	rpcClient *rpc.Client
	Data      *DataRpcWrapper
}

func NewWorkerProcess() *WorkerProces {
	proc := &WorkerProces{}
	proc.Commands = make(map[string]Command)

	return proc
}

func (worker *WorkerProces) SetRpcClient(client *rpc.Client) {
	worker.rpcClient = client
	worker.Data = &DataRpcWrapper{rpcClient: client}
}

func (worker *WorkerProces) LoadCommands() {
	worker.Commands = loadCommands(worker)
}

func (worker *WorkerProces) RpcCall(serviceMethod string, args interface{}, reply interface{}) error {
	err := worker.rpcClient.Call(serviceMethod, args, reply)
	if err != nil {
		println("[worker] RpcCall()", serviceMethod, err.Error())
	}
	return err
}

func (worker *WorkerProces) HandleRpcEvent(rpcEvent common.RpcEvent) {
	println("[worker] HandleRpcEvent()", rpcEvent.Name)
	if rpcEvent.Name == common.RpcEventClientStateName {
		event := rpcEvent.Event.(common.RpcEventClientState)
		// event.State 0=closed 1=connected
		println("[worker client state changed]", event.ClientID, event.State)
	}

	if rpcEvent.Name == common.RpcEventClientDataName {
		event := rpcEvent.Event.(common.RpcEventClientData)
		// log.Printf("[worker] client:%d %s", event.ClientID, string(event.Data))
		msg, _ := irc.ParseLine(string(event.Data))
		if msg == nil {
			return
		}

		runCommand(worker, event.ClientID, msg)
	}
}

func (worker *WorkerProces) WriteClient(clientID int, format string, args ...interface{}) {
	format = format + "\n"
	line := fmt.Sprintf(format, args...)
	println("WriteClient()", line)

	dataCall := common.RpcEventClientData{
		ClientID: clientID,
		Data:     []byte(line),
	}

	worker.RpcCall("clients.Write", dataCall, nil)
}
func (worker *WorkerProces) CloseClient(clientID int) {
	dataCall := common.RpcEventClientState{
		ClientID: clientID,
	}

	worker.rpcClient.Call("client.Close", dataCall, nil)
}
