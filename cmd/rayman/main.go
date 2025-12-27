// Command rayman is the game client.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/andersfylling/rayman-slides/internal/game"
	"github.com/andersfylling/rayman-slides/internal/protocol"
	"github.com/andersfylling/rayman-slides/internal/render"
)

// Version is set at build time
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize renderer (could be swapped for SDL, Vulkan, etc.)
	renderer := render.NewTcellRenderer()
	if err := renderer.Init(); err != nil {
		return fmt.Errorf("failed to initialize renderer: %w", err)
	}
	defer renderer.Close()

	// Create game world
	world := game.NewWorld()

	// Load demo level
	tileMap := game.DemoLevel()
	world.SetTileMap(tileMap)

	// Set tile map for rendering
	tiles := game.RenderTileMap(tileMap)
	renderer.SetTileMap(tiles)

	// Spawn player at starting position
	world.SpawnPlayer(1, "Player", 5, 10)

	// Spawn some enemies
	world.SpawnEnemy("slime", 15, 10)
	world.SpawnEnemy("slime", 28, 14)

	// Game loop
	ticker := time.NewTicker(time.Second / 60) // 60 ticks per second
	defer ticker.Stop()

	running := true
	var currentIntents protocol.Intent

	for running {
		// Process all pending input events
		for {
			event, ok := renderer.PollInput()
			if !ok {
				break
			}

			switch event.Type {
			case render.InputQuit:
				running = false
			case render.InputKey:
				currentIntents |= event.Intent
			case render.InputResize:
				// Handled by renderer
			}
		}

		// Wait for next tick
		select {
		case <-ticker.C:
			// Update game
			world.SetPlayerIntent(1, currentIntents)
			world.Update()

			// Clear intents after processing
			currentIntents = protocol.IntentNone

			// Get camera position (follow player)
			playerX, playerY, _ := world.GetPlayerPosition()
			camera := render.Camera{
				X: playerX,
				Y: playerY,
			}

			// Render frame
			renderer.BeginFrame()
			renderer.RenderWorld(world, camera)
			renderer.DrawHUD(fmt.Sprintf(" Tick: %d | WASD/Arrows: Move | Q/Esc: Quit ", world.Tick))
			renderer.EndFrame()
		}

		if !running {
			break
		}
	}

	return nil
}
