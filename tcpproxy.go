package main

import (
	"flag"
	"io"
	"log"
	"net"
)

func proxy(conn net.Conn, raddr string) {
	client, err := net.Dial("tcp", raddr)

	if err != nil {
		log.Fatalf("Unable : %v", err)
	}

	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(client, conn)
	}()

	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()
}

func main() {
	var (
		listenAddr = flag.String("l", "localhost:8080", "Listen on this address")
		remoteAddr = flag.String("r", "", "Remote address")
	)

	flag.Parse()

	ln, err := net.Listen("tcp", *listenAddr)

	if err != nil {
		log.Fatalf("Unable to start listener on %v: %v", listenAddr, err)
	}

	log.Printf("Started proxy on %s for %s", *listenAddr, *remoteAddr)

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Fatalf("Connection error: %v", err)
		}

		go proxy(conn, *remoteAddr)
	}
}
