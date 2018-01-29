package master

import (
	"sync"

	"../common"
)

type RpcQueueInterface struct {
	queueLock sync.Mutex
	queue     []common.RpcEvent
}

func (s *RpcQueueInterface) addEvent(name string, event interface{}) {
	s.queueLock.Lock()
	s.queue = append(s.queue, common.RpcEvent{Name: name, Event: event})
	s.queueLock.Unlock()
}

func (s *RpcQueueInterface) Get(eventType string, resp *interface{}) error {
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	if len(s.queue) == 0 {
		*resp = common.RpcEvent{}
		return nil
	}

	var event common.RpcEvent
	event, s.queue = s.queue[0], s.queue[1:]
	*resp = event

	return nil
}
