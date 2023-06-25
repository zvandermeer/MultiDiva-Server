package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"

	"github.com/ovandermeer/MultiDiva-Server/internal/configManager"
)

const (
	MajorServerVersion = 0
	MinorServerVersion = 1
)

type rankingData struct {
	client net.Conn
	score  int
	combo  string
	health string
}

var rooms []Room
var clients []Client
var serverQuitting bool
var server net.Listener

var playerRanking []rankingData

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

	err := server.Close()
	if err != nil {
		return
	}
}

func clientListener(c Client) {
listenLoop:
	for {
		dat := c.getJsonMessage()

		for i := range dat {

			instruction := dat[i]["Instruction"].(string)

			fmt.Println("INSTRUCTION: " + instruction)

			switch instruction {
			case "clientLogout":
				if c.RoomID != -1 {
					rooms[c.RoomID].removeClient(c)
				}

				for j := range clients {
					if clients[j].Connection == c.Connection { // Testing for connection because it's something static for a client, that is also unique to it
						clients[j] = clients[len(clients)-1]
						clients = clients[:len(clients)-1]
						break
					}
				}

				break listenLoop

			case "createRoom":
				if dat[i]["passwordProtected"] == "false" {
					publicRoom, _ := strconv.ParseBool(dat[i]["publicRoom"].(string))
					newRoom(dat[i]["roomName"].(string), publicRoom, &c)
				}

			case "joinRoom":
				foundRoom := false

				if c.RoomID == -1 {
					for j := range rooms {
						if rooms[j].RoomTitle == dat[i]["roomName"] {
							rooms[j].Members = append(rooms[i].Members, c)
							foundRoom = true

							c.RoomID = i

							break
						}
					}

					if !foundRoom {
						m := map[string]string{
							"Instruction": "roomConnectionUpdate",
							"Status":      "roomNotFound",
							"RoomName":    dat[i]["roomName"].(string),
						}

						c.sendJsonMessage(m)
					} else {
						m := map[string]string{
							"Instruction": "roomConnectionUpdate",
							"Status":      "connectedToRoom",
							"RoomName":    dat[i]["roomName"].(string),
						}

						c.sendJsonMessage(m)
					}
				}

			case "note":
				if c.RoomID != -1 {
					initialPos := -1

					for j := range playerRanking {
						if playerRanking[j].client == c.Connection {
							initialPos = j
							break
						}
					}

					if initialPos != -1 {
						playerRanking[initialPos].score, _ = strconv.Atoi(dat[i]["Score"].(string))

						sort.Slice(playerRanking, func(i, j int) bool {
							return playerRanking[i].score > playerRanking[j].score
						})

						finalPos := -1

						for j := range playerRanking {
							if playerRanking[j].client == c.Connection {
								finalPos = j
								break
							}
						}

						if finalPos != -1 {
							dat[i]["ranking"] = strconv.Itoa(finalPos)

							messageToSend, err := json.Marshal(dat[i])
							if err != nil {
								panic(err)
							}
							rooms[c.RoomID].sendToAllInRoom(messageToSend)

							if finalPos < initialPos {
								diff := initialPos - finalPos
								k := 1
								for k <= diff {
									m := map[string]string{
										"Instruction": "note",
										"Combo":       playerRanking[finalPos+k].combo,
										"Grade":       "1",
										"Score":       strconv.Itoa(playerRanking[finalPos+k].score),
										"ranking":     strconv.Itoa(finalPos + k),
									}

									messageToSend, err := json.Marshal(m)
									if err != nil {
										panic(err)
									}
									rooms[c.RoomID].sendToAllInRoom(messageToSend)
									k++
								}
							}

						} else {
							fmt.Println("Error, unable to rank player. Error code: 02")
						}

					} else {
						fmt.Println("Error, unable to rank player. Error code: 01")
					}
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
