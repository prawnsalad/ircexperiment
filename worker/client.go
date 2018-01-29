package worker

import (
	"fmt"
	"log"
	"net"
)

type RChannel struct {
	Worker        *WorkerProces
	Loaded        bool
	Modes         *ModeList
	Name          string
	ClientsLoaded bool
	Clients       []int
}

func NewRChannel(worker *WorkerProces) *RChannel {
	return &RChannel{Worker: worker}
}

func (c *RChannel) Load(name string) {
	c.Name = name

	modeMap := c.Worker.Data.ChannelModes(name)
	c.Modes = NewModeListFromMap(modeMap)

	c.Name = string(c.Worker.Data.ChannelGet(name, "name"))
	c.Loaded = true
}

func (c *RChannel) GetClients() []int {
	if !c.ClientsLoaded {
		clientsInChannel := c.Worker.Data.ChannelSGet(c.Name, "clients")
		for _, id := range clientsInChannel {
			c.Clients = append(c.Clients, byteAsInt(id))
		}

		c.ClientsLoaded = true
	}

	return c.Clients
}

func (c *RChannel) IsClientOn(clientID int) bool {
	clientsInChannel := c.GetClients()
	isInChannel := false
	for _, id := range clientsInChannel {
		if id == clientID {
			isInChannel = true
			break
		}
	}
	return isInChannel
}

type RClient struct {
	Worker     *WorkerProces
	Loaded     bool
	Id         int
	Registered bool
	Nick       string
	Username   string
	Hostname   string
	RealName   string
	Modes      *ModeList
}

func NewRClient(worker *WorkerProces) *RClient {
	return &RClient{Worker: worker}
}

func (c *RClient) Load(clientID int) {
	//println("Loading client", clientID)
	c.Id = clientID
	c.Registered = byteAsBool(c.Worker.Data.ClientGet(clientID, "registered"))
	c.Username = string(c.Worker.Data.ClientGet(clientID, "username"))
	c.Nick = string(c.Worker.Data.ClientGet(clientID, "nick"))
	c.RealName = string(c.Worker.Data.ClientGet(clientID, "realname"))
	c.Hostname = string(c.Worker.Data.ClientGet(clientID, "hostname"))

	modeMap := c.Worker.Data.ClientModes(clientID)
	c.Modes = NewModeListFromMap(modeMap)

	c.Loaded = true
}

func (c *RClient) Write(format string, args ...interface{}) {
	c.Worker.WriteClient(c.Id, format, args...)
}

func (c *RClient) WriteWithPrefix(format string, args ...interface{}) {
	prefix := fmt.Sprintf(":%s!%s@%s", c.Nick, c.Username, c.Hostname)
	println(c.Loaded, prefix)
	format = prefix + " " + format
	c.Worker.WriteClient(c.Id, format, args...)
}

func (c *RClient) WriteWithServerPrefix(format string, args ...interface{}) {
	prefix := string(c.Worker.Data.HashGet("server", "mask"))
	format = ":" + prefix + " " + format
	c.Worker.WriteClient(c.Id, format, args...)
}

func (c *RClient) CanRegister() bool {
	log.Printf("c.Username = '%s' c.Nick = '%s' c.RealName = '%s'", c.Username, c.Nick, c.RealName)
	if c.Username != "" && c.Nick != "" && c.RealName != "" {
		return true
	}

	return false
}

type Client struct {
	Id         int
	Conn       net.Conn
	Registered bool
	User       struct {
		Nick     string
		Username string
		Hostname string
		RealName string
		Meta     map[string]string
		Modes    *ModeList
	}
	Channels *ChannelList
}

var nextClientId int

func NewClient(conn net.Conn) *Client {
	nextClientId++
	c := &Client{
		Id:       nextClientId,
		Conn:     conn,
		Channels: NewChannelList(),
	}

	c.User.Meta = make(map[string]string)
	c.User.Modes = NewModeList()
	c.User.Hostname = conn.RemoteAddr().String()

	return c
}

func (c *Client) Write(format string, args ...interface{}) {
	format = format + "\n"
	line := fmt.Sprintf(format, args...)
	c.Conn.Write([]byte(line))
}

func (c *Client) WriteWithPrefix(format string, args ...interface{}) {
	prefix := fmt.Sprintf(":%s!%s@%s", c.User.Nick, c.User.Username, c.User.Hostname)
	format = prefix + " " + format + "\n"
	line := fmt.Sprintf(format, args...)
	c.Conn.Write([]byte(line))
}
func (c *Client) WriteWithServerPrefix(format string, args ...interface{}) {
	prefix := fmt.Sprintf(":%s!%s@%s", c.User.Nick, c.User.Username, c.User.Hostname)
	format = prefix + " " + format + "\n"
	line := fmt.Sprintf(format, args...)
	c.Conn.Write([]byte(line))
}

func canClientRegister(c *Client) bool {
	return c.User.Username != "" && c.User.Nick != "" && c.User.RealName != ""
}

func registerClient(c *RClient) {
	println("registerClient()")
	c.Registered = true
	// TODO: Add client to global client mapping or something

	c.WriteWithServerPrefix("001 %s :Welcome to the network", c.Nick)
	c.WriteWithServerPrefix("002 %s :Your host is spamnet, running version 1", c.Nick)

	isup := NewISupportBuilder()
	isup.Set("AWAYLEN", "500")
	isup.Set("CASEMAPPING", "ascii")
	isup.Set("CHANMODES", "beI,,lk,imntEs")
	isup.Set("CHANNELLEN", "64")
	isup.Set("CHANTYPES", "#")
	isup.Set("ELIST", "U")
	isup.Set("EXCEPTS", "")
	isup.Set("INVEX", "")
	isup.Set("KICKLEN", "1000")
	isup.Set("MAXLIST", "beI:60")
	isup.Set("MAXTARGETS", "4")
	isup.Set("MODES", "")
	isup.Set("MONITOR", "100")
	c.WriteWithServerPrefix("005 %s %s :are supported by this server", c.Nick, isup.AsString())

	c.WriteWithServerPrefix("422 :There is no MOTD on this server")
}

func addClientToChannel(client *Client, channel *Channel) bool {
	println("addClientToChannel()", client.User.Nick, channel.Name)
	channel.ClientsLock.Lock()
	defer channel.ClientsLock.Unlock()

	// Make sure this client doesn't exist in the channel already
	for _, compareClient := range channel.Clients {
		if client == compareClient {
			println("Client already in channel")
			return false
		}
	}

	channel.Clients = append(channel.Clients, client)
	client.Channels.Add(channel)
	return true
}

func delClientFromChannel(client *Client, channel *Channel) bool {
	channel.ClientsLock.Lock()
	defer channel.ClientsLock.Unlock()

	clientIdx := -1
	for idx, compareClient := range channel.Clients {
		if client == compareClient {
			clientIdx = idx
			break
		}
	}

	if clientIdx == -1 {
		return false
	}

	// Swap the client with the last one in the array, then remove the last item. Fast.
	i := clientIdx
	channel.Clients[len(channel.Clients)-1], channel.Clients[i] = channel.Clients[i], channel.Clients[len(channel.Clients)-1]
	channel.Clients = channel.Clients[:len(channel.Clients)-1]

	client.Channels.Del(channel.Name)
	return true
}
