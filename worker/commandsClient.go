package worker

import (
	"fmt"
	"net"
	"os"
	"strconv"
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
			// username := messageParam(m, 0)
			// realname := messageParam(m, 3)
		},
	}

	commands["NICK"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			nick := m.Params[0]
			c.SetNick(nick)
		},
	}

	commands["REG"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			// REG username password
			username := strings.ToLower(messageParam(m, 0))
			password := messageParam(m, 1)

			err := RegisterAccount(c.data, username, password)
			if err != nil {
				c.WriteStatus(err.Error())
			} else {
				c.WriteStatus("Account registered")
			}

		},
	}

	commands["PASS"] = Command{
		RequiresRegistered: false,
		Fn: func(c *Client, m *irc.Message) {
			pass := messageParam(m, 0)
			var username, network, password string

			// Split the user/network:password format before passing them to Auth
			parts := strings.SplitN(pass, ":", 2)
			if len(parts) == 2 {
				password = parts[1]
			}

			parts = strings.SplitN(parts[0], "/", 2)
			username = parts[0]
			if len(parts) == 2 {
				network = parts[1]
			}

			if !c.Auth(username, password) {
				println("PASS fail")
				c.WriteStatus("Invalid password")
				return
			}

			println("PASS OK", network)

			if network != "" {
				netInfo, netExists := c.GetNetworkInfo(network)
				if !netExists {
					c.WriteStatus("Network not found")
					return
				}

				println("Setting active net ID to " + strconv.Itoa(netInfo.ID))
				c.SetActiveNetID(netInfo.ID)

				net := c.NetworkData()
				c.Write(":%s NICK %s", c.Nick(), net.Nick())
				c.SetNick(net.Nick())

				// TODO: Dump network reg info here
			} else {
				c.WriteStatus("Woo, logged in!")
			}
		},
	}

	commands["PRIVMSG"] = Command{
		RequiresRegistered: true,
		Fn: func(c *Client, m *irc.Message) {
			target := messageParam(m, 0)
			if target != "*status" {
				return
			}

			if !c.HasAuthed() {
				c.WriteStatus("Must login first")
				return
			}

			message := messageParam(m, 1)
			parts := strings.Split(message, " ")
			arg := func(idx int) string {
				// This command args starts at idx 2
				return strSliceIdx(parts, idx)
			}
			if arg(0) == "network" && arg(1) == "add" {
				// network add freenode irc.freenode.net:+6667
				netName := arg(2)
				server := arg(3)

				hostStr, portStr, err := net.SplitHostPort(server)
				if err != nil || hostStr == "" {
					c.WriteStatus("Invalid server")
					return
				}

				host := hostStr
				port := 6667
				tls := false

				if portStr[0] == '+' {
					tls = true
					portStr = portStr[1:]
				}

				givenPort, _ := strconv.Atoi(portStr)
				if givenPort > 0 {
					port = givenPort
				}

				_, netExists := c.GetNetworkInfo(netName)
				if netExists {
					c.WriteStatus("That network already exists")
					return
				}

				newNet := NetworkInfo{}
				newNet.Name = netName
				newNet.Host = host
				newNet.Port = port
				newNet.TLS = tls
				newNet.AutoConnect = false

				savedNetwork := c.SaveNetworkInfo(newNet)

				c.WriteStatus(fmt.Sprintf("Network saved! (%d/%s)", savedNetwork.ID, savedNetwork.Name))
				return
			}

			if arg(0) == "network" && arg(1) == "" {
				nets := c.ListNetworks()
				for _, net := range nets {
					c.WriteStatus(fmt.Sprintf("%s %d", net.Name, net.ID))
				}
			}
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
