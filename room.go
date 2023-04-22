package main

import (
	"fmt"
)

type Room struct {
	RoomTitle         string
	PublicRoom        bool
	PasswordProtected bool
	Password          string // TODO make this more secure
	Leader            Client
	Members           []Client
}

func newRoom(roomName string, publicRoom bool, leader *Client) {
	r := Room{roomName, publicRoom, false, "", *leader, make([]Client, 0)}

	r.Members = append(r.Members, *leader)

	rooms = append(rooms, r)

	leader.RoomID = len(rooms) - 1

	// TODO make sure that room name isn't taken
}

func newSecureRoom(roomName string, publicRoom bool, leader Client, password string) {
	r := Room{roomName, publicRoom, true, password, leader, make([]Client, 0)}

	rooms = append(rooms, r)
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
		if r.Members[i].Connection == c.Connection { // Testing for connection because it's something static for a client, that is also unique to it
			r.Members[i] = r.Members[len(r.Members)-1]
			r.Members = r.Members[:len(r.Members)-1]
			fmt.Println(r.Members)
		}
	}
}
