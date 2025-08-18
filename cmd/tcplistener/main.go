package main

import (
	"fmt"
	"log"
	"net"

	"github.com/mahauni/http_scratch/internal/request"
)

var addr = ":42069"

func main() {
	n, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer n.Close()

	for {
		conn, err := n.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		r.Headers.Foreach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})
		fmt.Printf("Body:\n")
		fmt.Printf("%s:\n", r.Body)

		fmt.Println("Connection to", conn.RemoteAddr(), "closed")
	}
}
