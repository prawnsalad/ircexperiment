package common

import "encoding/gob"

// Server and worker both require RpcEvent gob encoding
func RegisterRpcGobTypes() {
	gob.Register(RpcEvent{})
	gob.Register(RpcEventClientState{})
	gob.Register(RpcEventClientData{})
}

// Events added to the event queue. Workers pick these up to act upon

type RpcEvent struct {
	Name  string
	Event interface{}
}

var RpcEventClientStateName = "client.state"

type RpcEventClientState struct {
	ClientID int
	State    int
	Reason   string
}

var RpcEventClientDataName = "client.data"

type RpcEventClientData struct {
	ClientID int
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
