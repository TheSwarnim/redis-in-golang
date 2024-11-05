package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// Store represents a simple in-memory key-value store
type Store struct {
	mu   sync.RWMutex
	data map[string]resp.Value
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]resp.Value),
	}
}

func (s *Store) Get(key string) (resp.Value, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	if !exists {
		return resp.Value{}, false
	}

	if val.IsExpired() {
		// Delete expired key under write lock
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		s.mu.RLock()
		return resp.Value{}, false
	}

	return val, true
}

func (s *Store) Set(key string, value resp.Value) {
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
	if len(args) < 2 {
		return resp.ErrorVal("Error: SET command requires at least 2 arguments")
	}

	key := args[0].Str
	value := args[1].Str
	var val resp.Value

	// Parse optional parameters
	for i := 2; i < len(args); i++ {
		switch strings.ToUpper(args[i].Str) {
		case "PX":
			if i+1 >= len(args) {
				return resp.ErrorVal("Error: PX option requires a value")
			}
			milliseconds, err := strconv.Atoi(args[i+1].Str)
			if err != nil {
				return resp.ErrorVal("Error: value is not an integer or out of range")
			}
			val = resp.BulkStringValWithExpiry(value, time.Duration(milliseconds)*time.Millisecond)
			i++
		default:
			return resp.ErrorVal("Error: invalid option")
		}
	}

	if val.Type == 0 { // No expiry set because default value is 0 for Type(byte)
		val = resp.BulkStringVal(value)
	}

	store.Set(key, val)
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

	return value
}
