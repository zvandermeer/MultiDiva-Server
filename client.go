package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type ScoreData struct {
	Score  int
	Combo  int
	Health int
}

type Client struct {
	Connection            net.Conn
	ClientUUID            string
	IncomingMessageBuffer chan map[string]interface{}
	OutgoingMessageBuffer chan []byte
	Username              string
	MyRoom                *Room
	CurrentScore          ScoreData
}

func newClient(thisConn net.Conn) *Client {
	c := Client{}

	c.Connection = thisConn
	c.IncomingMessageBuffer = make(chan map[string]interface{}, 20)
	c.OutgoingMessageBuffer = make(chan []byte, 20)
	c.MyRoom = nil

	go c.listen()
	go c.write()

	dat := <-c.IncomingMessageBuffer

	instruction := dat["Instruction"].(string)

	fmt.Println("FIRST INSTRUCTION: " + instruction)

	switch instruction {
	case "clientLogout":
		return nil
	case "login":
		ClientAPIVersion, _ := strconv.Atoi(dat["MajorClientVersion"].(string))
		if ClientAPIVersion == ServerAPIVersion {
			c.Username = dat["Username"].(string)
			c.ClientUUID = dat["UUID"].(string)

			c.sendSimpleInstruction("loginSuccess")
		} else {
			m := map[string]string{
				"Instruction":        "versionMismatch",
				"ServerAPIVersion":   strconv.Itoa(ServerAPIVersion),
				"ServerPatchVersion": strconv.Itoa(ServerPatchVersion),
			}

			c.sendJsonMessage(m)
		}

		return &c

	default:
		c.sendSimpleInstruction("invalidLogin")

		return nil
	}
}

func (c Client) listen() {
	buffer := make([]byte, 1024)
	for {
		clientMessage, err := c.Connection.Read(buffer)

		var clientMessageBytes []byte

		if err != nil {
			fmt.Println("Error reading:", err.Error())
			// TODO band aid fix until i can properly send a logout from client, windows uses "forcibly closed" while linux uses "EOF"
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
				fmt.Println("unexpected closure")
				c.closeClient()
			}

			break
		} else {
			clientMessageBytes = buffer[:clientMessage]
		}

		clientMessageStr := string(clientMessageBytes)

		if strings.Count(clientMessageStr, "\n") > 1 {
			messages := strings.Split(clientMessageStr, "\n")

			for _, message := range messages {
				c.processJSON([]byte(message))
			}
		} else {
			c.processJSON(clientMessageBytes)
		}
	}
}

func (c Client) write() {
	message := <-c.OutgoingMessageBuffer

	_, err := c.Connection.Write(message)
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}
}

func (c Client) processJSON(jsonMessage []byte) {
	var data map[string]interface{}

	if err := json.Unmarshal(jsonMessage, &data); err != nil {
		panic(err)
	} else {
		c.IncomingMessageBuffer <- data
	}
}

func (c *Client) processPackets(serverData *Server) {
	var connected = true

	for connected {
		data := <-c.IncomingMessageBuffer

		instruction := data["Instruction"].(string)

		fmt.Println("INSTRUCTION: " + instruction)

		switch instruction {
		case "clientLogout":
			if c.MyRoom != nil {
				c.MyRoom.removeClient(*c)
			}

			for j := range serverData.clients {
				if serverData.clients[j].ClientUUID == c.ClientUUID {
					serverData.clients[j] = serverData.clients[len(serverData.clients)-1]
					serverData.clients = serverData.clients[:len(serverData.clients)-1]
					break
				}
			}

			connected = false

			c.closeClient()

		case "createRoom":
			passwordProtected, _ := strconv.ParseBool(data["passwordProtected"].(string))

			var thisRoom *Room

			if passwordProtected {
				thisRoom = newRoom(data["roomName"].(string), c)
			} else {
				thisRoom = newProtectedRoom(data["roomName"].(string), c, data["password"].(string))
			}

			serverData.rooms = append(serverData.rooms, thisRoom)

			m := map[string]string{
				"Instruction": "roomConnectionUpdate",
				"Status":      "connectedAsLeader",
				"RoomName":    thisRoom.RoomTitle,
			}

			c.sendJsonMessage(m)

		case "joinRoom":
			foundRoom := false

			if c.MyRoom == nil {
				for _, room := range serverData.rooms {
					if room.RoomUUID == data["roomID"] {
						room.Members = append(room.Members, c)
						foundRoom = true

						c.MyRoom = room

						break
					}
				}

				if !foundRoom {
					m := map[string]string{
						"Instruction": "roomConnectionUpdate",
						"Status":      "roomNotFound",
						"RoomName":    data["roomName"].(string),
						"RoomID":      data["roomID"].(string),
					}

					c.sendJsonMessage(m)
				} else {
					m := map[string]string{
						"Instruction": "roomConnectionUpdate",
						"Status":      "connectedToRoom",
						"RoomName":    data["roomName"].(string),
						"RoomID":      data["roomID"].(string),
					}

					c.sendJsonMessage(m)
				}
			}

		case "leaveRoom":
			if c.MyRoom != nil {
				c.MyRoom.removeClient(*c)
			}

		case "songStart":
			c.MyRoom.GameRunning = true

			go c.MyRoom.updateRankings()

			c.MyRoom.sendToAllInRoom([]byte("instruction: game start")) // add song data etc to this

		case "songFinished":
			// TODO create json for all users final scores, expect to get this signal from leader

			time.Sleep(5 * time.Second)

			c.MyRoom.sendToAllInRoom([]byte("{\"Instruction\":\"FinalScore\"}"))

			c.MyRoom.GameRunning = false

		case "note":
			if c.MyRoom != nil {
				rank := -1

				c.CurrentScore.Score, _ = strconv.Atoi(data["Score"].(string))
				c.CurrentScore.Combo, _ = strconv.Atoi(data["Combo"].(string))
				c.CurrentScore.Health, _ = strconv.Atoi(data["Health"].(string))

				for j := range c.MyRoom.Ranking {
					if c.MyRoom.Ranking[j].ClientUUID == c.ClientUUID {
						rank = j
						break
					}
				}

				data["Ranking"] = strconv.Itoa(rank)

				messageToSend, err := json.Marshal(data)
				if err != nil {
					panic(err)
				}
				c.MyRoom.sendToOthersInRoom(messageToSend, *c)
			}
		}
	}
}

func (c Client) closeClient() {
	c.sendSimpleInstruction("serverClosing")
	c.Connection.Close()
}

func (c Client) sendSimpleInstruction(instruction string) {
	c.OutgoingMessageBuffer <- []byte("{\"Instruction\":\"" + instruction + "\"}")
}

func (c Client) sendJsonMessage(message map[string]string) {
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
	}

	c.OutgoingMessageBuffer <- data
}
