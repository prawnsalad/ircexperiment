package worker

import (
	"fmt"
	"net/rpc"
	"strings"

	"../common"
)

var dataRpcWrapperDebug bool

func debugLog(args ...interface{}) {
	//log.Println(args...)
}

type DataRpcWrapper struct {
	rpcClient *rpc.Client
}

// Hash functions
func (data *DataRpcWrapper) HashGet(key string, field string) []byte {
	debugLog("HashGet()", key, field)
	dataCall := common.RpcDataHash{
		Key:   strings.ToLower(key),
		Field: strings.ToLower(field),
	}
	var val []byte
	err := data.rpcClient.Call("data.HGet", dataCall, &val)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}

	return val
}

// Hash functions
func (data *DataRpcWrapper) HashGetAll(key string) map[string][]byte {
	debugLog("HashGetAll()", key)
	dataCall := common.RpcDataHash{
		Key: strings.ToLower(key),
	}
	var val map[string][]byte
	err := data.rpcClient.Call("data.HGetAll", dataCall, &val)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}

	return val
}

// Hash functions
func (data *DataRpcWrapper) HashSet(key string, field string, val []byte) {
	debugLog("HashSet()", key, field)
	dataCall := common.RpcDataHash{
		Key:   strings.ToLower(key),
		Field: strings.ToLower(field),
		Val:   val,
	}

	err := data.rpcClient.Call("data.HSet", dataCall, nil)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}
}

// Basic k/v
func (data *DataRpcWrapper) Get(key string) []byte {
	debugLog("Get()", key)
	dataCall := common.RpcDataHash{
		Key: strings.ToLower(key),
	}
	var val []byte
	err := data.rpcClient.Call("data.Get", dataCall, &val)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}

	return val
}

// Basic k/v
func (data *DataRpcWrapper) Set(key string, val []byte) {
	debugLog("Set()", key)
	dataCall := common.RpcDataHash{
		Key: strings.ToLower(key),
		Val: val,
	}

	err := data.rpcClient.Call("data.Set", dataCall, nil)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}
}

// Set of data (unordered lists)
func (data *DataRpcWrapper) SetGet(key string) [][]byte {
	debugLog("SetGet()", key)
	dataCall := common.RpcDataHash{
		Key: strings.ToLower(key),
	}

	var items [][]byte
	err := data.rpcClient.Call("data.SGet", dataCall, &items)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}

	return items
}

// Set of data (unordered lists)
func (data *DataRpcWrapper) SetAdd(key string, val []byte) {
	debugLog("SetAdd()", key)
	dataCall := common.RpcDataHash{
		Key: strings.ToLower(key),
		Val: val,
	}

	err := data.rpcClient.Call("data.SAdd", dataCall, nil)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}
}

// Set of data (unordered lists)
func (data *DataRpcWrapper) SetDel(key string, val []byte) {
	debugLog("SetDel()", key)
	dataCall := common.RpcDataHash{
		Key: strings.ToLower(key),
		Val: val,
	}

	err := data.rpcClient.Call("data.SDel", dataCall, nil)
	if err != nil {
		debugLog("RPC err:", err.Error())
	}
}

/**
 * Handly helpers
 */

func (data *DataRpcWrapper) ClientGet(clientID int, field string) []byte {
	key := fmt.Sprintf("client:%d", clientID)
	return data.HashGet(key, field)
}

func (data *DataRpcWrapper) ClientSet(clientID int, field string, val []byte) {
	key := fmt.Sprintf("client:%d", clientID)
	data.HashSet(key, field, val)
}

func (data *DataRpcWrapper) ClientSGet(clientID int, key string) [][]byte {
	finalKey := fmt.Sprintf("client:%d:%s", clientID, key)
	return data.SetGet(finalKey)
}

func (data *DataRpcWrapper) ClientSAdd(clientID int, key string, val []byte) {
	finalKey := fmt.Sprintf("client:%d:%s", clientID, key)
	data.SetAdd(finalKey, val)
}

func (data *DataRpcWrapper) ClientSDel(clientID int, key string, val []byte) {
	finalKey := fmt.Sprintf("client:%d:%s", clientID, key)
	data.SetDel(finalKey, val)
}

func (data *DataRpcWrapper) ClientModes(clientID int) map[string]string {
	key := fmt.Sprintf("clientmodes:%d", clientID)
	ret := data.HashGetAll(key)
	modes := make(map[string]string)
	for k, v := range ret {
		modes[k] = string(v)
	}
	return modes
}

func (data *DataRpcWrapper) ClientModeGet(clientID int, mode string) []byte {
	key := fmt.Sprintf("clientmodes:%d", clientID)
	return data.HashGet(key, mode)
}

func (data *DataRpcWrapper) ClientModeSet(clientID int, mode string, val []byte) {
	key := fmt.Sprintf("clientmodes:%d", clientID)
	data.HashSet(key, mode, val)
}

func (data *DataRpcWrapper) ChannelGet(chanName string, field string) []byte {
	key := fmt.Sprintf("channel:%s", chanName)
	return data.HashGet(key, field)
}

func (data *DataRpcWrapper) ChannelSet(chanName string, field string, val []byte) {
	key := fmt.Sprintf("channel:%s", chanName)
	data.HashSet(key, field, val)
}

func (data *DataRpcWrapper) ChannelModes(chanName string) map[string]string {
	key := fmt.Sprintf("channelmodes:%s", chanName)
	ret := data.HashGetAll(key)
	modes := make(map[string]string)
	for k, v := range ret {
		modes[k] = string(v)
	}
	return modes
}

func (data *DataRpcWrapper) ChannelModeGet(chanName string, mode string) []byte {
	key := fmt.Sprintf("channelmodes:%s", chanName)
	return data.HashGet(key, mode)
}

func (data *DataRpcWrapper) ChannelModeSet(chanName string, mode string, val []byte) {
	key := fmt.Sprintf("channelmodes:%s", chanName)
	data.HashSet(key, mode, val)
}

func (data *DataRpcWrapper) ChannelSGet(chanName string, key string) [][]byte {
	finalKey := fmt.Sprintf("channel:%s:%s", chanName, key)
	return data.SetGet(finalKey)
}

func (data *DataRpcWrapper) ChannelSAdd(chanName string, key string, val []byte) {
	finalKey := fmt.Sprintf("channel:%s:%s", chanName, key)
	data.SetAdd(finalKey, val)
}

func (data *DataRpcWrapper) ChannelSDel(chanName string, key string, val []byte) {
	finalKey := fmt.Sprintf("channel:%s:%s", chanName, key)
	data.SetDel(finalKey, val)
}
