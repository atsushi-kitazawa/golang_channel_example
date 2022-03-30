package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string

var (
	entering = make(chan client) //入室を関しするチャネル
	leaving  = make(chan client) //退室を監視するチャネル
	messages = make(chan string) //ブロードキャスト用のメッセージを保持
)

func main() {
	doMain()
}

func doMain() {
	listner, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()

	for {
		conn, err := listner.Accept()
		if err != nil {
			fmt.Print(err)
			continue
		}

		go connHandler(conn)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages: //メッセージを入室しているクライアントの送信用チャネルに送る
			for c := range clients {
				c <- msg
			}
		case c := <-entering:
			clients[c] = true
		case c := <-leaving:
			delete(clients, c)
			close(c)
		}
	}
}

func connHandler(conn net.Conn) {
	sender := make(chan string)
	go clientWriter(conn, sender)

	sender <- fmt.Sprintf("You are %s", conn.RemoteAddr().String())
	messages <- fmt.Sprintf("%s has arrived", conn.RemoteAddr().String())
	entering <- sender

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- fmt.Sprintf("%s : %s", conn.RemoteAddr().String(), input.Text())
	}

	leaving <- sender
	messages <- fmt.Sprintf("%s has left", conn.RemoteAddr().String())
	conn.Close()
}

func clientWriter(conn net.Conn, sender <-chan string) {
	for msg := range sender {
		fmt.Fprintln(conn, msg)
	}
}
