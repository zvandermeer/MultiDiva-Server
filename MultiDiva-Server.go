package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	SERVER_HOST    = "0.0.0.0"
	SERVER_TYPE    = "tcp"
	SERVER_VERSION = "0.1.0"
)

var channelList []chan string
var clientsConnected int
var newClientNum int
var serverQuitting bool

func main() {
	loadConfig()
	fmt.Println("Server Running...")

	server_port := viper.GetString("network.default_port")

	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+server_port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
	}

	fmt.Println("Listening on " + SERVER_HOST + ":" + server_port)

	clientsConnected = 0

	InterruptSignal := make(chan os.Signal, 1)
	signal.Notify(InterruptSignal, os.Interrupt)
	go func() {
		for range InterruptSignal {
			closeServer(server, InterruptSignal, channelList)
		}
	}()

	for {
		fmt.Println("Waiting for client...")

		connection, err := server.Accept()
		if !serverQuitting {
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
			}

			userChannel := make(chan string)

			channelList = append(channelList, userChannel)

			go processClient(connection, userChannel, clientsConnected, newClientNum)
			clientsConnected = clientsConnected + 1
			newClientNum = newClientNum + 1
			fmt.Println("client connected")
		} else {
			break
		}
	}
}

func processClient(connection net.Conn, userChannel chan string, totalClients int, clientNum int) {
	go receiveMessages(connection, userChannel, totalClients, clientNum)
	go sendMessages(connection, userChannel, totalClients, clientNum)
}

func receiveMessages(connection net.Conn, userChannel chan string, totalClients int, clientNum int) {

	for {
		var stringClientMessage string

		buffer := make([]byte, 1024)
		clientMessage, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			// TODO bandaid fix until i can properly send a logout from client, windows uses "forcibly closed" while linux uses "EOF"
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
				fmt.Println("unexpected closure")
				stringClientMessage = "/clientLogout"
			}
		} else {
			stringClientMessage = string(buffer[:clientMessage])
		}

		fmt.Println("Received: ", stringClientMessage, "from client", clientNum)

		if stringClientMessage == "/clientLogout" {
			userChannel <- "/closePipe"
			time.Sleep(10 * time.Millisecond)
			i := 0
			for {
				if channelList[i] == userChannel {
					channelList[i] = channelList[len(channelList)-1]
					channelList[len(channelList)-1] = nil
					channelList = channelList[:len(channelList)-1]
					break
				}
				i = i + 1
			}
			clientsConnected = clientsConnected - 1
			break
		}
		for _, item := range channelList {
			if item != userChannel {
				item <- stringClientMessage
			}
		}
	}
	connection.Close()
}

func sendMessages(connection net.Conn, userChannel chan string, totalClients int, clientNum int) {

	for {

		fmt.Printf("Client %v waiting for message...\n", clientNum)
		message := <-userChannel
		_, err := connection.Write([]byte(message))
		if err != nil {
			fmt.Println("Error writing:", err.Error())
		}
		fmt.Printf("Sent '%s' to client %v !\n", message, clientNum)
		if message == "/closePipe" {
			break
		}
	}
}

func closeServer(server net.Listener, exitSignal chan os.Signal, channelList []chan string) {
	fmt.Println("\nServer shutting down...")

	serverQuitting = true

	for _, item := range channelList {
		item <- "[server] Server shutting down..."
		time.Sleep(1 * time.Millisecond)
		item <- "/closePipe"
	}

	server.Close()
}

func loadConfig() {
	viper.SetConfigName("MultiDiva_server_config") // config file name without extension
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("default_port", "9988")
	viper.SetDefault("max_rooms", 15)
	viper.SetDefault("max_room_size", 6)
	viper.SetDefault("max_concurrent_users", 100)

	viper.SetDefault("config_version", 1)

	usingConfig := true

	if _, err := os.Stat("./MultiDiva_server_config.yml"); os.IsNotExist(err) {
		fmt.Println("Error reading config: MultiDiva_server_config.yml does not exist. Attempting to create one...")
		if _, err := os.Create("./MultiDiva_server_config.yml"); err != nil {
			fmt.Println("Error creating MultiDiva_server_config.yml:", err)
			usingConfig = false
		} else {
			viper.WriteConfigAs("./MultiDiva_server_config.yml")
			fmt.Println("Config created successfully!")
		}
	}

	if usingConfig {
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Println("Error reading config:", err)
		}
	}
}
