package main

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// Store represents a simple in-memory key-value store
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, exists := s.data[key]
	return val, exists
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func main() {
	fmt.Println("Server starting...")
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Println("Listening on port 6379...")

	store := NewStore()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("New connection from %s\n", conn.RemoteAddr())
		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store *Store) {
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

		response := handleCommand(value.Array, store)
		err = writer.Write(response)
		if err != nil {
			fmt.Printf("Error writing to %s: %s\n", conn.RemoteAddr(), err)
			return
		}
	}
}

func handleCommand(commands []resp.Value, store *Store) resp.Value {
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
	case "SET":
		return handleSet(args, store)
	case "GET":
		return handleGet(args, store)
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

func handleSet(args []resp.Value, store *Store) resp.Value {
	if len(args) != 2 {
		return resp.ErrorVal("Error: SET command requires exactly 2 arguments")
	}

	key := args[0].Str
	value := args[1].Str
	store.Set(key, value)

	return resp.SimpleStringVal("OK")
}

func handleGet(args []resp.Value, store *Store) resp.Value {
	if len(args) != 1 {
		return resp.ErrorVal("Error: GET command requires exactly 1 argument")
	}

	key := args[0].Str
	value, exists := store.Get(key)
	if !exists {
		return resp.NullBulkStringVal()
	}

	return resp.BulkStringVal(value)
}
