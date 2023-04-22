package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"

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

		clientMessageString := string(clientMessageBytes)
		var instructions []string
		instructions = append(instructions, clientMessageString)

		// Sometimes, if the client sends multiple messages too fast, the server reads multiple instructions as one.
		// This cause the json unmarshal to panic, and crash the server. Solution, split the message by closing braces.
		// This isn't the best solution, but it works so
		if strings.Count(clientMessageString, "}") > 1 {
			fmt.Println("FAILURE")
			fmt.Println(instructions)
			instructions = strings.Split(clientMessageString, "}")
			instructions = instructions[:len(instructions)-1]
			fmt.Println("POST")
			fmt.Println(instructions)
			for i := range instructions {
				instructions[i] += "}"
			}
			fmt.Println("POST POST")
			fmt.Println(instructions)
		}

		for i := range instructions {
			if err := json.Unmarshal([]byte(instructions[i]), &dat); err != nil {
				panic(err)
			}

			instruction := dat["Instruction"].(string)

			fmt.Println("INSTRUCTION: " + instruction)

			switch instruction {
			case "clientLogout":
				if c.RoomID != -1 {
					rooms[c.RoomID].removeClient(c)
				}

				for i := range clients {
					if clients[i].Connection == c.Connection { // Testing for connection because it's something static for a client, that is also unique to it
						clients[i] = clients[len(clients)-1]
						clients = clients[:len(clients)-1]
					}
				}

				break listenLoop
			case "createRoom":
				if dat["passwordProtected"] == "false" {
					publicRoom, _ := strconv.ParseBool(dat["publicRoom"].(string))
					newRoom(dat["roomName"].(string), publicRoom, &c)
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
				if c.RoomID != -1 {
					rooms[c.RoomID].sendToOthersInRoom(clientMessageBytes, c)
				}
			case "leaveRoom":
				if c.RoomID != -1 {
					rooms[c.RoomID].removeClient(c)
				}
			}
		}

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
