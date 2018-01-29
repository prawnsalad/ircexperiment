package worker

import (
	"errors"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"
	"time"

	"../common"
)

func RunRpcWorker(worker *WorkerProces, masterSock string) error {
	var rpcClient *rpc.Client

	if masterSock == "std" {
		// Since RPC calls are made over stdin/out, redirect the logger output to stderr
		log.SetOutput(os.Stderr)

		r := io.ReadCloser(os.Stdin)
		w := io.WriteCloser(os.Stdout)
		appPipe := &common.PipePair{
			Reader: &r,
			Writer: &w,
		}
		rpcClient = rpc.NewClient(io.ReadWriteCloser(appPipe))

	} else if strings.HasPrefix(masterSock, "unix:") {
		sock := strings.Replace(masterSock, "unix:", "", 1)
		conn, err := net.Dial("unix", sock)
		if err != nil {
			return errors.New("could not connect to master socket")
		}
		rpcClient = rpc.NewClient(conn)
	}

	worker.SetRpcClient(rpcClient)

	log.Print("Starting worker")

	for {
		if getEvent(worker, rpcClient) {
			continue
		} else {
			break
		}
	}

	log.Print("Ending worker")
	return nil
}

func getEvent(worker *WorkerProces, rpcClient *rpc.Client) bool {
	var e interface{}
	err := rpcClient.Call("queue.Get", "", &e)
	if err != nil {
		println("queue.Get()", err.Error())
		return false
	}

	event := e.(common.RpcEvent)
	if event.Name == "" {
		// No events to process yet. Wait a bit before asking for another event
		time.Sleep(time.Millisecond * 500)
		return true
	}

	//println("Worker got an event", event.Name)

	// TODO: If the event has a clientID, workerID = mod(num_works clientID), execute on workerID
	//       This will keep all single clients actions being processed in order on the same goroutine
	worker.HandleRpcEvent(event)
	println("handled event.")
	return true
}
