// Command lookup is the room code lookup service.
package main

import (
	"fmt"
	"os"
)

// Version is set at build time
var Version = "dev"

func main() {
	fmt.Printf("Room Lookup Service v%s\n", Version)
	fmt.Println("Starting HTTP server...")

	// TODO: Parse flags (--port, --ttl)
	// TODO: Initialize room store (in-memory or Redis)
	// TODO: Start HTTP server
	// POST /rooms - create room, returns code
	// GET /rooms/:code - lookup room, returns host IP:port
	// DELETE /rooms/:code - remove room

	os.Exit(0)
}
