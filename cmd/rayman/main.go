// Command rayman is the game client.
// Embeds server for local/singleplayer mode.
package main

import (
	"fmt"
	"os"
)

// Version is set at build time
var Version = "dev"

func main() {
	fmt.Printf("Rayman Terminal v%s\n", Version)
	fmt.Println("Client starting...")

	// TODO: Parse flags (--server, --connect, --ascii, --braille)
	// TODO: Detect terminal capabilities
	// TODO: Initialize renderer
	// TODO: Start embedded server or connect to remote
	// TODO: Run game loop

	os.Exit(0)
}
