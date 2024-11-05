package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Server struct {
	addr     string
	store    commands.Store
	registry *commands.Registry
}

func NewServer(addr string, store commands.Store) *Server {
	return &Server{
		addr:     addr,
		store:    store,
		registry: commands.NewRegistry(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	defer ln.Close()

	fmt.Printf("Listening on %s...\n", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("New connection from %s", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
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
			fmt.Printf("Expected array but got %v\n", value.Type)
			continue
		}

		response := s.handleCommand(value.Array)
		err = writer.Write(response)
		if err != nil {
			fmt.Printf("Error writing to %s: %s\n", conn.RemoteAddr(), err)
			return
		}
	}
}

func (s *Server) handleCommand(commands []resp.Value) resp.Value {
	if len(commands) == 0 {
		return resp.ErrorVal("ERR no command provided")
	}

	// the first command is the command name
	commandName := strings.ToUpper(commands[0].Str)
	handler, exists := s.registry.Get(commandName)
	if !exists {
		return resp.ErrorVal(fmt.Sprintf("ERR unknown command '%s'", commandName))
	}

	return handler(commands[1:], s.store)
}
