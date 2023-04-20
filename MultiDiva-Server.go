package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/ovandermeer/MultiDiva-Server/internal/configManager"
)

const (
	MajorServerVersion = 0
	MinorServerVersion = 1
)

var rooms []Room
var clients []Client
var serverQuitting bool
var server net.Listener

func main() {
	var err error
	myConfig := configManager.LoadConfig()
	fmt.Println("MultiDiva " + strconv.Itoa(MajorServerVersion) + "." + strconv.Itoa(MinorServerVersion) + " server running!")

	server, err = net.Listen("tcp", myConfig.BindAddress+":"+myConfig.Port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
	}

	fmt.Println("Listening on " + myConfig.BindAddress + ":" + myConfig.Port)

	InterruptSignal := make(chan os.Signal, 1)
	signal.Notify(InterruptSignal, os.Interrupt)
	go func() {
		for range InterruptSignal {
			closeServer()
		}
	}()

	for !serverQuitting {
		fmt.Println("Waiting for client...")

		connection, err := server.Accept()
		if err != nil && !serverQuitting {
			fmt.Println("Error accepting: ", err.Error())
		}

		go newClient(Client{"", connection, -1})
	}
}

func closeServer() {
	serverQuitting = true
	for i := range clients {
		clients[i].sendInstruction("serverClosing")
	}

	server.Close()
}

func clientListener(c Client) {
listenLoop:
	for {
		clientMessageBytes := c.getRawMessage()

		var dat map[string]interface{}

		if err := json.Unmarshal(clientMessageBytes, &dat); err != nil {
			panic(err)
		}

		instruction := dat["Instruction"].(string)

		fmt.Println("INSTRUCTION: " + instruction)

		switch instruction {
		case "clientLogout":
			if c.RoomID != -1 {
				rooms[c.RoomID].removeClient(c)
			}

			break listenLoop
		case "createRoom":
			if dat["passwordProtected"] == "false" {
				publicity, _ := strconv.ParseBool(dat["publicity"].(string))
				newRoom(dat["roomName"].(string), publicity, c)
			}

		case "joinRoom":
			foundRoom := false
			for i := range rooms {
				if rooms[i].RoomTitle == dat["roomName"] {
					rooms[i].Members = append(rooms[i].Members, c)
					foundRoom = true

					c.RoomID = i

					break
				}
			}

			if !foundRoom {
				c.sendInstruction("roomNotFound")
			}
		case "note":
			rooms[c.RoomID].sendToOthersInRoom(clientMessageBytes, c)
		case "leaveRoom":
			rooms[c.RoomID].removeClient(c)
		}
	}
}
