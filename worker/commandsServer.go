package worker

import (
	"strings"

	"github.com/kiwiirc/webircgateway/pkg/irc"
)

type ServerCommand struct {
	Fn      func(*DataWrapperNetwork, *irc.Message)
	Command string
}

func runServerCommand(worker *WorkerProces, connID int, msg *irc.Message) {
	cmd, cmdExists := worker.ServerCommands[strings.ToUpper(msg.Command)]
	if !cmdExists {
		return
	}

	net := NetworkFromConnID(worker.Data, connID)
	cmd.Fn(net, msg)
}

func loadServerCommands(worker *WorkerProces) (commands map[string]ServerCommand) {
	commands = make(map[string]ServerCommand)

	commands["001"] = ServerCommand{
		Fn: func(net *DataWrapperNetwork, m *irc.Message) {
			nick := messageParam(m, 0)
			net.SetNick(nick)
			net.SetRegistered(true)
		},
	}

	commands["NICK"] = ServerCommand{
		Fn: func(net *DataWrapperNetwork, m *irc.Message) {
			if m.Prefix.Nick == net.Nick() {
				newNick := messageParam(m, 0)
				net.SetNick(newNick)
			}
		},
	}

	return
}
