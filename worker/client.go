package worker

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

// Add extra client functions around the client data wrapper
type Client struct {
	*DataWrapperClient
	Worker *WorkerProces
}

func NewClient(worker *WorkerProces, clientID int) *Client {
	c := NewDataWrapperClient(worker.Data, clientID)
	return &Client{
		Worker:            worker,
		DataWrapperClient: c,
	}
}

func NewClientFromDataWrapper(worker *WorkerProces, c *DataWrapperClient) *Client {
	return &Client{
		Worker:            worker,
		DataWrapperClient: c,
	}
}

func (c *Client) Write(format string, args ...interface{}) {
	c.Worker.WriteClient(c.ClientID, format, args...)
}

func (c *Client) WriteStatus(format string, args ...interface{}) {
	format = ":*status!this-is@your.bnc NOTICE * :" + format
	c.Worker.WriteClient(c.ClientID, format, args...)
}

func (c *Client) WriteWithServerPrefix(format string, args ...interface{}) {
	format = ":*status!this-is@your.bnc " + format
	c.Worker.WriteClient(c.ClientID, format, args...)
}

func (d *Client) Auth(username string, password string) bool {
	return AuthAccount(d, username, password)
}

func (d *Client) HasAuthed() bool {
	if d.UserID() == 0 {
		return false
	}

	return true
}

func (d *Client) GetNetworkInfo(networkName string) (netInfo NetworkInfo, exists bool) {
	netNameLower := strings.ToLower(networkName)
	userID := d.UserID()
	data := d.data

	// Lookup the ID for this network. Doubles as a check to see if it exists
	netIDRaw := data.HashGet(fmt.Sprintf("user:%d:networks", userID), netNameLower)
	netID := byteAsInt(netIDRaw)

	if netID == 0 {
		exists = false
		return
	}

	hashKey := fmt.Sprintf("user:%d:network:%s", userID, netID)
	netInfoRaw := d.data.HashGetAll(hashKey)

	if len(netInfoRaw) == 0 {
		exists = false
		return
	}

	netInfo.ID = netID
	netInfo.Name = networkName
	netInfo.Host = string(netInfoRaw["host"])
	netInfo.TLS = byteAsBool(netInfoRaw["tls"])
	netInfo.Port = byteAsInt(netInfoRaw["port"])
	netInfo.AutoConnect = byteAsBool(netInfoRaw["auto_connect"])

	exists = true
	return
}

func (d *Client) SaveNetworkInfo(network NetworkInfo) (savedNetwork NetworkInfo) {
	netNameLower := strings.ToLower(network.Name)
	userID := d.UserID()
	data := d.data

	if network.ID == 0 {
		// New network. Add it.
		netID := 0
		// TODO: Better ID generation
		rand.Seed(int64(rand.Intn(10000)))
		for i := 0; i < 200; i++ {
			netID += rand.Intn(1000)
		}

		hashKey := fmt.Sprintf("user:%d:network:%s", userID, netID)
		data.HashSet(hashKey, "id", intAsByte(netID))
		data.HashSet(hashKey, "host", []byte(network.Host))
		data.HashSet(hashKey, "port", intAsByte(network.Port))
		data.HashSet(hashKey, "tls", boolAsByte(network.TLS))
		data.HashSet(hashKey, "auto_connect", boolAsByte(network.AutoConnect))

		// Add the new network into the users network ID list
		data.HashSet(fmt.Sprintf("user:%d:networks", userID), netNameLower, intAsByte(netID))

		savedNetwork = network
		savedNetwork.ID = netID
	} else {
		hashKey := fmt.Sprintf("user:%d:network:%s", userID, network.ID)

		data.HashSet(hashKey, "host", []byte(network.Host))
		data.HashSet(hashKey, "port", intAsByte(network.Port))
		data.HashSet(hashKey, "tls", boolAsByte(network.TLS))
		data.HashSet(hashKey, "auto_connect", boolAsByte(network.AutoConnect))

		savedNetwork = network
	}

	return
}

func (d *Client) ListNetworks() (networks []NetworkInfo) {
	storedNets := d.data.HashGetAll("user:" + strconv.Itoa(d.UserID()) + ":networks")
	for netName, _ := range storedNets {
		net, exists := d.GetNetworkInfo(netName)
		if exists {
			networks = append(networks, net)
		}
	}

	return
}

func registerClient(c *Client) {
	println("registerClient()")
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
