package worker

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
	format = ":*status!this-is@your.bnc NOTICE " + format
	c.Worker.WriteClient(c.ClientID, format, args...)
}

func (c *Client) WriteWithServerPrefix(format string, args ...interface{}) {
	format = ":*status!this-is@your.bnc " + format
	c.Worker.WriteClient(c.ClientID, format, args...)
}

func (d *Client) HasAuthed() bool {
	if d.UserID() == 0 {
		return false
	} else {
		return true
	}
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
