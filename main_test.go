package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoomName(t *testing.T) {
	name := joinRoomName("/join room1")
	assert.Equal(t, "room1", name)
}

func TestInitRoom(t *testing.T) {
	initRoom()
	for _, r := range rooms {
		fmt.Println(r.name)
	}
}

func TestJoinRoom(t *testing.T) {
	initRoom()

	client := make(chan string)
	r := joinRoom("room1", client)
	assert.Equal(t, r.name, "room1")

	r = joinRoom("room_no_exist", client)
	assert.Equal(t, r.name, "default_room")
}
