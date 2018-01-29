package master

import (
	"../common"
	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

type RpcDataInterface struct {
	DB *ledis.DB
}

func NewRpcDataInterface() *RpcDataInterface {
	c := lediscfg.NewConfigDefault()
	c.DBName = "memory"
	l, _ := ledis.Open(c)

	rpc := &RpcDataInterface{}
	rpc.DB, _ = l.Select(0)

	rpc.DB.HSet([]byte("server"), []byte("mask"), []byte("serv.serv"))
	return rpc
}
func (data *RpcDataInterface) Get(key string, resp *[]byte) error {
	println("data.Get()")
	val, _ := data.DB.Get([]byte(key))
	*resp = val
	return nil
}

func (data *RpcDataInterface) Set(e common.RpcDataKeyVal, resp *int) error {
	data.DB.Set([]byte(e.Key), e.Val)
	return nil
}

func (data *RpcDataInterface) HSet(e common.RpcDataHash, resp *int) error {
	data.DB.Set([]byte(e.Key), e.Val)
	data.DB.HSet([]byte(e.Key), []byte(e.Field), e.Val)
	return nil
}

func (data *RpcDataInterface) HGet(e common.RpcDataHash, resp *[]byte) error {
	val, _ := data.DB.HGet([]byte(e.Key), []byte(e.Field))
	*resp = val
	return nil
}

func (data *RpcDataInterface) HGetAll(e common.RpcDataHash, resp *map[string][]byte) error {
	valRaw, _ := data.DB.HGetAll([]byte(e.Key))
	val := make(map[string][]byte)
	for _, pair := range valRaw {
		val[string(pair.Field)] = pair.Value
	}

	*resp = val
	return nil
}

func (data *RpcDataInterface) SGet(e common.RpcDataKeyVal, resp *[][]byte) error {
	items, _ := data.DB.SMembers([]byte(e.Key))
	*resp = items
	return nil
}

func (data *RpcDataInterface) SAdd(e common.RpcDataKeyVal, resp *int) error {
	data.DB.SAdd([]byte(e.Key), e.Val)
	return nil
}

func (data *RpcDataInterface) SDel(e common.RpcDataKeyVal, resp *int) error {
	data.DB.SRem([]byte(e.Key), e.Val)
	return nil
}
