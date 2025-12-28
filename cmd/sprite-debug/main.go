// Command sprite-debug generates a debug GIF showing all sprite regions.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sort"
)

type SpriteRegion struct {
	X       int  `json:"x"`
	Y       int  `json:"y"`
	W       int  `json:"w"`
	H       int  `json:"h"`
	AnchorX int  `json:"anchorX"`
	AnchorY int  `json:"anchorY"`
	FlipX   bool `json:"flipX,omitempty"`
}

type AtlasData struct {
	Image   string                  `json:"image"`
	Sprites map[string]SpriteRegion `json:"sprites"`
}

var data AtlasData
var atlasImg image.Image
var palette color.Palette

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load atlas metadata
	jsonData, err := os.ReadFile("assets/sprites/atlas.json")
	if err != nil {
		return fmt.Errorf("reading atlas.json: %w", err)
	}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("parsing atlas.json: %w", err)
	}

	// Load atlas image
	imgFile, err := os.Open("assets/sprites/atlas.jpg")
	if err != nil {
		return fmt.Errorf("opening atlas image: %w", err)
	}
	defer imgFile.Close()

	atlasImg, _, err = image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("decoding atlas image: %w", err)
	}

	// Build palette
	palette = buildPalette(atlasImg)

	// Sort sprite names for consistent ordering
	var names []string
	for name := range data.Sprites {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print sprite list
	fmt.Println("Sprites in atlas:")
	fmt.Println("─────────────────")
	for _, name := range names {
		region := data.Sprites[name]
		fmt.Printf("  %-30s  x:%-4d y:%-4d w:%-4d h:%-4d anchor:(%d,%d)\n",
			name, region.X, region.Y, region.W, region.H, region.AnchorX, region.AnchorY)
	}
	fmt.Printf("\nTotal: %d sprites\n\n", len(names))

	// Generate static grid GIF
	if err := generateStaticGrid(names); err != nil {
		return err
	}

	// Generate animated preview GIF
	if err := generateAnimatedPreview(); err != nil {
		return err
	}

	return nil
}

func generateStaticGrid(names []string) error {
	outGif := &gif.GIF{}

	// Calculate grid layout
	cols := 8
	cellW, cellH := 80, 80
	rows := (len(names) + cols - 1) / cols
	gridW := cols * cellW
	gridH := rows * cellH

	bounds := image.Rect(0, 0, gridW, gridH)
	frame := image.NewPaletted(bounds, palette)

	// Fill with dark background
	draw.Draw(frame, bounds, &image.Uniform{color.RGBA{30, 30, 40, 255}}, image.Point{}, draw.Src)

	// Draw each sprite in the grid
	for i, name := range names {
		region := data.Sprites[name]
		col := i % cols
		row := i / cols

		cellX := col * cellW
		cellY := row * cellH

		spriteRect := image.Rect(region.X, region.Y, region.X+region.W, region.Y+region.H)

		offsetX := cellX + (cellW-region.W)/2
		offsetY := cellY + (cellH-region.H)/2

		drawRegion(frame, atlasImg, spriteRect, image.Pt(offsetX, offsetY), region.FlipX)
		drawDot(frame, offsetX+region.AnchorX, offsetY+region.AnchorY, color.RGBA{255, 0, 0, 255})
		drawBorder(frame, offsetX, offsetY, region.W, region.H, color.RGBA{100, 100, 100, 255})
	}

	outGif.Image = append(outGif.Image, frame)
	outGif.Delay = append(outGif.Delay, 0)

	outFile, err := os.Create("sprites.debug.gif")
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer outFile.Close()

	if err := gif.EncodeAll(outFile, outGif); err != nil {
		return fmt.Errorf("encoding gif: %w", err)
	}

	fmt.Printf("Generated: sprites.debug.gif (%dx%d) - static grid\n", gridW, gridH)
	return nil
}

