package main

import (
	"flag"
	"log"

	"./common"
	"./master"
	"./worker"
)

func main() {
	connectMaster := flag.String("worker", "", "")
	workerPath := flag.String("worker-path", "", "Path to the worker binary")
	flag.Parse()

	common.RegisterRpcGobTypes()

	if *connectMaster == "" {
		log.SetPrefix("[master] ")

		// Listens for incoming clients and acts as a data store
		proc := master.NewMasterProcess()
		go master.RunRpcServer(proc, *workerPath, *connectMaster)
		proc.ListenForClients()

	} else {
		log.SetPrefix("[worker] ")

		// Processes incoming events from clients
		proc := worker.NewWorkerProcess()
		proc.LoadCommands()
		worker.RunRpcWorker(proc, *connectMaster)
	}
}
