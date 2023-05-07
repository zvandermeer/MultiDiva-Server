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

		for i := range dat {
			instruction := dat[i]["Instruction"].(string)

			fmt.Println("FIRST INSTRUCTION: " + instruction)

			switch instruction {
			case "clientLogout":
				break
			case "login":
				MajorClientVersion, _ := strconv.Atoi(dat[i]["MajorClientVersion"].(string))
				if MajorClientVersion == MajorServerVersion {
					c.Username = dat[i]["Username"].(string)

					clients = append(clients, c)

					go clientListener(c)

					c.sendInstruction("loginSuccess")
				} else {
					m := map[string]string{
						"Instruction":        "versionMismatch",
						"MajorServerVersion": strconv.Itoa(MajorServerVersion),
						"MinorServerVersion": strconv.Itoa(MinorServerVersion),
					}

					c.sendJsonMessage(m)
				}

			default:
				c.sendInstruction("invalidLogin")
			}
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

	fmt.Println("Received: ", string(clientMessageBytes))
	return
}

func (c Client) getJsonMessage() (returnData []map[string]interface{}) {
	clientMessageBytes := c.getRawMessage()

	clientMessageString := string(clientMessageBytes)
	var messages []string
	messages = append(messages, clientMessageString)

	// Sometimes, if the client sends multiple messages too fast, the server reads multiple instructions as one.
	// This cause the json unmarshal to panic, and crash the server. Solution, split the message by closing braces.
	// This isn't the best solution, but it works so
	if strings.Count(clientMessageString, "}") > 1 {
		fmt.Println("FAILURE")
		fmt.Println(messages)
		messages = strings.Split(clientMessageString, "}")
		messages = messages[:len(messages)-1]
		fmt.Println("POST")
		fmt.Println(messages)
		for i := range messages {
			messages[i] += "}"
		}
		fmt.Println("POST POST")
		fmt.Println(messages)
	}

	for i := range messages {
		var data map[string]interface{}

		if err := json.Unmarshal([]byte(messages[i]), &data); err != nil {
			panic(err)
		}

		returnData = append(returnData, data)
	}

	return
}

func (c Client) sendMessage(message []byte) {
	fmt.Println("Writing: " + string(message))

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
