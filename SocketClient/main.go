package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

/*

You should run the socket server before triggering this.

*/

func main() {
	// net.DNSError (?)
	ipAddr, err := net.LookupAddr("google.com")
	if err != nil {
		fmt.Printf("Lookup failed: %v", ipAddr)
	}
	fmt.Printf("Found: %v", ipAddr)


	conn, err := net.Dial("tcp", "127.0.0.1:9001")
	if err != nil { panic(err) }
	defer func(conn net.Conn) {_ = conn.Close() }(conn)

	// check for a welcome message
	err = conn.SetReadDeadline(deadline(500 * time.Millisecond))
	if err != nil {panic(err)}
	reply, _ := bufio.NewReader(conn).ReadString('\n')
	if reply != "" {
		fmt.Print("    |", reply)
	}
	_ = conn.SetDeadline(time.Time{})

	clientLoop(conn)
}

func clientLoop(conn net.Conn) {
	// read from stdin, and relay to the server
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(":> ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		_, err = fmt.Fprintf(conn, cmd+"\n") // send to server
		if err != nil {
			fmt.Printf("Remote server disconnected. Ending. (%v)", err)
			return
		}

		err = conn.SetReadDeadline(deadline(500 * time.Millisecond)) // if the server doesn't reply, don't wait forever
		if err != nil {
			panic(err)
		}

		reply, err := bufio.NewReader(conn).ReadString('\n') // blocking read from server
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				fmt.Printf("Server was quiet (%v)", err)
				continue
			}
			fmt.Printf("Remote server disconnected. Ending. (%v) / %T", err, err)
			return
		}

		fmt.Print("    |", reply)
	}
}

func deadline(d time.Duration) time.Time {
	w := time.Now()
	return w.Add(d)
}
