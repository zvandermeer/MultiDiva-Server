package main

import (
	"fmt"
	"net"
)

func startListening(serverData *Server, myConfig ConfigData) {
	var err error

	serverData.listener, err = net.Listen("tcp", myConfig.BindAddress+":"+myConfig.Port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
	}

	fmt.Println("Listening on " + myConfig.BindAddress + ":" + myConfig.Port)

	for {
		fmt.Println("Waiting for client...")

		connection, err := serverData.listener.Accept()
		if err != nil {
			fmt.Println("Error connecting to client: ", err.Error())
			break
		}	

		c := newClient(connection)

		if(c == nil) {
			connection.Close()
		} else {
			go c.processPackets(serverData)

			serverData.clients = append(serverData.clients, c)
		}
	}
}