package master

import (
	"bufio"
	"io"
	"log"
	"net"

	"../common"
)

type MasterProcess struct {
	Clients map[int]io.ReadWriteCloser
	//Channels   *ChannelList
	EventQueue *RpcQueueInterface

	nextClientID int
}

func NewMasterProcess() *MasterProcess {
	proc := &MasterProcess{}
	proc.EventQueue = &RpcQueueInterface{}
	proc.Clients = make(map[int]io.ReadWriteCloser)
	//proc.Channels = NewChannelList()
	proc.nextClientID = 1

	return proc
}

func (master *MasterProcess) ListenForClients() {
	srv, _ := net.Listen("tcp", ":7500")
	log.Println("Listening on " + srv.Addr().String())
	for {
		conn, err := srv.Accept()
		if err != nil {
			break
		}

		clientID := master.nextClientID
		master.nextClientID++

		master.handleConn(conn, clientID)
	}
}

func (master *MasterProcess) handleConn(conn net.Conn, clientID int) {
	master.Clients[clientID] = io.ReadWriteCloser(conn)

	master.EventQueue.addEvent(common.RpcEventClientStateName, common.RpcEventClientState{
		ClientID: clientID,
		State:    1,
	})

	go func() {
		lineReader := bufio.NewReader(conn)
		for {
			line, err := lineReader.ReadSlice('\n')
			if err != nil {
				break
			}

			master.EventQueue.addEvent(common.RpcEventClientDataName, common.RpcEventClientData{
				ClientID: clientID,
				Data:     line,
			})
		}

		master.EventQueue.addEvent(common.RpcEventClientStateName, common.RpcEventClientState{
			ClientID: clientID,
			State:    0,
		})

		delete(master.Clients, clientID)
	}()
}
