package main

import (
	"bufio"
	"errors"
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

type client struct {
	name   string
	sender chan<- string
}

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
				c.sender <- msg
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
	c := client{
		name:   conn.RemoteAddr().String(),
		sender: sender,
	}
	go clientWriter(conn, sender)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg := input.Text()
		if strings.HasPrefix(msg, "/join") {
			n := trimCmd(msg, "/join")
			if checkAlreadyJoined(n, c) {
				sender <- fmt.Sprintf("You are already joined %s", n)
				continue
			}
			r = joinRoom(n, c)
			sender <- fmt.Sprintf("[%s] You are %s", r.name, c.name)
			r.messages <- fmt.Sprintf("[%s] %s has arrived", r.name, c.name)
			r.entering <- c
			continue
		}

		if strings.HasPrefix(msg, "/leave") {
			r.leaving <- c
			r.messages <- fmt.Sprintf("[%s] %s has left", r.name, c.name)
			r = nil
			continue
		}

		if strings.HasPrefix(msg, "/create") {
			n := trimCmd(msg, "/create")
			r := createRoom(n)
			go r.broadcaster()
			continue
		}

		if strings.HasPrefix(msg, "/members") {
			if r == nil {
				sender <- "Please join room"
				continue
			}
			var members string
			for c := range r.clients {
				members += c.name + ","
			}
			sender <- members
			continue
		}

		if strings.HasPrefix(msg, "/switch") {
			n := trimCmd(msg, "/switch")
			if !checkAlreadyJoined(n, c) {
				sender <- fmt.Sprintf("Please join room %s", n)
				continue
			}
			r, _ = getRoom(n)
			sender <- fmt.Sprintf("switch %s", n)
			continue
		}

		if strings.HasPrefix(msg, "/current") {
			sender <- r.name
			continue
		}

		if strings.HasPrefix(msg, "/login") {
			n := trimCmd(msg, "/login")
			c.name = n
			continue
		}

		if r != nil {
			r.messages <- fmt.Sprintf("[%s] %s : %s", r.name, c.name, input.Text())
			fmt.Println(r.clients)
		} else {
			sender <- "Please join room"
		}
	}

	if r != nil {
		delete(r.clients, c)
		r.messages <- fmt.Sprintf("[%s] %s has left", r.name, c.name)
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

func checkAlreadyJoined(name string, c client) bool {
	for _, r := range rooms {
		if r.name == name {
			for cc := range r.clients {
				if cc.name == c.name {
					return true
				}
			}
		} else {
			continue
		}
	}
	return false
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

func getRoom(name string) (*room, error) {
	for _, r := range rooms {
		if r.name == name {
			return &r, nil
		}
	}
	return nil, errors.New("not exist " + name)
}

func trimCmd(msg, cmd string) string {
	s := strings.TrimPrefix(msg, cmd)
	return strings.Trim(s, " ")
}
