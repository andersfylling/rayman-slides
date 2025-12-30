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
	// Hitbox relative to sprite origin
	HitX int `json:"hitX,omitempty"`
	HitY int `json:"hitY,omitempty"`
	HitW int `json:"hitW,omitempty"`
	HitH int `json:"hitH,omitempty"`
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
	jsonData, err := os.ReadFile("assets/sprites/default/atlas.json")
	if err != nil {
		return fmt.Errorf("reading atlas.json: %w", err)
	}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("parsing atlas.json: %w", err)
	}

	// Load atlas image (use filename from JSON)
	imgFile, err := os.Open("assets/sprites/default/" + data.Image)
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
		hitInfo := ""
		if region.HitW > 0 && region.HitH > 0 {
			if region.HitX != 0 || region.HitY != 0 || region.HitW != region.W || region.HitH != region.H {
				hitInfo = fmt.Sprintf(" hit:(%d,%d,%d,%d)", region.HitX, region.HitY, region.HitW, region.HitH)
			}
		}
		fmt.Printf("  %-25s  x:%-4d y:%-4d w:%-4d h:%-4d anchor:(%d,%d)%s\n",
			name, region.X, region.Y, region.W, region.H, region.AnchorX, region.AnchorY, hitInfo)
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

	width := 900
	height := 700

	numFrames := 32
	frameDelay := 8 // 80ms per frame

	for f := 0; f < numFrames; f++ {
		bounds := image.Rect(0, 0, width, height)
		frame := image.NewPaletted(bounds, palette)

		// Sky gradient background
		for y := 0; y < height; y++ {
			shade := uint8(60 + y*40/height)
			for x := 0; x < width; x++ {
				frame.Set(x, y, color.RGBA{shade / 2, shade / 2, shade, 255})
			}
		}

		// ======== SCENE 1: WORLD (top half) ========
		drawLabel(frame, 10, 10, "WORLD SCENE")

		groundY := 220
		tileSize := 50

		// Draw grass ground
		for x := 0; x < width; x += tileSize {
			drawSpriteAtSize(frame, "tile_grass", x, groundY, tileSize, tileSize)
		}

		// Draw some platforms with wood/stone
		drawSpriteAtSize(frame, "tile_wood", 300, groundY-80, tileSize, tileSize)
		drawSpriteAtSize(frame, "tile_wood", 350, groundY-80, tileSize, tileSize)
		drawSpriteAtSize(frame, "tile_stone", 400, groundY-80, tileSize, tileSize)

		// Cloud platform
		drawSpriteAtSize(frame, "tile_cloud", 550, groundY-150, 80, 50)
		drawSpriteAtSize(frame, "tile_cloud", 700, groundY-120, 80, 50)

		// Player walking on grass
		walkFrame := (f / 2 % 4) + 1
		walkSprite := fmt.Sprintf("player_walk_%d", walkFrame)
		playerX := 50 + (f * 6 % 200)
		drawSpriteAtScaled(frame, walkSprite, playerX, groundY, 0.6)

		// Player jumping on platform
		jumpPhase := f % 16
		jumpY := groundY - 80
		if jumpPhase < 8 {
			jumpY = groundY - 80 - jumpPhase*4
		} else {
			jumpY = groundY - 80 - (16-jumpPhase)*4
		}
		drawSpriteAtScaled(frame, "player_jump", 350, jumpY, 0.6)

		// Slime on ground bouncing
		slimeFrame := (f / 4 % 4) + 1
		slimeSprite := fmt.Sprintf("slime_%d", slimeFrame)
		bounceY := 0
		if f%8 < 4 {
			bounceY = (f % 4) * 4
		} else {
			bounceY = (4 - f%4) * 4
		}
		drawSpriteAtScaled(frame, slimeSprite, 650, groundY-bounceY, 0.5)

		// Bat flying
		batFrame := (f / 2 % 4) + 1
		batSprite := fmt.Sprintf("bat_%d", batFrame)
		batX := 750 + int(float64(f%16-8)*3)
		batY := groundY - 100 + int(float64(f%8-4)*2)
		drawSpriteAtScaled(frame, batSprite, batX, batY, 0.5)

		// Orbs floating
		orbFrame := (f / 4 % 3) + 1
		orbSprite := fmt.Sprintf("orb_%d", orbFrame)
		drawSpriteAtScaled(frame, orbSprite, 320, groundY-120, 0.5)
		drawSpriteAtScaled(frame, orbSprite, 370, groundY-120, 0.5)

		// ======== SCENE 2: COMBAT (middle) ========
		drawLabel(frame, 10, 280, "COMBAT SCENE")

		combatGroundY := 420
		// Draw dirt ground for combat area
		for x := 0; x < width; x += tileSize {
			drawSpriteAtSize(frame, "tile_dirt", x, combatGroundY, tileSize, tileSize)
		}

		// Player attacking
		var attackSprite string
		if f%16 < 8 {
			attackSprite = "player_attack_1"
		} else {
			attackSprite = "player_attack_2"
		}
		drawSpriteAtScaled(frame, attackSprite, 100, combatGroundY, 0.7)

		// Fist projectile flying
		fistFrame := (f / 3 % 3) + 1
		fistSprite := fmt.Sprintf("fist_%d", fistFrame)
		fistX := 180 + (f * 12 % 300)
		drawSpriteAtScaled(frame, fistSprite, fistX, combatGroundY-50, 0.6)

		// Slime getting hit (shaking)
		slimeHitX := 500 + (f%4 - 2)
		drawSpriteAtScaled(frame, slimeSprite, slimeHitX, combatGroundY-bounceY, 0.6)

		// Health pickup
		drawSpriteAtScaled(frame, "health", 700, combatGroundY-30, 0.6)

		// Cage with character
		drawSpriteAtScaled(frame, "cage_closed", 800, combatGroundY, 0.6)

		// ======== SCENE 3: HAZARDS (bottom) ========
		drawLabel(frame, 10, 480, "HAZARDS + TILES")

		hazardY := 620
		// Draw various hazard tiles
		drawSpriteAtSize(frame, "tile_spikes", 50, hazardY, 60, 60)
		drawLabel(frame, 50, hazardY+65, "SPIKES")

		drawSpriteAtSize(frame, "tile_water", 150, hazardY, 60, 60)
		drawLabel(frame, 150, hazardY+65, "WATER")

		drawSpriteAtSize(frame, "tile_fire", 250, hazardY, 60, 70)
		drawLabel(frame, 250, hazardY+75, "FIRE")

		// Show tile types
		drawSpriteAtSize(frame, "tile_grass", 380, hazardY, 50, 50)
		drawSpriteAtSize(frame, "tile_dirt", 440, hazardY, 50, 50)
		drawSpriteAtSize(frame, "tile_wood", 500, hazardY, 50, 50)
		drawSpriteAtSize(frame, "tile_stone", 560, hazardY, 50, 50)
		drawSpriteAtSize(frame, "tile_cloud", 620, hazardY, 70, 40)

		// Collectibles
		drawSpriteAtScaled(frame, "orb_1", 720, hazardY+20, 0.5)
		drawSpriteAtScaled(frame, "orb_2", 760, hazardY+20, 0.5)
		drawSpriteAtScaled(frame, "orb_3", 800, hazardY+20, 0.5)
		drawSpriteAtScaled(frame, "health", 850, hazardY+20, 0.5)

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

	// Define distinct colors for different sprite groups based on position
	colorAnchor := color.RGBA{255, 255, 255, 255} // White

	// Color palette for different groups
	groupColors := []color.RGBA{
		{0, 255, 0, 255},     // Green - walk frames
		{0, 200, 255, 255},   // Cyan - idle/jump
		{255, 100, 100, 255}, // Light red - attack
		{255, 0, 255, 255},   // Magenta - fist
		{0, 100, 255, 255},   // Blue - tiles row 1
		{255, 255, 0, 255},   // Yellow - items (ting, health, cage)
		{255, 150, 0, 255},   // Orange - slimes
		{200, 0, 200, 255},   // Purple - bats
		{100, 255, 100, 255}, // Light green - slime death / misc
		{0, 255, 200, 255},   // Teal - tiles row 2
		{255, 100, 200, 255}, // Pink - effects
	}

	// Hitbox color
	hitboxColor := color.RGBA{255, 100, 100, 255} // Red for hitbox

	// Draw borders and anchors for each sprite
	for name, region := range data.Sprites {
		// Choose color based on position in atlas
		var borderColor color.RGBA

		// Determine group by Y position and X position
		if region.Y < 160 { // Row 1
			if region.X < 600 {
				borderColor = groupColors[0] // Green - walk frames
			} else if region.X < 720 {
				borderColor = groupColors[1] // Cyan - jump
			} else {
				borderColor = groupColors[6] // Orange - slimes
			}
		} else if region.Y < 350 { // Row 2
			if region.X < 600 {
				borderColor = groupColors[2] // Light red - attack frames
			} else {
				borderColor = groupColors[7] // Purple - bats
			}
		} else if region.Y < 480 { // Row 3
			if region.X < 500 {
				borderColor = groupColors[3] // Magenta - fist frames
			} else {
				borderColor = groupColors[8] // Light green - misc
			}
		} else if region.Y < 620 { // Row 4
			if region.X < 530 {
				borderColor = groupColors[4] // Blue - tiles
			} else if region.X < 800 {
				borderColor = groupColors[5] // Yellow - tings
			} else {
				borderColor = groupColors[5] // Yellow - health/cage
			}
		} else { // Row 5
			if region.X < 400 {
				borderColor = groupColors[9] // Teal - hazard tiles
			} else {
				borderColor = groupColors[10] // Pink - effects
			}
		}

		// Draw border rectangle (2px thick for visibility)
		drawBorderRGBA(overlay, region.X, region.Y, region.W, region.H, borderColor, 2)

		// Draw hitbox if different from visual bounds
		if region.HitW > 0 && region.HitH > 0 {
			if region.HitX != 0 || region.HitY != 0 || region.HitW != region.W || region.HitH != region.H {
				drawBorderRGBA(overlay, region.X+region.HitX, region.Y+region.HitY,
					region.HitW, region.HitH, hitboxColor, 1)
			}
		}

		// Draw anchor point as a cross
		anchorX := region.X + region.AnchorX
		anchorY := region.Y + region.AnchorY
		drawCrossRGBA(overlay, anchorX, anchorY, colorAnchor, 5)

		// Draw label with ID at top-left of sprite region
		bgColor := color.RGBA{0, 0, 0, 200}
		drawLabelRGBA(overlay, region.X+3, region.Y+3, name, borderColor, bgColor)
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
	fmt.Println("  Colored borders = visual sprite bounds")
	fmt.Println("  Red inner boxes = hitboxes (when different from visual)")
	fmt.Println("  White crosses = anchor points")
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

// drawLabelRGBA draws text on an RGBA image
func drawLabelRGBA(img *image.RGBA, x, y int, text string, textColor, bgColor color.RGBA) {
	charWidth := 6 // 5 pixels + 1 spacing
	textWidth := len(text) * charWidth
	textHeight := 9

	// Draw background rectangle
	for dy := -1; dy < textHeight; dy++ {
		for dx := -2; dx < textWidth+2; dx++ {
			img.Set(x+dx, y+dy, bgColor)
		}
	}

	for i, c := range text {
		// Convert to uppercase
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}

		glyph, ok := font5x7[c]
		if !ok {
			continue
		}

		charX := x + i*charWidth
		for row := 0; row < 7; row++ {
			for col := 0; col < 5; col++ {
				if glyph[row]&(1<<(4-col)) != 0 {
					img.Set(charX+col, y+row, textColor)
				}
			}
		}
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

func drawSpriteAtScaled(frame *image.Paletted, spriteName string, x, y int, scale float64) {
	region, ok := data.Sprites[spriteName]
	if !ok {
		return
	}

	w := int(float64(region.W) * scale)
	h := int(float64(region.H) * scale)
	anchorX := int(float64(region.AnchorX) * scale)
	anchorY := int(float64(region.AnchorY) * scale)

	drawX := x - anchorX
	drawY := y - anchorY

	spriteRect := image.Rect(region.X, region.Y, region.X+region.W, region.Y+region.H)
	drawRegionScaled(frame, atlasImg, spriteRect, image.Pt(drawX, drawY), w, h)
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
