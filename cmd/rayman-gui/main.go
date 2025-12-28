//go:build gio

// Command rayman-gui is the graphical game client using Gio.
package main

import (
	"fmt"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	"github.com/andersfylling/rayman-slides/internal/game"
	"github.com/andersfylling/rayman-slides/internal/input"
	"github.com/andersfylling/rayman-slides/internal/render"
)

type keyboardTag struct{}

func main() {
	go func() {
		if err := run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run() error {
	window := new(app.Window)
	window.Option(
		app.Title("Rayman Slides"),
		app.Size(unit.Dp(1280), unit.Dp(720)),
	)

	inputSystem := input.NewGioInput()
	renderer := render.NewGioRenderer()

	world := game.NewWorld()
	tileMap := game.DemoLevelForViewport(80, 45)
	world.SetTileMap(tileMap)
	world.SpawnPlayer(1, "Player", 5, 10)
	world.SpawnEnemy("slime", 15, 10)
	world.SpawnEnemy("slime", 28, 14)

	tiles := game.RenderTileMap(tileMap)
	renderer.SetTileMap(tiles)

	// For single player, we don't need the full client/server setup
	// Just track key state and apply directly to world
	keyState := input.NewKeyState()

	var ops op.Ops
	var tag keyboardTag
	var click gesture.Click
	hasFocus := false
	focusRequested := false

	// Track time for fixed timestep
	lastUpdate := time.Now()
	tickDuration := time.Second / 60

	for {
		e := window.Event()

		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Create a clickable area covering the whole window
			area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
			event.Op(gtx.Ops, &tag)
			click.Add(gtx.Ops)
			area.Pop()

			// Check for clicks to grab focus
			for {
				ev, ok := click.Update(gtx.Source)
				if !ok {
					break
				}
				if ev.Kind == gesture.KindClick {
					gtx.Execute(key.FocusCmd{Tag: &tag})
					focusRequested = true
				}
			}

			// Request focus on first frame
			if !focusRequested {
				gtx.Execute(key.FocusCmd{Tag: &tag})
				focusRequested = true
			}

			// Check for focus events
			for {
				ev, ok := gtx.Event(key.FocusFilter{Target: &tag})
				if !ok {
					break
				}
				if fe, ok := ev.(key.FocusEvent); ok {
					hasFocus = fe.Focus
				}
			}

			// Process key events
			for {
				ev, ok := gtx.Event(key.Filter{Focus: &tag, Name: ""})
				if !ok {
					break
				}
				if ke, ok := ev.(key.Event); ok {
					inputSystem.HandleKeyEvent(ke)
				}
			}

			// Fixed timestep game updates
			now := time.Now()
			for now.Sub(lastUpdate) >= tickDuration {
				// Process input events
				events := inputSystem.Poll()
				for _, ev := range events {
					switch ev.Type {
					case input.KeyDown:
						keyState.SetPressed(ev.Key, true)
					case input.KeyUp:
						keyState.SetPressed(ev.Key, false)
					}
				}

				// Check for quit
				if keyState.IsPressed(input.KeyQuit) {
					return nil
				}

				// Apply intents to world and update
				world.SetPlayerIntent(1, keyState.ToIntents())
				world.Update()
				lastUpdate = lastUpdate.Add(tickDuration)
			}

			// Render with clamped camera
			playerX, playerY, _ := world.GetPlayerPosition()

			// Calculate viewport size in world units
			tileSize := float64(render.GioTilePixels)
			viewportW := float64(gtx.Constraints.Max.X) / tileSize
			viewportH := float64(gtx.Constraints.Max.Y) / tileSize

			// Clamp camera to keep map edges at screen edges
			mapW := float64(tileMap.Width)
			mapH := float64(tileMap.Height)

			camX := playerX
			camY := playerY

			// Clamp horizontal
			minCamX := viewportW / 2
			maxCamX := mapW - viewportW/2
			if maxCamX < minCamX {
				camX = mapW / 2 // Map smaller than viewport, center it
			} else if camX < minCamX {
				camX = minCamX
			} else if camX > maxCamX {
				camX = maxCamX
			}

			// Clamp vertical
			minCamY := viewportH / 2
			maxCamY := mapH - viewportH/2
			if maxCamY < minCamY {
				camY = mapH / 2 // Map smaller than viewport, center it
			} else if camY < minCamY {
				camY = minCamY
			} else if camY > maxCamY {
				camY = maxCamY
			}

			renderer.SetCamera(render.Camera{X: camX, Y: camY})
			renderer.SetWorld(world)

			hint := "Click window to focus | "
			if hasFocus {
				hint = ""
			}
			renderer.SetHUD(fmt.Sprintf("%sTick: %d | WASD: Move | J: Attack | Q/Esc: Quit", hint, world.Tick))
			renderer.Layout(gtx)

			e.Frame(gtx.Ops)
			window.Invalidate()
		}
	}
}
