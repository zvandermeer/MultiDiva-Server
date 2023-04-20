package main

type Room struct {
	RoomTitle         string
	Publicity         bool
	PasswordProtected bool
	Password          string // TODO make this more secure
	Leader            Client
	Members           []Client
}

func newRoom(roomName string, publicity bool, leader Client) {
	r := Room{roomName, publicity, false, "", leader, make([]Client, 0)}

	rooms = append(rooms, r)
}

func newSecureRoom(roomName string, publicity bool, leader Client, password string) {
	r := Room{roomName, publicity, true, password, leader, make([]Client, 0)}

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
		if r.Members[i] == c {
			r.Members[i] = r.Members[len(r.Members)-1]
			r.Members = r.Members[:len(r.Members)-1]
		}
	}
}
