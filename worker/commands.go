package worker

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/kiwiirc/webircgateway/pkg/irc"
)

type Command struct {
	Fn                 func(*RClient, *irc.Message)
	Command            string
	RequiresRegistered bool
	RequiresUMode      string
}

func runCommand(worker *WorkerProces, clientID int, msg *irc.Message) {
	cmd, cmdExists := worker.Commands[strings.ToUpper(msg.Command)]
	if !cmdExists {
		worker.WriteClient(clientID, "NOT_FOUND "+msg.Command)
		return
	}

	client := NewRClient(worker)
	client.Load(clientID)

	if cmd.RequiresRegistered && !client.Registered {
		worker.WriteClient(clientID, "REQUIRES_REG")
		return
	}

	clientModes := byteSliceAsStrings(worker.Data.ClientSGet(clientID, "modes"))
	if cmd.RequiresUMode != "" && !contains(clientModes, cmd.RequiresUMode) {
		worker.WriteClient(clientID, "NEEDS_MODE %s", cmd.RequiresUMode)
		return
	}

	cmd.Fn(client, msg)

	if !client.Registered && !cmd.RequiresRegistered && client.CanRegister() {
		registerClient(client)
		worker.Data.ClientSet(clientID, "registered", boolAsInt(true))
		worker.Data.ClientModeSet(clientID, "i", boolAsByte(true))
	}
}

