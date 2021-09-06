package main

import (
	"fmt"
	"net"
	"os"
)

// From https://varshneyabhi.wordpress.com/2014/12/23/simple-udp-clientserver-in-golang/

func CheckError(err error) {
	if err  != nil {
		fmt.Println("Error: " , err)
		os.Exit(0)
	}
}

func main() {
	// Prepare an address at any address at port 9002
	ServerAddr,err := net.ResolveUDPAddr("udp",":9002")
	CheckError(err)

	// Now listen at selected port
	con, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer func(ServerConn *net.UDPConn) {_ = ServerConn.Close() }(con)

	buf := make([]byte, 1024)

	for {
		n,addr,err := con.ReadFromUDP(buf)
		fmt.Println("Received ",string(buf[0:n]), " from ",addr)

		_, err = con.WriteToUDP([]byte("OK"), addr)
		if err != nil {
			fmt.Println(err)
		}

		if err != nil {
			fmt.Println("Error: ",err)
		}
	}
}