func generateAnimatedPreview() error {
	outGif := &gif.GIF{}

	// Layout: 4 columns showing different animations
	// Row 1: Player walk, Player on ground walking
	// Row 2: Player charge + punch, Slime
	// Row 3: Bat flying, Fist projectile
	// Row 4: Tiles demo

	width := 400
	height := 400
	groundY := 70 // Ground level within each cell

	// Animation frames (8 frames total for smooth loop)
	numFrames := 16
	frameDelay := 10 // 100ms per frame

	for f := 0; f < numFrames; f++ {
		bounds := image.Rect(0, 0, width, height)
		frame := image.NewPaletted(bounds, palette)

		// Fill with dark background
		draw.Draw(frame, bounds, &image.Uniform{color.RGBA{30, 30, 40, 255}}, image.Point{}, draw.Src)

		// Draw section labels
		drawLabel(frame, 10, 10, "Walk Cycle")
		drawLabel(frame, 210, 10, "Walk on Ground")
		drawLabel(frame, 10, 110, "Charge -> Punch")
		drawLabel(frame, 210, 110, "Slime")
		drawLabel(frame, 10, 210, "Bat Flying")
		drawLabel(frame, 210, 210, "Fist Projectile")
		drawLabel(frame, 10, 310, "Tiles")
		drawLabel(frame, 210, 310, "Collectibles")

		// === Walk Cycle (top-left) ===
		walkFrame := (f % 8) + 1
		walkSprite := fmt.Sprintf("player_walk_%d", walkFrame)
		drawSpriteAt(frame, walkSprite, 80, 30+groundY)

		// === Walk on Ground (top-right) ===
		// Draw ground tiles
		for x := 200; x < 400; x += 32 {
			drawSpriteAtSize(frame, "tile_ground", x, 30+groundY, 32, 32)
		}
		// Draw walking player moving
		walkX := 220 + (f * 8 % 160)
		drawSpriteAt(frame, walkSprite, walkX, 30+groundY)

		// === Charge -> Punch (middle-left) ===
		var chargeSprite string
		if f < 4 {
			chargeSprite = "player_charge_right_1"
		} else if f < 8 {
			chargeSprite = "player_charge_right_2"
		} else if f < 12 {
			chargeSprite = "player_charge_right_3"
		} else {
			chargeSprite = "player_punch_right"
		}
		drawSpriteAt(frame, chargeSprite, 80, 130+groundY)

		// === Slime (middle-right) ===
		slimeFrame := (f/4)%4 + 1
		slimeSprite := fmt.Sprintf("slime_%d", slimeFrame)
		// Draw ground
		for x := 200; x < 400; x += 32 {
			drawSpriteAtSize(frame, "tile_ground", x, 130+groundY, 32, 32)
		}
		// Slime bouncing
		bounceY := 0
		if f%8 < 4 {
			bounceY = (f % 4) * 3
		} else {
			bounceY = (4 - (f % 4)) * 3
		}
		drawSpriteAt(frame, slimeSprite, 300, 130+groundY-bounceY)

		// === Bat Flying (bottom-left row 1) ===
		batFrame := (f/2)%4 + 1
		batSprite := fmt.Sprintf("bat_%d", batFrame)
		// Bat flying in sine wave
		batX := 30 + (f * 6 % 140)
		batY := 230 + int(float64(f%16)*1.5)
		if f >= 8 {
			batY = 230 + int(float64(16-f%16)*1.5)
		}
		drawSpriteAt(frame, batSprite, batX, batY)

		// === Fist Projectile (bottom-right row 1) ===
		fistFrame := (f / 2) % 3
		fistSprites := []string{"fist_right_1", "fist_right_2", "fist_right_3"}
		fistSprite := fistSprites[fistFrame]
		fistX := 220 + (f * 10 % 160)
		drawSpriteAt(frame, fistSprite, fistX, 250)

		// === Tiles (bottom-left row 2) ===
		tiles := []string{"tile_ground", "tile_wood", "tile_brick", "tile_cloud"}
		for i, tile := range tiles {
			drawSpriteAtSize(frame, tile, 20+i*40, 330, 32, 32)
		}

		// === Collectibles (bottom-right row 2) ===
		tingFrame := (f / 4) % 3
		tingSprites := []string{"ting_1", "ting_2", "ting_3"}
		drawSpriteAt(frame, tingSprites[tingFrame], 240, 350)
		drawSpriteAt(frame, "health", 290, 350)
		drawSpriteAt(frame, "cage", 340, 350)

		outGif.Image = append(outGif.Image, frame)
		outGif.Delay = append(outGif.Delay, frameDelay)
	}

	outFile, err := os.Create("sprites.animated.gif")
	if err != nil {
		return fmt.Errorf("creating animated output file: %w", err)
	}
	defer outFile.Close()

	if err := gif.EncodeAll(outFile, outGif); err != nil {
		return fmt.Errorf("encoding animated gif: %w", err)
	}

	fmt.Printf("Generated: sprites.animated.gif (%dx%d) - animated preview\n", width, height)
	fmt.Println("Red dots show anchor points")
	return nil
}

