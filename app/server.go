package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
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

	parser := resp.NewParser(conn)
	writer := resp.NewWriter(conn)

	for {
		value, err := parser.Parse()
		if err != nil {
			fmt.Printf("Error parsing from %s: %s\n", conn.RemoteAddr(), err)
			return
		}

		if value.Type != resp.Array {
			fmt.Printf("Expected array but got %T\n", value.Type)
			continue
		}

		response := handleCommand(value.Array)
		err = writer.Write(response)
		if err != nil {
			fmt.Printf("Error writing to %s: %s\n", conn.RemoteAddr(), err)
			return
		}
	}
}

func handleCommand(commands []resp.Value) resp.Value {
	if len(commands) == 0 {
		return resp.ErrorVal("Error: no command provided")
	}

	// the first command is the command name
	command := strings.ToUpper(commands[0].Str)
	args := commands[1:]

	switch command {
	case "PING":
		return handlePing(args)
	case "ECHO":
		return handleEcho(args)
	default:
		return resp.ErrorVal(fmt.Sprintf("Error: unknown command '%s'", command))
	}
}

func handlePing(args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.SimpleStringVal("PONG")
	}

	// If there are arguments, return the first one
	return resp.BulkStringVal(args[0].Str)
}

func handleEcho(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.ErrorVal("Error: ECHO command requires exactly 1 argument")
	}

	return resp.BulkStringVal(args[0].Str)
}
