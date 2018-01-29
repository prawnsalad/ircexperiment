package worker

import (
	"encoding/binary"
	"sync"

	"github.com/kiwiirc/webircgateway/pkg/irc"
)

const (
	DbClientKeyRegistered = "registered"
	DbClientKeyNick       = "nick"
	DbClientKeyUsername   = "username"
	DbClientKeyRealname   = "realname"
	DbClientKeyHostname   = "hostname"
)

type ModeList struct {
	sync.Mutex
	Modes map[string]string
}

func NewModeList() *ModeList {
	return &ModeList{
		Modes: make(map[string]string),
	}
}

func NewModeListFromMap(modes map[string]string) *ModeList {
	return &ModeList{
		Modes: modes,
	}
}

func (m *ModeList) Has(mode string) bool {
	m.Lock()
	_, exists := m.Modes[mode]
	m.Unlock()
	return exists
}

func (m *ModeList) Get(mode string) string {
	m.Lock()
	val, _ := m.Modes[mode]
	m.Unlock()
	return val
}

func (m *ModeList) Set(mode string, val string) {
	m.Lock()
	m.Modes[mode] = val
	m.Unlock()
}

func (m *ModeList) Del(mode string) {
	m.Lock()
	delete(m.Modes, mode)
	m.Unlock()
}

func messageParam(msg *irc.Message, idx int) string {
	if idx > len(msg.Params)-1 {
		return ""
	}

	return msg.Params[idx]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

/*
func val(slice []string, item string) string {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
*/
func byteAsInt(inp []byte) int {
	if len(inp) == 0 {
		return 0
	}
	return int(binary.BigEndian.Uint32(inp))
}

func intAsByte(inp int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(inp))
	return b
}

func byteAsBool(inp []byte) bool {
	if len(inp) == 0 {
		return false
	}

	intVal := byteAsInt(inp)
	if intVal == 0 {
		return false
	}

	return true
}

func boolAsByte(inp bool) []byte {
	if inp {
		return intAsByte(1)
	} else {
		return intAsByte(0)
	}
}

func boolAsInt(val bool) []byte {
	if val {
		return intAsByte(1)
	} else {
		return intAsByte(0)
	}
}

func byteSliceAsStrings(inp [][]byte) []string {
	s := []string{}
	for _, b := range inp {
		s = append(s, string(b))
	}
	return s
}
