package master

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"

	"../common"
	"github.com/kardianos/osext"
)

func RunRpcServer(proc *MasterProcess, workerPath, masterSock string) {
	rpc.RegisterName("queue", proc.EventQueue)
	rpc.RegisterName("conns", &RpcConnInterface{MasterProc: proc})
	rpc.RegisterName("data", NewRpcDataInterface())

	if masterSock == "" {
		masterSock = "unix:/tmp/go.sock"
		masterSock = "std"
	}

	if workerPath == "" {
		workerPath, _ = osext.Executable()
	}
	exArgs := []string{"--worker=" + masterSock}

	println("Using worker binary", workerPath)

	for {
		if masterSock == "std" {
			app := exec.Command(workerPath, exArgs...)
			app.Stderr = os.Stdout

			stdIn, _ := app.StdinPipe()
			stdOut, _ := app.StdoutPipe()

			appPipe := &common.PipePair{
				Reader: &stdOut,
				Writer: &stdIn,
			}

			println("Serving worker process")
			err := app.Start()
			if err != nil {
				log.Fatal(err.Error())
			}

			rpc.ServeConn(appPipe)

		} else if strings.HasPrefix(masterSock, "unix:") {
			sock := strings.Replace(masterSock, "unix:", "", 1)
			ln, err := net.Listen("unix", sock)
			if err != nil {
				log.Fatal("Listen error: ", err)
			}

			go rpc.Accept(ln)

			app := exec.Command(workerPath, exArgs...)
			app.Stderr = os.Stderr

			app.Start()
			app.Wait()
		}

		println("Worker disconnected from main process. Restarting worker.")
		time.Sleep(time.Second)
	}
}
