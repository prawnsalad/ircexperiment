package worker

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"../common"
	"github.com/kiwiirc/webircgateway/pkg/irc"
)

type WorkerProces struct {
	ClientCommands map[string]Command
	ServerCommands map[string]ServerCommand
	rpcClient      *rpc.Client
	Data           *DataRpcWrapper
}

func NewWorkerProcess() *WorkerProces {
	proc := &WorkerProces{}
	proc.ClientCommands = make(map[string]Command)
	proc.ServerCommands = make(map[string]ServerCommand)

	return proc
}

func (worker *WorkerProces) SetRpcClient(client *rpc.Client) {
	worker.rpcClient = client
	worker.Data = &DataRpcWrapper{rpcClient: client}
}

func (worker *WorkerProces) LoadCommands() {
	worker.ClientCommands = loadClientCommands(worker)
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
	if rpcEvent.Name == common.RpcEventConnStateName {
		event := rpcEvent.Event.(common.RpcEventConnState)
		// event.State 0=closed 1=connected
		log.Printf("client state changed %d (%d %s)", event.State, event.ConnID, event.RAddress)
		if event.State == common.RpcEventConnStateOpen {
			host, _, _ := net.SplitHostPort(event.RAddress)
			worker.Data.ClientSet(event.ConnID, DbClientKeyHostname, []byte(host))
		}
	}

	if rpcEvent.Name == common.RpcEventConnDataName {
		event := rpcEvent.Event.(common.RpcEventConnData)
		log.Printf("[worker] data in client:%d %s", event.ConnID, string(event.Data))
		msg, _ := irc.ParseLine(string(event.Data))
		if msg == nil {
			return
		}

		if event.ConnType == common.RpcEventConnTypeIn {
			runClientCommand(worker, event.ConnID, msg)
		} else if event.ConnType == common.RpcEventConnTypeIn {
			//runInCommand(worker, event.ConnID, msg)
		}
	}
}

func (worker *WorkerProces) WriteClient(clientID int, format string, args ...interface{}) {
	format = format + "\n"
	line := fmt.Sprintf(format, args...)
	println("WriteClient()", line)

	dataCall := common.RpcEventConnData{
		ConnID: clientID,
		Data:   []byte(line),
	}

	worker.RpcCall("conns.Write", dataCall, nil)
}
func (worker *WorkerProces) CloseClient(clientID int) {
	dataCall := common.RpcEventConnState{
		ConnID: clientID,
	}

	worker.rpcClient.Call("conns.Close", dataCall, nil)
}
