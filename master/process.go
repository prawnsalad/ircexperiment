package master

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"

	"../common"
)

type Conn struct {
	Id   int
	Conn io.ReadWriteCloser
	Type int
}

type MasterProcess struct {
	ConnsMutex sync.Mutex
	Conns      map[int]*Conn
	EventQueue *RpcQueueInterface

	nextConnID int
}

func NewMasterProcess() *MasterProcess {
	proc := &MasterProcess{}
	proc.EventQueue = &RpcQueueInterface{}
	proc.Conns = make(map[int]*Conn)
	proc.nextConnID = 1

	return proc
}

func (master *MasterProcess) AddConn(connRwc io.ReadWriteCloser, connType int) *Conn {
	master.ConnsMutex.Lock()

	connID := master.nextConnID
	master.nextConnID++

	conn := &Conn{
		Id:   connID,
		Conn: connRwc,
		Type: connType,
	}

	master.Conns[connID] = conn
	master.ConnsMutex.Unlock()
	return conn

}

func (master *MasterProcess) RemoveConn(connID int) {
	master.ConnsMutex.Lock()
	delete(master.Conns, connID)
	master.ConnsMutex.Unlock()

}

func (master *MasterProcess) GetConn(connID int) *Conn {
	master.ConnsMutex.Lock()
	conn, _ := master.Conns[connID]
	master.ConnsMutex.Unlock()

	return conn

}

func (master *MasterProcess) ListenForClients() {
	srv, srvErr := net.Listen("tcp", ":7500")
	if srvErr != nil {
		log.Fatalln(srvErr.Error())
	}
	log.Println("Listening on " + srv.Addr().String())

	for {
		conn, err := srv.Accept()
		if err != nil {
			break
		}

		master.handleIncomingConn(io.ReadWriteCloser(conn), conn.RemoteAddr().String())
	}
}

func (master *MasterProcess) handleIncomingConn(connRwc io.ReadWriteCloser, rAddress string) {
	conn := master.AddConn(connRwc, common.RpcEventConnTypeIn)

	master.EventQueue.addEvent(common.RpcEventConnStateName, common.RpcEventConnState{
		ConnID:   conn.Id,
		ConnType: common.RpcEventConnTypeIn,
		State:    common.RpcEventConnStateOpen,
		RAddress: rAddress,
	})

	go master.pipeConn(conn)
}

// pipeConn - Take a Conn instance and pass any data it receives from it into the queue
func (master *MasterProcess) pipeConn(conn *Conn) {
	lineReader := bufio.NewReader(conn.Conn)
	for {
		line, err := lineReader.ReadSlice('\n')
		if err != nil {
			break
		}

		master.EventQueue.addEvent(common.RpcEventConnDataName, common.RpcEventConnData{
			ConnID:   conn.Id,
			ConnType: conn.Type,
			Data:     line,
		})
	}

	master.EventQueue.addEvent(common.RpcEventConnStateName, common.RpcEventConnState{
		ConnID: conn.Id,
		State:  common.RpcEventConnStateClosed,
	})

	master.RemoveConn(conn.Id)
}
