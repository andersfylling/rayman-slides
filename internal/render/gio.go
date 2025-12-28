//go:build gio

package render

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"

	"github.com/andersfylling/rayman-slides/internal/game"
)

const (
	// Tile size in pixels
	GioTilePixels = 16
)

// GioRenderer is a pure display using Gio.
type GioRenderer struct {
	tileSize int
	tileMap  [][]rune
	world    *game.World
	camera   Camera
	hudText  string
	theme    *material.Theme
}

// NewGioRenderer creates a new Gio display.
func NewGioRenderer() *GioRenderer {
	return &GioRenderer{
		tileSize: GioTilePixels,
		theme:    material.NewTheme(),
	}
}

// SetTileMap sets the background tile map.
func (r *GioRenderer) SetTileMap(tiles [][]rune) {
	r.tileMap = tiles
}

// SetWorld sets the world to render.
func (r *GioRenderer) SetWorld(world *game.World) {
	r.world = world
}

// SetCamera sets the camera position.
func (r *GioRenderer) SetCamera(camera Camera) {
	r.camera = camera
}

// SetHUD sets the HUD text.
func (r *GioRenderer) SetHUD(text string) {
	r.hudText = text
}

// ViewportSize returns viewport in world units.
func (r *GioRenderer) ViewportSize(gtx layout.Context) (width, height float64) {
	return float64(gtx.Constraints.Max.X) / float64(r.tileSize),
		float64(gtx.Constraints.Max.Y) / float64(r.tileSize)
}

// Layout renders the game frame.
func (r *GioRenderer) Layout(gtx layout.Context) layout.Dimensions {
	// Clear background
	paint.Fill(gtx.Ops, color.NRGBA{20, 20, 40, 255})

	if r.world == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	// Calculate camera offset
	screenW := float64(gtx.Constraints.Max.X)
	screenH := float64(gtx.Constraints.Max.Y)
	cameraOffsetX := screenW/2 - r.camera.X*float64(r.tileSize)
	cameraOffsetY := screenH/2 - r.camera.Y*float64(r.tileSize)

	// Render tile map
	if r.tileMap != nil {
		r.drawTileMap(gtx.Ops, cameraOffsetX, cameraOffsetY, screenW, screenH)
	}

	// Render entities
	for _, entity := range r.world.GetRenderables() {
		r.drawEntity(gtx.Ops, entity, cameraOffsetX, cameraOffsetY)
	}

	// Draw HUD
	if r.hudText != "" {
		r.drawHUD(gtx)
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (r *GioRenderer) drawTileMap(ops *op.Ops, offsetX, offsetY, screenW, screenH float64) {
	ts := float64(r.tileSize)

	for y, row := range r.tileMap {
		for x, tile := range row {
			if tile == ' ' || tile == 0 {
				continue
			}

			px := float64(x)*ts + offsetX
			py := float64(y)*ts + offsetY

			// Cull off-screen tiles
			if px < -ts || px > screenW || py < -ts || py > screenH {
				continue
			}

			var tileColor color.NRGBA
			switch tile {
			case '#':
				tileColor = color.NRGBA{100, 80, 60, 255}
			case '=':
				tileColor = color.NRGBA{80, 80, 80, 255}
			case '~':
				tileColor = color.NRGBA{50, 100, 200, 255}
			default:
				tileColor = color.NRGBA{60, 60, 60, 255}
			}

			drawRect(ops, int(px), int(py), r.tileSize, r.tileSize, tileColor)
		}
	}
}

func (r *GioRenderer) drawEntity(ops *op.Ops, entity game.Renderable, offsetX, offsetY float64) {
	ts := float64(r.tileSize)
	px := entity.X*ts + offsetX
	py := entity.Y*ts + offsetY

	w, h := int(ts*0.8), int(ts*0.9)

	var entityColor color.NRGBA
	switch {
	case len(entity.SpriteID) >= 6 && entity.SpriteID[:6] == "player":
		entityColor = color.NRGBA{0, 200, 0, 255}
		if len(entity.SpriteID) > 13 && entity.SpriteID[7:13] == "charge" {
			entityColor = color.NRGBA{255, 200, 0, 255}
		}
		if len(entity.SpriteID) >= 12 && entity.SpriteID[7:12] == "punch" {
			entityColor = color.NRGBA{200, 255, 0, 255}
		}
	case entity.SpriteID == "fist_right" || entity.SpriteID == "fist_left":
		entityColor = color.NRGBA{255, 255, 0, 255}
		w, h = int(ts*0.4), int(ts*0.4)
	case entity.SpriteID == "slime":
		entityColor = color.NRGBA{0, 180, 0, 255}
	case entity.SpriteID == "bat":
		entityColor = color.NRGBA{150, 0, 150, 255}
	default:
		entityColor = color.NRGBA{255, 0, 0, 255}
	}

	// Center on position
	px -= float64(w) / 2
	py -= float64(h)

	drawRect(ops, int(px), int(py), w, h, entityColor)
}

func (r *GioRenderer) drawHUD(gtx layout.Context) {
	// Draw HUD text at top-left
	label := material.Body1(r.theme, r.hudText)
	label.Color = color.NRGBA{255, 255, 255, 255}
	label.Alignment = text.Start
	label.Layout(gtx)
}

// drawRect draws a filled rectangle.
func drawRect(ops *op.Ops, x, y, w, h int, c color.NRGBA) {
	defer clip.Rect{Min: image.Pt(x, y), Max: image.Pt(x+w, y+h)}.Push(ops).Pop()
	paint.Fill(ops, c)
}
