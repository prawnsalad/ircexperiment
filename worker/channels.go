package worker

import (
	"fmt"
	"strings"
	"sync"
)

type ChannelList struct {
	sync.Mutex
	Channels map[string]*Channel
}

func NewChannelList() *ChannelList {
	return &ChannelList{
		Channels: make(map[string]*Channel),
	}
}

func (chans *ChannelList) Get(name string) *Channel {
	chans.Lock()
	c, _ := chans.Channels[strings.ToLower(name)]
	chans.Unlock()
	return c
}

func (chans *ChannelList) Add(c *Channel) error {
	if c.Name == "" {
		return fmt.Errorf("no name set on the channel")
	}
	chans.Lock()
	chans.Channels[strings.ToLower(c.Name)] = c
	chans.Unlock()

	return nil
}

func (chans *ChannelList) GetOrAdd(name string) *Channel {
	chans.Lock()
	defer chans.Unlock()

	c, _ := chans.Channels[strings.ToLower(name)]
	if c != nil {
		return c
	}

	c = NewChannel()
	c.Name = name

	chans.Channels[strings.ToLower(c.Name)] = c
	return c
}

func (chans *ChannelList) Del(name string) {
	chans.Lock()
	delete(chans.Channels, strings.ToLower(name))
	chans.Unlock()
}

func (chans *ChannelList) Len() int {
	chans.Lock()
	length := len(chans.Channels)
	chans.Unlock()
	return length
}

type Channel struct {
	Name        string
	Modes       *ModeList
	Topic       string
	ClientsLock sync.Mutex
	Clients     []*Client
	Signals     chan ChannelEvent
}

func NewChannel() *Channel {
	channel := &Channel{
		Modes:   NewModeList(),
		Signals: make(chan ChannelEvent),
	}

	go func() {
		for event := range channel.Signals {
			switch event.Type {
			case "privmsg":
				line := fmt.Sprintf(
					":%s!%s@%s PRIVMSG %s :%s",
					event.Source.User.Nick,
					event.Source.User.Username,
					event.Source.User.Hostname,
					channel.Name,
					event.Message,
				)

				channel.ClientsLock.Lock()
				for _, client := range channel.Clients {
					if client != event.Source {
						client.Write(line)
					}
				}
				channel.ClientsLock.Unlock()

			case "topic":
				channel.Topic = event.Message
				line := fmt.Sprintf("TOPIC %s :%s", channel.Name, channel.Topic)
				channel.ClientsLock.Lock()
				for _, client := range channel.Clients {
					client.Write(line)
				}
				channel.ClientsLock.Unlock()
			}
		}
	}()
	return channel
}

type ChannelEvent struct {
	Type    string
	Message string
	Source  *Client
}
