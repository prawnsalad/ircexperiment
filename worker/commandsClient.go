package worker

import (
	"fmt"
	"os"
	"strings"

	"../common"
	"github.com/kiwiirc/webircgateway/pkg/irc"
)

type Command struct {
	Fn                 func(*Client, *irc.Message)
	Command            string
	RequiresRegistered bool
	RequiresUMode      string
}

func runClientCommand(worker *WorkerProces, clientID int, msg *irc.Message) {
	cmd, cmdExists := worker.ClientCommands[strings.ToUpper(msg.Command)]
	if !cmdExists {
		worker.WriteClient(clientID, "NOT_FOUND "+msg.Command)
		return
	}

	//c := NewDataWrapperClient(worker.Data, clientID)
	c := NewClient(worker, clientID)

	if cmd.RequiresRegistered && !c.HasAuthed() {
		worker.WriteClient(clientID, "REQUIRES_REG")
		return
	}

	clientModes := byteSliceAsStrings(worker.Data.ClientSGet(clientID, "modes"))
	if cmd.RequiresUMode != "" && !contains(clientModes, cmd.RequiresUMode) {
		worker.WriteClient(clientID, "NEEDS_MODE %s", cmd.RequiresUMode)
		return
	}

	cmd.Fn(c, msg)
	/*
		TODO: if not authed, now we have all the info to auth, then auth
		if !client.Registered && !cmd.RequiresRegistered && client.CanRegister() {
			registerClient(client)
			worker.Data.ClientSet(clientID, DbClientKeyRegistered, boolAsInt(true))
			worker.Data.ClientModeSet(clientID, "i", boolAsByte(true))
		}
	*/
}

func loadClientCommands(worker *WorkerProces) (commands map[string]Command) {
	commands = make(map[string]Command)

	commands["KILL"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			os.Exit(0)
		},
	}

	commands["PING"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			c.WriteWithServerPrefix("PONG %s", messageParam(m, 0))
		},
	}

	commands["USER"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			//username := messageParam(m, 0)
			//realname := messageParam(m, 3)
		},
	}

	commands["NICK"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			nick := m.Params[0]
			c.SetNick(nick)
		},
	}

	commands["PASS"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			pass := messageParam(m, 0)
			var username, network, password string
			fmt.Sscanf(pass, "%s/%s:%s", &username, &network, &password)

			// TODO: Auth here with persistent storage layer
			// TODO: Get the network ID from the network name here
			netID := 1
			c.SetActiveNetID(netID)

			net := c.NetworkData()
			c.Write(":%s NICK %s", c.Nick(), net.Nick())
			c.SetNick(net.Nick())

			// TODO: Dump network reg info here
		},
	}

	commands["JOIN"] = Command{
		RequiresRegistered: true,
		Fn: func(c *Client, m *irc.Message) {
			// chanName := messageParam(m, 0)
			// password := messageParam(m, 1)
		},
	}
	/*
		commands["SPAM"] = Command{
			RequiresRegistered: false,
			Fn: func(c *Client, m *irc.Message) {
				for i := 1; i < 100; i++ {
					go SimUser("nick"+strconv.Itoa(i), c)
				}
			},
		}
	*/

	commands["WGET"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			var err error
			worker.RpcCall("conns.Open", common.RpcEventConnState{
				RAddress: "irc.freenode.net:6667",
				Tls:      false,
			}, &err)

			if err != nil {
				println(err.Error())
			}
		},
	}
	return
}