func drawSpriteAt(frame *image.Paletted, spriteName string, x, y int) {
	region, ok := data.Sprites[spriteName]
	if !ok {
		return
	}

	// Draw at position, using anchor
	drawX := x - region.AnchorX
	drawY := y - region.AnchorY

	spriteRect := image.Rect(region.X, region.Y, region.X+region.W, region.Y+region.H)
	drawRegion(frame, atlasImg, spriteRect, image.Pt(drawX, drawY), region.FlipX)

	// Draw anchor point
	drawDot(frame, x, y, color.RGBA{255, 0, 0, 255})
}

func drawSpriteAtSize(frame *image.Paletted, spriteName string, x, y, w, h int) {
	region, ok := data.Sprites[spriteName]
	if !ok {
		return
	}

	// Scale sprite to fit w x h
	spriteRect := image.Rect(region.X, region.Y, region.X+region.W, region.Y+region.H)
	drawRegionScaled(frame, atlasImg, spriteRect, image.Pt(x, y), w, h)
}

func drawLabel(frame *image.Paletted, x, y int, text string) {
	// Simple pixel text - just draw dots for now as placeholder
	for i, c := range text {
		if c != ' ' {
			frame.Set(x+i*4, y, color.RGBA{150, 150, 150, 255})
			frame.Set(x+i*4+1, y, color.RGBA{150, 150, 150, 255})
			frame.Set(x+i*4, y+1, color.RGBA{150, 150, 150, 255})
		}
	}
}

// buildPalette creates a 256-color palette from the atlas image
func buildPalette(img image.Image) color.Palette {
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{30, 30, 40, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{100, 100, 100, 255},
		color.RGBA{150, 150, 150, 255},
	}

	bounds := img.Bounds()
	colorMap := make(map[color.RGBA]int)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 2 {
			r, g, b, a := img.At(x, y).RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			colorMap[c]++
		}
	}

	type colorCount struct {
		c     color.RGBA
		count int
	}
	var counts []colorCount
	for c, count := range colorMap {
		counts = append(counts, colorCount{c, count})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	for i := 0; i < len(counts) && len(palette) < 256; i++ {
		c := counts[i].c
		duplicate := false
		for _, existing := range palette {
			er, eg, eb, _ := existing.RGBA()
			if abs(int(c.R)-int(er>>8)) < 4 &&
				abs(int(c.G)-int(eg>>8)) < 4 &&
				abs(int(c.B)-int(eb>>8)) < 4 {
				duplicate = true
				break
			}
		}
		if !duplicate {
			palette = append(palette, c)
		}
	}

	return palette
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawRegion(dst *image.Paletted, src image.Image, srcRect image.Rectangle, dstPt image.Point, flipX bool) {
	w := srcRect.Dx()
	h := srcRect.Dy()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcX := srcRect.Min.X + x
			if flipX {
				srcX = srcRect.Max.X - 1 - x
			}
			c := src.At(srcX, srcRect.Min.Y+y)
			_, _, _, a := c.RGBA()
			if a > 128 {
				dst.Set(dstPt.X+x, dstPt.Y+y, c)
			}
		}
	}
}

func drawRegionScaled(dst *image.Paletted, src image.Image, srcRect image.Rectangle, dstPt image.Point, dstW, dstH int) {
	srcW := srcRect.Dx()
	srcH := srcRect.Dy()

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := srcRect.Min.X + x*srcW/dstW
			srcY := srcRect.Min.Y + y*srcH/dstH
			c := src.At(srcX, srcY)
			_, _, _, a := c.RGBA()
			if a > 128 {
				dst.Set(dstPt.X+x, dstPt.Y+y, c)
			}
		}
	}
}

func drawDot(dst *image.Paletted, x, y int, c color.Color) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx*dx+dy*dy <= 1 {
				dst.Set(x+dx, y+dy, c)
			}
		}
	}
}

func drawBorder(dst *image.Paletted, x, y, w, h int, c color.Color) {
	for dx := 0; dx < w; dx++ {
		dst.Set(x+dx, y, c)
		dst.Set(x+dx, y+h-1, c)
	}
	for dy := 0; dy < h; dy++ {
		dst.Set(x, y+dy, c)
		dst.Set(x+w-1, y+dy, c)
	}
}
