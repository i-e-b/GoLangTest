package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// From https://varshneyabhi.wordpress.com/2014/12/23/simple-udp-clientserver-in-golang/

func CheckError(err error) {
	if err  != nil {
		fmt.Println("Error: " , err)
	}
}

func main() {
	ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:9002")
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	defer func(Conn *net.UDPConn) {_ = Conn.Close() }(conn)

	inBuf := make([]byte, 1024)

	i := 0
	for {
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		_,err := conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}

		_ = conn.SetReadDeadline(deadline(time.Millisecond * 250))
		n, _, err := conn.ReadFromUDP(inBuf)
		if err != nil {
			fmt.Println(msg, err)
		} else {
			fmt.Println(string(inBuf[0:n]))
		}

		time.Sleep(time.Second * 1)
	}
}

func deadline(d time.Duration) time.Time {
	w := time.Now()
	return w.Add(d)
}