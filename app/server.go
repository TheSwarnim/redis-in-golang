package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	fmt.Println("Server starting...")
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening on localhost:6379")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("New connection from %s\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		fmt.Printf("Connection closed from %s\n", conn.RemoteAddr())
	}()

	fmt.Printf("Handling connection from %s\n", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Printf("Received from %s: %s\n", conn.RemoteAddr(), text)

		if strings.TrimSpace(text) == "PING" {
			_, err := conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Printf("Error writing to %s: %s\n", conn.RemoteAddr(), err)
				return
			}
			fmt.Printf("Sent PONG to %s\n", conn.RemoteAddr())
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Scanner error for %s: %s\n", conn.RemoteAddr(), err)
	}
}
