package main

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
    store := store.NewStore()
    server := server.NewServer(":6379", store)

    log.Println("Starting server...")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
