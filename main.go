package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type room struct {
	name     string
	clients  map[client]bool
	entering chan client //入室を監視するチャネル
	leaving  chan client //退室を監視するチャネル
	messages chan string //ブロードキャスト用のメッセージを保持
}

var defaultRoom = room{
	name:     "default_room",
	clients:  make(map[client]bool),
	entering: make(chan client),
	leaving:  make(chan client),
	messages: make(chan string),
}

type client chan<- string

var (
	rooms = make([]room, 0)
)

func main() {
	initRoom()
	doMain()
}

func doMain() {
	listner, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range rooms {
		rr := r
		go rr.broadcaster()
	}

	for {
		conn, err := listner.Accept()
		if err != nil {
			fmt.Print(err)
			continue
		}

		go connHandler(conn)
	}
}

func (r *room) broadcaster() {
	for {
		select {
		case msg := <-r.messages: //メッセージを入室しているクライアントの送信用チャネルに送る
			for c := range r.clients {
				c <- msg
			}
		case c := <-r.entering:
			r.clients[c] = true
		case c := <-r.leaving:
			delete(r.clients, c)
		}
	}
}

func connHandler(conn net.Conn) {
	var r *room
	sender := make(chan string)
	go clientWriter(conn, sender)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg := input.Text()
		if strings.HasPrefix(msg, "/join") {
			if r != nil {
				sender <- "Please leave room before join"
				continue
			}
			n := joinRoomName(msg)
			r = joinRoom(n, sender)
			sender <- fmt.Sprintf("[%s] You are %s", r.name, conn.RemoteAddr().String())
			r.messages <- fmt.Sprintf("[%s] %s has arrived", r.name, conn.RemoteAddr().String())
			r.entering <- sender
			continue
		}

		if strings.HasPrefix(msg, "/leave") {
			r.leaving <- sender
			r.messages <- fmt.Sprintf("[%s] %s has left", r.name, conn.RemoteAddr().String())
			r = nil
			continue
		}

		if strings.HasPrefix(msg, "/create") {
			n := createRoomName(msg)
			r := createRoom(n)
			go r.broadcaster()
			continue
		}

		if strings.HasPrefix(msg, "/members") {
			continue
		}

		if r != nil {
			r.messages <- fmt.Sprintf("[%s] %s : %s", r.name, conn.RemoteAddr().String(), input.Text())
		} else {
			sender <- "Please join room"
		}
	}

	if r != nil {
		delete(r.clients, sender)
		r.messages <- fmt.Sprintf("[%s] %s has left", r.name, conn.RemoteAddr().String())
	}
	close(sender)
	conn.Close()
}

func clientWriter(conn net.Conn, sender <-chan string) {
	for msg := range sender {
		fmt.Fprintln(conn, msg)
	}
}

func initRoom() {
	r1 := room{
		name:     "room1",
		clients:  make(map[client]bool),
		entering: make(chan client),
		leaving:  make(chan client),
		messages: make(chan string),
	}
	r2 := room{
		name:     "room2",
		clients:  make(map[client]bool),
		entering: make(chan client),
		leaving:  make(chan client),
		messages: make(chan string),
	}
	rooms = append(rooms, defaultRoom)
	rooms = append(rooms, r2)
	rooms = append(rooms, r1)
}

// TODO room function
func joinRoom(name string, c client) *room {
	for _, r := range rooms {
		if r.name == name {
			r.clients[c] = true
			return &r
		}
	}
	return &defaultRoom
}

func createRoom(name string) *room {
	r := room{
		name:     name,
		clients:  make(map[client]bool),
		entering: make(chan client),
		leaving:  make(chan client),
		messages: make(chan string),
	}
	rooms = append(rooms, r)
	return &r
}

func joinRoomName(input string) string {
	return strings.TrimPrefix(input, "/join ")
}

func createRoomName(input string) string {
	s := strings.TrimPrefix(input, "/create ")
	return strings.Trim(s, " ")
}
