//go:build gio

package render

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"

	"github.com/andersfylling/rayman-slides/internal/game"
)

const (
	// Tile size in pixels for world-to-screen conversion
	GioTilePixels = 32
)

// GioRenderer renders using Gio with sprite atlas support.
type GioRenderer struct {
	tileSize int
	tileMap  [][]rune
	world    *game.World
	camera   Camera
	hudText  string
	theme    *material.Theme

	// Sprite atlas
	atlas    *Atlas
	atlasOp  paint.ImageOp
	useAtlas bool
}

// NewGioRenderer creates a new Gio renderer (without sprites).
func NewGioRenderer() *GioRenderer {
	return &GioRenderer{
		tileSize: GioTilePixels,
		theme:    material.NewTheme(),
		useAtlas: false,
	}
}

// LoadSprites loads the sprite atlas from a filesystem.
func (r *GioRenderer) LoadSprites(fsys fs.FS) error {
	atlas, err := LoadAtlas(fsys)
	if err != nil {
		return err
	}
	r.atlas = atlas
	r.atlasOp = paint.NewImageOp(atlas.Image)
	r.useAtlas = true
	fmt.Println("Sprite atlas loaded successfully")
	return nil
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

			// Try to draw from atlas first
			if r.useAtlas {
				var spriteID string
				switch tile {
				case '#':
					spriteID = "tile_ground"
				case '=':
					spriteID = "tile_wood"
				case '~':
					spriteID = "tile_water"
				case '^':
					spriteID = "tile_spikes"
				default:
					spriteID = "tile_brick"
				}

				if region, ok := r.atlas.GetRegion(spriteID); ok {
					r.drawSprite(ops, int(px), int(py), r.tileSize, r.tileSize, region)
					continue
				}
			}

			// Fallback to colored rectangles
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

	// Try sprite atlas first
	if r.useAtlas {
		if region, ok := r.atlas.GetRegion(entity.SpriteID); ok {
			// Calculate draw position using anchor
			drawX := int(px) - region.AnchorX
			drawY := int(py) - region.AnchorY

			r.drawSprite(ops, drawX, drawY, region.W, region.H, region)
			return
		}
	}

	// Fallback to colored rectangles
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
	drawX := int(px) - w/2
	drawY := int(py) - h

	drawRect(ops, drawX, drawY, w, h, entityColor)
}

// drawSprite draws a sprite from the atlas
func (r *GioRenderer) drawSprite(ops *op.Ops, x, y, w, h int, region SpriteRegion) {
	// Create transformation stack
	defer op.Offset(image.Pt(x, y)).Push(ops).Pop()

	// Clip to the target draw area
	defer clip.Rect{Max: image.Pt(w, h)}.Push(ops).Pop()

	// Calculate scale to fit the target size
	scaleX := float32(w) / float32(region.W)
	scaleY := float32(h) / float32(region.H)

	// Apply scale
	if region.FlipX {
		// Flip horizontally: translate to right edge, scale negatively
		op.Affine(f32.Affine2D{}.
			Offset(f32.Pt(float32(w), 0)).
			Scale(f32.Pt(0, 0), f32.Pt(-scaleX, scaleY))).Add(ops)
	} else {
		op.Affine(f32.Affine2D{}.
			Scale(f32.Pt(0, 0), f32.Pt(scaleX, scaleY))).Add(ops)
	}

	// Offset to sprite position in atlas (negative to select the right region)
	op.Affine(f32.Affine2D{}.Offset(f32.Pt(float32(-region.X), float32(-region.Y)))).Add(ops)

	// Draw the atlas image
	r.atlasOp.Add(ops)
	paint.PaintOp{}.Add(ops)
}

func (r *GioRenderer) drawHUD(gtx layout.Context) {
	label := material.Body1(r.theme, r.hudText)
	label.Color = color.NRGBA{255, 255, 255, 255}
	label.Alignment = text.Start
	label.Layout(gtx)
}

// drawRect draws a filled rectangle (fallback when no atlas)
func drawRect(ops *op.Ops, x, y, w, h int, c color.NRGBA) {
	defer clip.Rect{Min: image.Pt(x, y), Max: image.Pt(x+w, y+h)}.Push(ops).Pop()
	paint.Fill(ops, c)
}
