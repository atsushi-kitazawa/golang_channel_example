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
	clients  map[*client]bool
	entering chan *client //入室を監視するチャネル
	leaving  chan *client //退室を監視するチャネル
	messages chan string  //ブロードキャスト用のメッセージを保持
}

var defaultRoom = room{
	name:     "default_room",
	clients:  make(map[*client]bool),
	entering: make(chan *client),
	leaving:  make(chan *client),
	messages: make(chan string),
}

type client struct {
	name   string
	rooms  map[string]struct{}
	sender chan<- string
}

const (
	joinCmd      = "/join"
	leaveCmd     = "/leave"
	createCmd    = "/create"
	switchCmd    = "/switch"
	membersCmd   = "/members"
	roomsCmd     = "/rooms"
	joinRoomsCmd = "/joinRooms"
	currentCmd   = "/current"
	loginCmd     = "/login"
)

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
	// current room
	var r *room

	sender := make(chan string)
	c := &client{
		name:   conn.RemoteAddr().String(),
		rooms:  make(map[string]struct{}),
		sender: sender,
	}
	go clientWriter(conn, sender)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg := input.Text()

		if strings.HasPrefix(msg, joinRoomsCmd) {
			var r string
			for rr := range c.rooms {
				r += rr + ","
			}
			sender <- r
			continue
		}

		if strings.HasPrefix(msg, joinCmd) {
			n := trimCmd(msg, joinCmd)
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

		if strings.HasPrefix(msg, leaveCmd) {
			delete(c.rooms, r.name)
			r.leaving <- c
			r.messages <- fmt.Sprintf("[%s] %s has left", r.name, c.name)
			r = nil
			continue
		}

		if strings.HasPrefix(msg, createCmd) {
			n := trimCmd(msg, createCmd)
			r := createRoom(n)
			go r.broadcaster()
			continue
		}

		if strings.HasPrefix(msg, membersCmd) {
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

		if strings.HasPrefix(msg, roomsCmd) {
			var r string
			for _, rr := range rooms {
				r += rr.name + ","
			}
			sender <- r
			continue
		}

		if strings.HasPrefix(msg, switchCmd) {
			n := trimCmd(msg, switchCmd)
			if !checkAlreadyJoined(n, c) {
				sender <- fmt.Sprintf("Please join room %s", n)
				continue
			}
			r, _ = getRoom(n)
			sender <- fmt.Sprintf("switch %s", n)
			continue
		}

		if strings.HasPrefix(msg, currentCmd) {
			sender <- r.name
			continue
		}

		if strings.HasPrefix(msg, loginCmd) {
			n := trimCmd(msg, loginCmd)
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
		clients:  make(map[*client]bool),
		entering: make(chan *client),
		leaving:  make(chan *client),
		messages: make(chan string),
	}
	r2 := room{
		name:     "room2",
		clients:  make(map[*client]bool),
		entering: make(chan *client),
		leaving:  make(chan *client),
		messages: make(chan string),
	}
	rooms = append(rooms, defaultRoom)
	rooms = append(rooms, r2)
	rooms = append(rooms, r1)
}

// TODO room function
func joinRoom(name string, c *client) *room {
	for _, r := range rooms {
		if r.name == name {
			r.clients[c] = true
			c.rooms[name] = struct{}{}
			return &r
		}
	}
	return &defaultRoom
}

func checkAlreadyJoined(name string, c *client) bool {
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
		clients:  make(map[*client]bool),
		entering: make(chan *client),
		leaving:  make(chan *client),
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
