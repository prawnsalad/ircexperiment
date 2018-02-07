package common

import "encoding/gob"

// Server and worker both require RpcEvent gob encoding
func RegisterRpcGobTypes() {
	gob.Register(RpcEvent{})
	gob.Register(RpcEventConnState{})
	gob.Register(RpcEventConnData{})
}

// Events added to the event queue. Workers pick these up to act upon

type RpcEvent struct {
	Name  string
	Event interface{}
}

const RpcEventConnStateName = "conn.state"
const RpcEventConnTypeIn = 0
const RpcEventConnTypeOut = 1

const RpcEventConnStateClosed = 0
const RpcEventConnStateOpen = 1

type RpcEventConnState struct {
	ConnID int
	// ConnType either incoming or outgoing connection
	ConnType int
	State    int
	// Reason for the state change. Used in a closing state
	Reason string
	// RAddress "address:port"
	RAddress string
	Tls      bool
}

const RpcEventConnDataName = "conn.data"

type RpcEventConnData struct {
	ConnID   int
	ConnType int
	Data     []byte
}

type RpcDataKeyVal struct {
	Key string
	Val []byte
}

type RpcDataHash struct {
	Key   string
	Field string
	Val   []byte
}
