package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	fmt.Println("Starting server on :9001")

	// listen on port
	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		panic(err)
	}

	fmt.Print("Ready. Connect with\r\n    o 127.0.0.1 9001\r\non Windows telnet app")

	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}
	defer func(conn net.Conn) { _ = conn.Close() }(conn)

	clearConsole(conn)
	writeConsole(conn, "Welcome to UselessNet. Type 'die' to end session.")

	sessionLoop(conn)
}

func clearConsole(conn net.Conn) {
	_, err := conn.Write([]byte("\u001B[2J")) // VT clear
	if err != nil {
		panic(err)
	}
}
func writeConsole(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}

func sessionLoop(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			panic(err)
		}

		switch {
		case is(message, "die"):
			writeConsole(conn, "bye!\r\n")
			return

		case is(message, "pink"):
			// Working: 7 - invert, 30 - 37 colors
			// Not working: 90-97 colors, 48+ colors
			writeConsole(conn, "\u001B[35m;You got it.\r\n")

		case is(message, "white"):
			writeConsole(conn, "\u001B[37mPlain it is\r\n")

		case is(message, "go west"):
			writeConsole(conn, "It is dark, you may be eaten by a grue.\r\n")

		default:
			writeConsole(conn, "Sure, tell me more.\r\n")
		}

		fmt.Printf("New msg - %v", message)
	}
}

func is(in, msg string) bool{
	return in == msg || in == msg+"\n" || in == msg+"\r\n"
}
