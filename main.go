package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	doMain()
}

func doMain() {
	listner, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listner.Accept()
		if err != nil {
			panic(err)
		}

		go echoHandler(conn)
	}
}

func echoHandler(conn net.Conn) {
	for {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("connection reset by peer.")
			return
		}
		conn.Write([]byte(conn.RemoteAddr().String() + " > " + data))
	}
}
