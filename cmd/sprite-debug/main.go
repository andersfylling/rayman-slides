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
	"image/png"
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

	// Generate atlas overlay PNG
	if err := generateAtlasOverlay(); err != nil {
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
		drawLabel(frame, 10, 10, "PLAYER WALK (8 FRAMES)")
		drawLabel(frame, 210, 10, "WALK ON TILES")
		drawLabel(frame, 10, 110, "CHARGE -> PUNCH ATTACK")
		drawLabel(frame, 210, 110, "SLIME ENEMY (4 FRAMES)")
		drawLabel(frame, 10, 210, "BAT ENEMY (4 FRAMES)")
		drawLabel(frame, 210, 210, "FIST PROJECTILE")
		drawLabel(frame, 10, 310, "TILE TYPES")
		drawLabel(frame, 210, 310, "COLLECTIBLE ITEMS")

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

// generateAtlasOverlay creates a PNG with sprite borders drawn on the atlas
func generateAtlasOverlay() error {
	bounds := atlasImg.Bounds()

	// Create RGBA copy of the atlas
	overlay := image.NewRGBA(bounds)
	draw.Draw(overlay, bounds, atlasImg, bounds.Min, draw.Src)

	// Define colors for different sprite types
	colorPlayer := color.RGBA{0, 255, 0, 255}   // Green
	colorEnemy := color.RGBA{255, 0, 0, 255}    // Red
	colorTile := color.RGBA{0, 150, 255, 255}   // Blue
	colorItem := color.RGBA{255, 255, 0, 255}   // Yellow
	colorEffect := color.RGBA{255, 0, 255, 255} // Magenta
	colorAnchor := color.RGBA{255, 255, 255, 255} // White

	// Draw borders and anchors for each sprite
	for name, region := range data.Sprites {
		// Choose color based on sprite type
		var borderColor color.RGBA
		switch {
		case len(name) >= 6 && name[:6] == "player":
			borderColor = colorPlayer
		case len(name) >= 4 && name[:4] == "fist":
			borderColor = colorPlayer
		case len(name) >= 5 && name[:5] == "slime":
			borderColor = colorEnemy
		case len(name) >= 3 && name[:3] == "bat":
			borderColor = colorEnemy
		case len(name) >= 4 && name[:4] == "tile":
			borderColor = colorTile
		case len(name) >= 4 && name[:4] == "ting":
			borderColor = colorItem
		case name == "health" || name == "cage":
			borderColor = colorItem
		case name == "dust_1" || name == "dust_2" || name == "impact" || name == "sparkle":
			borderColor = colorEffect
		default:
			borderColor = color.RGBA{200, 200, 200, 255} // Gray for unknown
		}

		// Draw border rectangle (2px thick for visibility)
		drawBorderRGBA(overlay, region.X, region.Y, region.W, region.H, borderColor, 2)

		// Draw anchor point as a cross
		anchorX := region.X + region.AnchorX
		anchorY := region.Y + region.AnchorY
		drawCrossRGBA(overlay, anchorX, anchorY, colorAnchor, 5)
	}

	// Save as PNG
	outFile, err := os.Create("sprites.debug.png")
	if err != nil {
		return fmt.Errorf("creating overlay output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, overlay); err != nil {
		return fmt.Errorf("encoding overlay png: %w", err)
	}

	fmt.Printf("Generated: sprites.debug.png (%dx%d) - atlas with region borders\n", bounds.Dx(), bounds.Dy())
	fmt.Println("  Green=player, Red=enemies, Blue=tiles, Yellow=items, Magenta=effects")
	fmt.Println("  White crosses mark anchor points")
	return nil
}

// drawBorderRGBA draws a rectangle border on an RGBA image
func drawBorderRGBA(img *image.RGBA, x, y, w, h int, c color.RGBA, thickness int) {
	// Top and bottom edges
	for dx := 0; dx < w; dx++ {
		for t := 0; t < thickness; t++ {
			img.Set(x+dx, y+t, c)
			img.Set(x+dx, y+h-1-t, c)
		}
	}
	// Left and right edges
	for dy := 0; dy < h; dy++ {
		for t := 0; t < thickness; t++ {
			img.Set(x+t, y+dy, c)
			img.Set(x+w-1-t, y+dy, c)
		}
	}
}

// drawCrossRGBA draws a cross marker on an RGBA image
func drawCrossRGBA(img *image.RGBA, x, y int, c color.RGBA, size int) {
	// Horizontal line
	for dx := -size; dx <= size; dx++ {
		img.Set(x+dx, y, c)
	}
	// Vertical line
	for dy := -size; dy <= size; dy++ {
		img.Set(x, y+dy, c)
	}
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

// 5x7 bitmap font for A-Z, 0-9, and common punctuation
var font5x7 = map[rune][7]uint8{
	'A': {0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'B': {0x1E, 0x11, 0x11, 0x1E, 0x11, 0x11, 0x1E},
	'C': {0x0E, 0x11, 0x10, 0x10, 0x10, 0x11, 0x0E},
	'D': {0x1E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x1E},
	'E': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x1F},
	'F': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x10},
	'G': {0x0E, 0x11, 0x10, 0x17, 0x11, 0x11, 0x0E},
	'H': {0x11, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'I': {0x0E, 0x04, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'J': {0x07, 0x02, 0x02, 0x02, 0x02, 0x12, 0x0C},
	'K': {0x11, 0x12, 0x14, 0x18, 0x14, 0x12, 0x11},
	'L': {0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x1F},
	'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
	'N': {0x11, 0x11, 0x19, 0x15, 0x13, 0x11, 0x11},
	'O': {0x0E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'P': {0x1E, 0x11, 0x11, 0x1E, 0x10, 0x10, 0x10},
	'Q': {0x0E, 0x11, 0x11, 0x11, 0x15, 0x12, 0x0D},
	'R': {0x1E, 0x11, 0x11, 0x1E, 0x14, 0x12, 0x11},
	'S': {0x0E, 0x11, 0x10, 0x0E, 0x01, 0x11, 0x0E},
	'T': {0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04},
	'U': {0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'V': {0x11, 0x11, 0x11, 0x11, 0x11, 0x0A, 0x04},
	'W': {0x11, 0x11, 0x11, 0x15, 0x15, 0x1B, 0x11},
	'X': {0x11, 0x11, 0x0A, 0x04, 0x0A, 0x11, 0x11},
	'Y': {0x11, 0x11, 0x0A, 0x04, 0x04, 0x04, 0x04},
	'Z': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x10, 0x1F},
	'0': {0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E},
	'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'2': {0x0E, 0x11, 0x01, 0x02, 0x04, 0x08, 0x1F},
	'3': {0x0E, 0x11, 0x01, 0x06, 0x01, 0x11, 0x0E},
	'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
	'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
	'6': {0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E},
	'7': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08},
	'8': {0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E},
	'9': {0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C},
	' ': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	'-': {0x00, 0x00, 0x00, 0x1F, 0x00, 0x00, 0x00},
	'>': {0x08, 0x04, 0x02, 0x01, 0x02, 0x04, 0x08},
	'<': {0x02, 0x04, 0x08, 0x10, 0x08, 0x04, 0x02},
	'+': {0x00, 0x04, 0x04, 0x1F, 0x04, 0x04, 0x00},
	'.': {0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x0C},
	',': {0x00, 0x00, 0x00, 0x00, 0x04, 0x04, 0x08},
	'!': {0x04, 0x04, 0x04, 0x04, 0x04, 0x00, 0x04},
	'?': {0x0E, 0x11, 0x01, 0x02, 0x04, 0x00, 0x04},
	'/': {0x01, 0x01, 0x02, 0x04, 0x08, 0x10, 0x10},
	'(': {0x02, 0x04, 0x08, 0x08, 0x08, 0x04, 0x02},
	')': {0x08, 0x04, 0x02, 0x02, 0x02, 0x04, 0x08},
	':': {0x00, 0x0C, 0x0C, 0x00, 0x0C, 0x0C, 0x00},
}

func drawLabel(frame *image.Paletted, x, y int, text string) {
	textColor := color.RGBA{220, 220, 220, 255}
	shadowColor := color.RGBA{0, 0, 0, 255}
	charWidth := 6 // 5 pixels + 1 spacing

	for i, c := range text {
		// Convert to uppercase
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}

		glyph, ok := font5x7[c]
		if !ok {
			continue // Skip unknown characters
		}

		charX := x + i*charWidth
		for row := 0; row < 7; row++ {
			for col := 0; col < 5; col++ {
				if glyph[row]&(1<<(4-col)) != 0 {
					// Draw shadow (offset by 1,1)
					frame.Set(charX+col+1, y+row+1, shadowColor)
				}
			}
		}
		for row := 0; row < 7; row++ {
			for col := 0; col < 5; col++ {
				if glyph[row]&(1<<(4-col)) != 0 {
					frame.Set(charX+col, y+row, textColor)
				}
			}
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
