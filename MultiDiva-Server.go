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
	SERVER_TYPE    = "tcp"
	SERVER_VERSION = "0.1.0"
)

type Room struct {
	RoomTitle         string
	Publicity         bool
	PasswordProtected bool
	Password          string // TODO make this more secure
	Leader            Client
	Members           []Client
}

type Client struct {
	Username   string
	Connection net.Conn
	RoomID     int
}

var rooms []Room
var clients []Client
var serverQuitting bool
var server net.Listener

func main() {
	var err error
	myConfig := configManager.LoadConfig()
	fmt.Println("MultiDiva " + SERVER_VERSION + " server running!")

	server, err = net.Listen(SERVER_TYPE, myConfig.BindAddress+":"+myConfig.Port)
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
			rooms[c.RoomID].removeClient(c)

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

func newClient(c Client) {
	if !serverQuitting {
		dat := c.getJsonMessage()

		instruction := dat["Instruction"].(string)

		fmt.Println("FIRST INSTRUCTION: " + instruction)

		switch instruction {
		case "clientLogout":
			// Handle
		case "login":
			c.Username = dat["username"].(string)

			clients = append(clients, c)

			go clientListener(c)

		default:
			c.sendInstruction("invalidLogin")

		}
	} else {
		c.sendInstruction("serverClosing")
	}
}

func (c Client) getRawMessage() (clientMessageBytes []byte) {
	buffer := make([]byte, 1024)
	clientMessage, err := c.Connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		// TODO bandaid fix until i can properly send a logout from client, windows uses "forcibly closed" while linux uses "EOF"
		if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
			fmt.Println("unexpected closure")
			clientMessageBytes = []byte("{\"Instruction\":\"clientLogout\"}")
		}
	} else {
		clientMessageBytes = buffer[:clientMessage]
	}

	fmt.Println("Received: ", clientMessageBytes)
	return
}

func (c Client) getJsonMessage() (data map[string]interface{}) {
	clientMessageBytes := c.getRawMessage()

	if err := json.Unmarshal(clientMessageBytes, &data); err != nil {
		panic(err)
	}

	return
}

func (c Client) sendMessage(message []byte) {
	_, err := c.Connection.Write(message)
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}
}

func (r Room) sendToOthersInRoom(message []byte, c Client) {
	for i := range r.Members {
		if r.Members[i] != c {
			r.Members[i].sendMessage(message)
		}
	}
}

func (r Room) sendToAllInRoom(message []byte) {
	for i := range r.Members {
		r.Members[i].sendMessage(message)
	}
}

func (r Room) removeClient(c Client) {
	for i := range r.Members {
		if r.Members[i] == c {
			r.Members[i] = r.Members[len(r.Members)-1]
			r.Members = r.Members[:len(r.Members)-1]
		}
	}
}

func (c Client) sendInstruction(instruction string) {
	c.sendMessage([]byte("{\"Instruction\":\"" + instruction + "\"}"))
}

func newRoom(roomName string, publicity bool, leader Client) {
	r := Room{roomName, publicity, false, "", leader, make([]Client, 0)}

	rooms = append(rooms, r)
}
