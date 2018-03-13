package master

import (
	"log"

	"../common"
	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

type RpcDataInterface struct {
	DB      *ledis.DB
	Persist *ledis.DB
}

func NewRpcDataInterface() *RpcDataInterface {
	rpc := &RpcDataInterface{}

	c := lediscfg.NewConfigDefault()
	c.DBName = "memory"
	l, _ := ledis.Open(c)
	rpc.DB, _ = l.Select(0)

	c = lediscfg.NewConfigDefault()
	c.DataDir = "./db/"
	l, openErr := ledis.Open(c)
	if openErr != nil {
		log.Fatal("Error opening database: " + openErr.Error())
	}
	rpc.Persist, _ = l.Select(0)

	// Set the BNC server name
	rpc.DB.HSet([]byte("server"), []byte("mask"), []byte("serv.serv"))

	return rpc
}

func (data *RpcDataInterface) SelectStore(storeType int) *ledis.DB {
	if storeType == common.DataStoreTemporary {
		return data.DB
	}

	if storeType == common.DataStorePersistent {
		return data.Persist
	}

	// For ease of use, return the temporary database by default
	return data.DB
}

func (data *RpcDataInterface) Get(e common.RpcData, resp *[]byte) error {
	println("data.Get()")
	store := data.SelectStore(e.Store)
	val, _ := store.Get([]byte(e.Key))
	*resp = val
	return nil
}

func (data *RpcDataInterface) Set(e common.RpcData, resp *int) error {
	store := data.SelectStore(e.Store)
	store.Set([]byte(e.Key), e.Val)
	return nil
}

func (data *RpcDataInterface) HSet(e common.RpcData, resp *int) error {
	store := data.SelectStore(e.Store)
	store.Set([]byte(e.Key), e.Val)
	store.HSet([]byte(e.Key), []byte(e.Field), e.Val)
	return nil
}

func (data *RpcDataInterface) HGet(e common.RpcData, resp *[]byte) error {
	store := data.SelectStore(e.Store)
	val, _ := store.HGet([]byte(e.Key), []byte(e.Field))
	*resp = val
	return nil
}

func (data *RpcDataInterface) HGetAll(e common.RpcData, resp *map[string][]byte) error {
	store := data.SelectStore(e.Store)
	valRaw, _ := store.HGetAll([]byte(e.Key))
	val := make(map[string][]byte)
	for _, pair := range valRaw {
		val[string(pair.Field)] = pair.Value
	}

	*resp = val
	return nil
}

func (data *RpcDataInterface) SGet(e common.RpcData, resp *[][]byte) error {
	store := data.SelectStore(e.Store)
	items, _ := store.SMembers([]byte(e.Key))
	*resp = items
	return nil
}

func (data *RpcDataInterface) SAdd(e common.RpcData, resp *int) error {
	store := data.SelectStore(e.Store)
	store.SAdd([]byte(e.Key), e.Val)
	return nil
}

func (data *RpcDataInterface) SDel(e common.RpcData, resp *int) error {
	store := data.SelectStore(e.Store)
	store.SRem([]byte(e.Key), e.Val)
	return nil
}
