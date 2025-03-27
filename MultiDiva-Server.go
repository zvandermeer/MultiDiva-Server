package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
)

const (
	ServerAPIVersion   = 0
	ServerPatchVersion = 1
)

type Server struct {
	listener net.Listener
	rooms []*Room
	clients []*Client
}

func main() {
	myConfig := LoadConfig()
	fmt.Println("MultiDiva " + strconv.Itoa(ServerAPIVersion) + "." + strconv.Itoa(ServerPatchVersion) + " server starting...")

	serverData := Server{}

	InterruptSignal := make(chan os.Signal, 1)
	signal.Notify(InterruptSignal, os.Interrupt)
	go func() {
		for range InterruptSignal {
			closeServer(serverData)
		}
	}()

	startListening(&serverData, myConfig)
}

func closeServer(serverData Server) {
	fmt.Println("\nClosing server...")

	for _, client := range serverData.clients {
		client.closeClient()
	}

	err := serverData.listener.Close()
	if err != nil {
		return
	}
}


//func listRooms() {
//	var myRooms []Room
//
//	for i := range rooms {
//		m := map[string]string{
//			"Name":      rooms[i].RoomTitle,
//			"connected": strconv.Itoa(len(rooms[i].Members)),
//		}
//	}
//
//	m := map[string]string{
//		"Instruction": "ServerList",
//		"":            "John Doe",
//	}
//}
