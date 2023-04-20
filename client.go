package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	Username   string
	Connection net.Conn
	RoomID     int
}

func newClient(c Client) {
	if !serverQuitting {
		dat := c.getJsonMessage()

		instruction := dat["Instruction"].(string)

		fmt.Println("FIRST INSTRUCTION: " + instruction)

		switch instruction {
		case "clientLogout":
			break
		case "login":
			MajorClientVersion, _ := strconv.Atoi(dat["MajorClientVersion"].(string))
			if MajorClientVersion == MajorServerVersion {
				c.Username = dat["Username"].(string)

				clients = append(clients, c)

				go clientListener(c)
			} else {
				m := map[string]string{
					"Instruction":        "versionMismatch",
					"MajorServerVersion": strconv.Itoa(MajorServerVersion),
				}

				c.sendJsonMessage(m)
			}

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

func (c Client) sendInstruction(instruction string) {
	c.sendMessage([]byte("{\"Instruction\":\"" + instruction + "\"}"))
}

func (c Client) sendJsonMessage(message map[string]string) {
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
	}

	c.sendMessage(data)
}
