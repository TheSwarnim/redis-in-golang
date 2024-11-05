// internal/commands/commands.go
package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// Handler represents a command handler function
type Handler func([]resp.Value, Store) resp.Value

// Command represents a Redis command
type Command struct {
	Name    string
	Handler Handler
}

// Store interface defines the methods required for command handlers
type Store interface {
	Get(key string) (resp.Value, bool)
	Set(key string, value resp.Value)
}

// Registry holds all available commands
type Registry struct {
	commands map[string]Handler
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	r := &Registry{
		commands: make(map[string]Handler),
	}
	r.registerCommands()
	return r
}

// Register adds a new command to the registry
func (r *Registry) Register(name string, handler Handler) {
	r.commands[name] = handler
}

// Get returns a command handler by name
func (r *Registry) Get(name string) (Handler, bool) {
	handler, ok := r.commands[name]
	return handler, ok
}

// registerCommands registers all available commands
func (r *Registry) registerCommands() {
	r.Register("PING", Ping)
	r.Register("ECHO", Echo)
	r.Register("SET", Set)
	r.Register("GET", Get)
}

// Command Handlers
func Ping(args []resp.Value, _ Store) resp.Value {
	if len(args) == 0 {
		return resp.SimpleStringVal("PONG")
	}

	// If there are arguments, return the first one
	return resp.BulkStringVal(args[0].Str)
}

func Echo(args []resp.Value, _ Store) resp.Value {
	if len(args) != 1 {
		return resp.ErrorVal("Error: ECHO command requires exactly 1 argument")
	}

	return resp.BulkStringVal(args[0].Str)
}

func Set(args []resp.Value, store Store) resp.Value {
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

func Get(args []resp.Value, store Store) resp.Value {
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
