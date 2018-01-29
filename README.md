# ircexperiment
Experimental IRC server. Reloadable 0 downtime in golang

## What's the experiment?
IRC connections are stateful. This means that if the IRC server dies then the client connections also die. This complicates server restarts and updates.

This project aims to get around this by running a master process that accepts connections and buffers its incoming data while also storing state data relating to the server and clients. A worker process is spawned that requests client events from the master process, such as connections, disconnections and received data. This worker processes the events as they occur and may store or retrieve data from the master process.

All communication is done via RPC between the master and worker processes; This may be either a pipe between the two processes or via a unix socket.

If the worker process dies due to an error or killed by a server admin, the master process automatically respawns a new worker process to continue where the previous left off.

Since the worker process is respawned automatically, this allows the server admin to update the server without any client downtime by replacing the server executable before killing the worker process.
