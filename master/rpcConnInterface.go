package master

import (
	"io"
	"net"
	"sync"

	"../common"
)

type RpcConnInterface struct {
	queueLock  sync.Mutex
	queue      []interface{}
	MasterProc *MasterProcess
}

func (s *RpcConnInterface) Write(e common.RpcEventConnData, resp *int) error {
	//println("[master] Writing data to client", e.ClientID, e.Data)
	conn := s.MasterProc.GetConn(e.ConnID)
	if conn != nil {
		conn.Conn.Write(e.Data)
	}
	return nil
}

func (s *RpcConnInterface) Close(e common.RpcEventConnState, resp *int) error {
	conn := s.MasterProc.GetConn(e.ConnID)
	if conn != nil {
		conn.Conn.Close()
		s.MasterProc.RemoveConn(conn.Id)
		println("Closing client", conn.Id)
	}
	return nil
}

func (s *RpcConnInterface) Open(e common.RpcEventConnState, resp *error) error {
	tcpConn, err := net.Dial("tcp", e.RAddress)
	if err != nil {
		resp = &err
	}

	connRwc := io.ReadWriteCloser(tcpConn)
	conn := s.MasterProc.AddConn(connRwc, common.RpcEventConnTypeOut)
	go s.MasterProc.pipeConn(conn)
	return nil
}
