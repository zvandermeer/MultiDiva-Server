package main

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

type RankingData struct {
	ClientUUID string
	score      int // Make goroutine every 500ms update rankings
}

type Room struct {
	RoomUUID          string
	RoomTitle         string
	PasswordProtected bool
	Password          string // TODO make this more secure
	Leader            *Client
	Members           []*Client
	Ranking           []RankingData
	GameRunning       bool
}

func newRoom(roomName string, leader *Client) *Room {
	r := Room{}

	r.RoomUUID = uuid.NewString()
	r.RoomTitle = roomName
	r.PasswordProtected = false
	r.Leader = leader

	r.Members = append(r.Members, leader)

	return &r
}

func newProtectedRoom(roomName string, leader *Client, password string) *Room {
	r := Room{}

	r.RoomUUID = uuid.NewString()
	r.RoomTitle = roomName
	r.PasswordProtected = true
	r.Password = password
	r.Leader = leader

	r.Members = append(r.Members, leader)

	return &r
}

func (r Room) sendToOthersInRoom(message []byte, c Client) {
	for _, Member := range r.Members {
		if Member.ClientUUID != c.ClientUUID {
			Member.OutgoingMessageBuffer <- message
		}
	}
}

func (r Room) sendToAllInRoom(message []byte) {
	for _, Member := range r.Members {
		Member.OutgoingMessageBuffer <- message
	}
}

func (r *Room) removeClient(c Client) {
	for i := range r.Members {
		if r.Members[i].Connection == c.Connection { // Testing for connection because it's something static for a client, that is also unique to it
			r.Members[i] = r.Members[len(r.Members)-1]
			r.Members = r.Members[:len(r.Members)-1]
			break
		}
	}
}

func (r *Room) updateRankings() {
	for {
		if(!r.GameRunning) {
			break
		}

		time.Sleep(100 * time.Millisecond)

		sort.Slice(r.Ranking, func(i, j int) bool {
			return r.Ranking[i].score > r.Ranking[j].score
		  })
	}
}
