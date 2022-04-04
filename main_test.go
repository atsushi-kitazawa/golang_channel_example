package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoomName(t *testing.T) {
	name := trimCmd("/join room1", "/join")
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

	client := &client{
		name:   "user1",
		sender: make(chan string),
	}
	r := joinRoom("room1", client)
	assert.Equal(t, r.name, "room1")

	r = joinRoom("room_no_exist", client)
	assert.Equal(t, r.name, "default_room")
}

func TestCheckHasJoined(t *testing.T) {
	initRoom()

	client := &client{
		name:   "user1",
		sender: make(chan string),
	}
	joinRoom("room1", client)

	ret := checkAlreadyJoined("room1", client)
	assert.Equal(t, true, ret)
	ret = checkAlreadyJoined("room2", client)
	assert.Equal(t, false, ret)

	client.name = "user11"
	ret = checkAlreadyJoined("room1", client)
	assert.Equal(t, true, ret)
	ret = checkAlreadyJoined("room2", client)
	assert.Equal(t, false, ret)
}
