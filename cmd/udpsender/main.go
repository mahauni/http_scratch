package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	serverAddr := "localhost:42069"
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resoloving UDP address: %s\n", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error dialing UDP: %s\n", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n", serverAddr)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading message: %s\n", err.Error())
			os.Exit(1)
		}

		_, err = conn.Write([]byte(msg))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending message: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Message sent: %s", msg)
	}
}