func loadCommands(worker *WorkerProces) (commands map[string]Command) {
	commands = make(map[string]Command)

	commands["KILL"] = Command{
		RequiresRegistered: false,
		Fn: func(c *RClient, m *irc.Message) {
			os.Exit(0)
		},
	}

	commands["PING"] = Command{
		RequiresRegistered: false,
		Fn: func(c *RClient, m *irc.Message) {
			c.WriteWithPrefix("PONG %s", messageParam(m, 1))
		},
	}

	commands["USER"] = Command{
		RequiresRegistered: false,
		Fn: func(c *RClient, m *irc.Message) {
			username := messageParam(m, 0)
			realname := messageParam(m, 3)
			worker.Data.ClientSet(c.Id, "username", []byte(username))
			worker.Data.ClientSet(c.Id, "realname", []byte(realname))
			worker.Data.ClientSet(c.Id, "hostname", []byte("host.isp"))
			c.Username = username
			c.RealName = realname
		},
	}

	commands["NICK"] = Command{
		RequiresRegistered: false,
		Fn: func(c *RClient, m *irc.Message) {
			nick := m.Params[0]
			// TODO: Make sure nick is not already in use
			worker.Data.ClientSet(c.Id, "nick", []byte(nick))
			c.Nick = nick
			if c.Registered {
				c.WriteWithPrefix("NICK %s", c.Nick)
			}
		},
	}

	commands["JOIN"] = Command{
		RequiresRegistered: true,
		Fn: func(c *RClient, m *irc.Message) {
			chanName := messageParam(m, 0)
			password := messageParam(m, 1)

			channel := NewRChannel(worker)
			channel.Load(chanName)

			if channel.Modes.Get("K") != password {
				c.WriteWithServerPrefix("invalid chan password")
			} else {
				worker.Data.ChannelSAdd(chanName, "clients", intAsByte(c.Id))
				c.WriteWithPrefix("JOIN %s", chanName)
			}
		},
	}

	commands["NAMES"] = Command{
		RequiresRegistered: true,
		Fn: func(c *RClient, m *irc.Message) {
			chanName := messageParam(m, 0)

			channel := NewRChannel(worker)
			channel.Load(chanName)

			if !channel.IsClientOn(c.Id) {
				c.WriteWithServerPrefix("Not in that channel")
				return
			}

			masks := []string{}
			for _, id := range channel.GetClients() {
				client := RClient{}
				client.Load(id)
				mask := fmt.Sprintf(":%s!%s@%s", client.Nick, client.Username, client.Hostname)
				masks = append(masks, mask)
			}

			//353 Guest52 = #test :Guest52!~kiwiidev@10.0.5.1 NotPrawn1!~darren@10.0.5.1
			for _, line := range joinStringsWithMaxLength(masks, 300, " ") {
				c.WriteWithServerPrefix("353 %s = %s :%s", c.Nick, channel.Name, line)
			}

			c.WriteWithServerPrefix("366 %s %s :End of NAMES list", c.Nick, channel.Name)
		},
	}

	commands["PRIVMSG"] = Command{
		RequiresRegistered: true,
		Fn: func(c *RClient, m *irc.Message) {
			// PRIVMSG #test :message
			chanName := messageParam(m, 0)
			channel := NewRChannel(worker)
			channel.Load(chanName)

			if !channel.IsClientOn(c.Id) {
				c.WriteWithServerPrefix("Not in that channel")
				return
			}

			mask := fmt.Sprintf(":%s!%s@%s", c.Nick, c.Username, c.Hostname)
			data := messageParam(m, 1)
			for _, id := range channel.GetClients() {
				line := fmt.Sprintf(":%s PRIVMSG %s :%s", mask, channel.Name, data)
				worker.WriteClient(id, line)
			}
		},
	}

	commands["TOPIC"] = Command{
		RequiresRegistered: true,
		Fn: func(c *RClient, m *irc.Message) {
			// TOPIC #test :new topic
			chanName := messageParam(m, 0)
			channel := NewRChannel(worker)
			channel.Load(chanName)

			if !channel.IsClientOn(c.Id) {
				c.WriteWithServerPrefix("Not in that channel")
				return
			}

			newTopic := messageParam(m, 1)
			if newTopic != "" {
				worker.Data.ChannelSet(chanName, "topic", []byte(newTopic))
				for _, id := range channel.GetClients() {
					line := fmt.Sprintf("TOPIC %s :%s", channel.Name, newTopic)
					worker.WriteClient(id, line)
				}
			} else {
				topic := string(worker.Data.ChannelGet(chanName, "topic"))
				worker.WriteClient(c.Id, "TOPIC %s :%s", channel.Name, topic)
			}
		},
	}

	commands["AUTH"] = Command{
		RequiresRegistered: false,
		Fn: func(c *RClient, m *irc.Message) {
			user := ""
			pass := ""

			if len(m.Params) == 1 {
				user = c.Nick
				pass = messageParam(m, 0)
			} else {
				user = messageParam(m, 0)
				pass = messageParam(m, 1)
			}

			if user == "" || pass == "" {
				c.Write("invalid args")
				return
			}

			if user == "root" && pass == "12" {
				worker.Data.ClientModeSet(c.Id, "O", []byte{})
			}
		},
	}

	commands["MODE"] = Command{
		RequiresRegistered: false,
		Fn: func(c *RClient, m *irc.Message) {
			//target := messageParam(m, 0)
			//if target == "" {
			//	target == c.User.Nick
			//}

			// TODO: Get the correct client here
			c.Modes.Lock()
			modes := ""
			for mode, _ := range c.Modes.Modes {
				modes += mode
			}
			c.Modes.Unlock()
			worker.WriteClient(c.Id, "MODE %s +%s", c.Nick, modes)
		},
	}
	/*
		commands["SPAM"] = Command{
			RequiresRegistered: false,
			Fn: func(c *RClient, m *irc.Message) {
				for i := 1; i < 100; i++ {
					go SimUser("nick"+strconv.Itoa(i), c)
				}
			},
		}
	*/
	return
}

func SimUser(nick string, c *Client) {
	mask := fmt.Sprintf("%s!%s@host.com", nick, nick)
	fmt.Println("SimUser() " + mask)
	for {
		c.Write(":%s JOIN #chan1", mask)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		c.Write(":%s PART #chan1", mask)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
	}
}
