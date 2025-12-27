// Command rayserver is the dedicated game server.
package main

import (
	"fmt"
	"os"
)

// Version is set at build time
var Version = "dev"

func main() {
	fmt.Printf("Rayman Server v%s\n", Version)
	fmt.Println("Server starting...")

	// TODO: Parse flags (--port, --max-players, --map)
	// TODO: Load map
	// TODO: Initialize ECS world
	// TODO: Start network listener
	// TODO: Run tick loop

	os.Exit(0)
}
