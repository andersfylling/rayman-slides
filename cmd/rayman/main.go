// Command rayman is the game client.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/andersfylling/rayman-slides/internal/game"
	"github.com/andersfylling/rayman-slides/internal/protocol"
	"github.com/andersfylling/rayman-slides/internal/render"
	"github.com/gdamore/tcell/v2"
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
	// Initialize renderer
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

	// Spawn player at starting position
	world.SpawnPlayer(1, "Player", 5, 10)

	// Spawn some enemies
	world.SpawnEnemy("slime", 15, 10)
	world.SpawnEnemy("slime", 28, 14)

	// Get tile map rendering
	tiles := game.RenderTileMap(tileMap)

	// Game loop
	screen := renderer.Screen()
	ticker := time.NewTicker(time.Second / 60) // 60 FPS
	defer ticker.Stop()

	running := true
	var currentIntents protocol.Intent

	// Input goroutine
	inputChan := make(chan tcell.Event, 10)
	go func() {
		for {
			ev := screen.PollEvent()
			if ev == nil {
				return
			}
			inputChan <- ev
		}
	}()

	for running {
		select {
		case ev := <-inputChan:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlC:
					running = false
				case tcell.KeyLeft:
					currentIntents |= protocol.IntentLeft
				case tcell.KeyRight:
					currentIntents |= protocol.IntentRight
				case tcell.KeyUp:
					currentIntents |= protocol.IntentJump
				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q', 'Q':
						running = false
					case 'a', 'A':
						currentIntents |= protocol.IntentLeft
					case 'd', 'D':
						currentIntents |= protocol.IntentRight
					case 'w', 'W', ' ':
						currentIntents |= protocol.IntentJump
					case 'j', 'J':
						currentIntents |= protocol.IntentAttack
					}
				}
			case *tcell.EventResize:
				screen.Sync()
			}

		case <-ticker.C:
			// Update game
			world.SetPlayerIntent(1, currentIntents)
			world.Update()

			// Clear intents after processing (simple approach for held keys)
			// In a real implementation, we'd track key up/down properly
			currentIntents = protocol.IntentNone

			// Render
			renderGame(renderer, world, tiles)
		}
	}

	return nil
}

func renderGame(r *render.TcellRenderer, world *game.World, tiles [][]rune) {
	r.Clear()

	screenW, screenH := r.Size()

	// Get player position for camera
	playerX, playerY, hasPlayer := world.GetPlayerPosition()
	if !hasPlayer {
		playerX, playerY = 20, 10
	}

	// Camera offset (center player on screen)
	cameraX := int(playerX) - screenW/2
	cameraY := int(playerY) - screenH/2

	// Clamp camera
	if cameraX < 0 {
		cameraX = 0
	}
	if cameraY < 0 {
		cameraY = 0
	}
	if len(tiles) > 0 && len(tiles[0]) > 0 {
		maxCamX := len(tiles[0]) - screenW
		maxCamY := len(tiles) - screenH
		if cameraX > maxCamX && maxCamX > 0 {
			cameraX = maxCamX
		}
		if cameraY > maxCamY && maxCamY > 0 {
			cameraY = maxCamY
		}
	}

	// Render tiles
	for y := 0; y < screenH && y+cameraY < len(tiles); y++ {
		for x := 0; x < screenW && x+cameraX < len(tiles[0]); x++ {
			tileY := y + cameraY
			tileX := x + cameraX
			if tileY >= 0 && tileY < len(tiles) && tileX >= 0 && tileX < len(tiles[0]) {
				ch := tiles[tileY][tileX]
				if ch != ' ' {
					r.SetCell(x, y, ch, render.ColorWhite, render.ColorBlack)
				}
			}
		}
	}

	// Render entities
	for _, e := range world.GetRenderables() {
		screenX := int(e.X) - cameraX
		screenY := int(e.Y) - cameraY

		if screenX >= 0 && screenX < screenW && screenY >= 0 && screenY < screenH {
			fg := render.Color{
				R: uint8((e.Color >> 16) & 0xFF),
				G: uint8((e.Color >> 8) & 0xFF),
				B: uint8(e.Color & 0xFF),
			}
			r.SetCell(screenX, screenY, e.Char, fg, render.ColorBlack)
		}
	}

	// Draw HUD
	hudY := screenH - 1
	hud := fmt.Sprintf(" Tick: %d | WASD/Arrows: Move | Q/Esc: Quit ", world.Tick)
	r.DrawString(0, hudY, hud, render.ColorYellow, render.Color{R: 40, G: 40, B: 40})

	r.Flush()
}
