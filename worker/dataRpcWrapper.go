package worker

import (
	"fmt"
	"net/rpc"
	"strconv"
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

type DataWrapper struct {
	rpcClient *rpc.Client
}

// Hash functions
func (data *DataWrapper) HashGet(key string, field string) []byte {
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
func (data *DataWrapper) HashGetAll(key string) map[string][]byte {
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
func (data *DataWrapper) HashSet(key string, field string, val []byte) {
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

func MakeGetter(data *DataRpcWrapper, key string, field string) func() []byte {
	return func() []byte {
		return data.HashGet(key, field)
	}
}
func MakeSetter(data *DataRpcWrapper, key string, field string) func([]byte) {
	return func(val []byte) {
		data.HashSet(key, field, val)
	}
}

func MakeGetterInt(data *DataRpcWrapper, key string, field string) func() int {
	getter := MakeGetter(data, key, field)
	return func() int {
		bytes := getter()
		return byteAsInt(bytes)
	}
}
func MakeSetterInt(data *DataRpcWrapper, key string, field string) func(int) {
	setter := MakeSetter(data, key, field)
	return func(val int) {
		bytes := intAsByte(val)
		setter(bytes)
	}
}

func MakeGetterBool(data *DataRpcWrapper, key string, field string) func() bool {
	getter := MakeGetter(data, key, field)
	return func() bool {
		bytes := getter()
		return byteAsBool(bytes)
	}
}
func MakeSetterBool(data *DataRpcWrapper, key string, field string) func(bool) {
	setter := MakeSetter(data, key, field)
	return func(val bool) {
		bytes := boolAsByte(val)
		setter(bytes)
	}
}

func MakeGetterString(data *DataRpcWrapper, key string, field string) func() string {
	getter := MakeGetter(data, key, field)
	return func() string {
		bytes := getter()
		return string(bytes)
	}
}
func MakeSetterString(data *DataRpcWrapper, key string, field string) func(string) {
	setter := MakeSetter(data, key, field)
	return func(val string) {
		bytes := []byte(val)
		setter(bytes)
	}
}

type DataWrapperClient struct {
	data           *DataRpcWrapper
	ClientID       int
	Caps           DataWrapperHash
	Nick           func() string
	SetNick        func(string)
	RemoteAddr     func() string
	SetRemoteAddr  func(string)
	UserID         func() int
	SetUserID      func(int)
	ActiveNetID    func() int
	SetActiveNetID func(int)
}

func NewDataWrapperClient(data *DataRpcWrapper, clientID int) *DataWrapperClient {
	d := &DataWrapperClient{
		data:     data,
		ClientID: clientID,
	}

	key := fmt.Sprintf("client:%d", clientID)
	d.Nick = MakeGetterString(d.data, key, "nick")
	d.SetNick = MakeSetterString(d.data, key, "nick")
	d.RemoteAddr = MakeGetterString(d.data, key, "remote_addr")
	d.SetRemoteAddr = MakeSetterString(d.data, key, "remote_addr")
	d.UserID = MakeGetterInt(d.data, key, "user_id")
	d.SetUserID = MakeSetterInt(d.data, key, "user_id")
	d.ActiveNetID = MakeGetterInt(d.data, key, "active_net_id")
	d.SetActiveNetID = MakeSetterInt(d.data, key, "active_net_id")

	d.Caps = DataWrapperHash{
		data:   data,
		prefix: key + ":caps",
	}

	return d
}

func (d *DataWrapperClient) NetworkData() *DataWrapperNetwork {
	activeNet := d.ActiveNetID()
	if activeNet == 0 {
		return nil
	}

	return NewDataWrapperNetwork(d.data, d.UserID(), activeNet)
}

type DataWrapperHash struct {
	data   *DataRpcWrapper
	prefix string
}

func (d *DataWrapperHash) Get(field string) []byte {
	return d.data.HashGet(d.prefix, field)
}
func (d *DataWrapperHash) Set(field string, val []byte) {
	d.data.HashSet(d.prefix, field, val)
}

type DataWrapperNetwork struct {
	data          *DataRpcWrapper
	ConnID        func() int
	SetConnID     func(int)
	Modes         DataWrapperHash
	Connected     func() bool
	SetConnected  func(bool)
	Registered    func() bool
	SetRegistered func(bool)
	Nick          func() string
	SetNick       func(string)
	NetName       func() string
	SetNetName    func(string)
}

func (d *DataWrapperNetwork) DataWrapperClients() []*DataWrapperClient {
	clientConnIDs := d.data.SetGet(fmt.Sprintf("outgoing:%d:clients", d.ConnID()))
	var clients []*DataWrapperClient
	for _, x := range clientConnIDs {
		xx := byteAsInt(x)
		client := NewDataWrapperClient(d.data, xx)
		clients = append(clients, client)
	}

	return clients
}

func NewDataWrapperNetwork(data *DataRpcWrapper, userID int, netID int) *DataWrapperNetwork {
	d := &DataWrapperNetwork{
		data: data,
	}

	key := fmt.Sprintf("client:%d:%d", userID, netID)
	d.ConnID = MakeGetterInt(d.data, key, "conn_id")
	d.SetConnID = MakeSetterInt(d.data, key, "conn_id")
	d.Connected = MakeGetterBool(d.data, key, "connected")
	d.SetConnected = MakeSetterBool(d.data, key, "connected")
	d.Registered = MakeGetterBool(d.data, key, "registered")
	d.SetRegistered = MakeSetterBool(d.data, key, "registered")
	d.Nick = MakeGetterString(d.data, key, "nick")
	d.SetNick = MakeSetterString(d.data, key, "nick")
	d.NetName = MakeGetterString(d.data, key, "net_name")
	d.SetNetName = MakeSetterString(d.data, key, "net_name")

	d.Modes = DataWrapperHash{
		data:   data,
		prefix: key + ":modes",
	}

	return d
}

func NetworkFromConnID(data *DataRpcWrapper, connID int) *DataWrapperNetwork {
	connOwner := data.HashGetAll(fmt.Sprintf("outgoing:%d", connID))
	if len(connOwner) == 0 {
		return nil
	}

	f := byteMapAsStrings(connOwner)
	userID, _ := strconv.Atoi(f["user_id"])
	netID, _ := strconv.Atoi(f["net_id"])

	if userID == 0 || netID == 0 {
		return nil
	}

	net := NewDataWrapperNetwork(data, userID, netID)
	return net
}
