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

	var data AtlasData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("parsing atlas.json: %w", err)
	}

	// Load atlas image
	imgFile, err := os.Open("assets/sprites/atlas.jpg")
	if err != nil {
		return fmt.Errorf("opening atlas image: %w", err)
	}
	defer imgFile.Close()

	atlasImg, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("decoding atlas image: %w", err)
	}

	// Sort sprite names for consistent ordering
	var names []string
	for name := range data.Sprites {
		names = append(names, name)
	}
	sort.Strings(names)

	// Create GIF with one frame per sprite
	outGif := &gif.GIF{}

	// Calculate grid layout
	cols := 8
	cellW, cellH := 80, 80 // Each cell size
	rows := (len(names) + cols - 1) / cols
	gridW := cols * cellW
	gridH := rows * cellH

	// Create a single frame showing all sprites in a grid with labels
	bounds := image.Rect(0, 0, gridW, gridH)
	palette := buildPalette(atlasImg)
	frame := image.NewPaletted(bounds, palette)

	// Fill with dark background
	draw.Draw(frame, bounds, &image.Uniform{color.RGBA{30, 30, 40, 255}}, image.Point{}, draw.Src)

	// Draw each sprite in the grid
	for i, name := range names {
		region := data.Sprites[name]
		col := i % cols
		row := i / cols

		// Calculate cell position
		cellX := col * cellW
		cellY := row * cellH

		// Extract sprite from atlas
		spriteRect := image.Rect(region.X, region.Y, region.X+region.W, region.Y+region.H)

		// Calculate centered position within cell
		offsetX := cellX + (cellW-region.W)/2
		offsetY := cellY + (cellH-region.H)/2

		// Draw sprite
		drawRegion(frame, atlasImg, spriteRect, image.Pt(offsetX, offsetY))

		// Draw red dot at anchor point
		anchorScreenX := offsetX + region.AnchorX
		anchorScreenY := offsetY + region.AnchorY
		drawDot(frame, anchorScreenX, anchorScreenY, color.RGBA{255, 0, 0, 255})

		// Draw border around sprite region
		drawBorder(frame, offsetX, offsetY, region.W, region.H, color.RGBA{100, 100, 100, 255})
	}

	outGif.Image = append(outGif.Image, frame)
	outGif.Delay = append(outGif.Delay, 0) // No animation, just static

	// Also create individual sprite frames for animation preview
	fmt.Println("Sprites in atlas:")
	fmt.Println("─────────────────")
	for _, name := range names {
		region := data.Sprites[name]
		fmt.Printf("  %-30s  x:%-4d y:%-4d w:%-4d h:%-4d anchor:(%d,%d)\n",
			name, region.X, region.Y, region.W, region.H, region.AnchorX, region.AnchorY)
	}
	fmt.Printf("\nTotal: %d sprites\n", len(names))

	// Write GIF
	outFile, err := os.Create("sprites.debug.gif")
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer outFile.Close()

	if err := gif.EncodeAll(outFile, outGif); err != nil {
		return fmt.Errorf("encoding gif: %w", err)
	}

	fmt.Printf("\nGenerated: sprites.debug.gif (%dx%d)\n", gridW, gridH)
	fmt.Println("Red dots show anchor points")
	return nil
}

// buildPalette creates a 256-color palette from the atlas image
func buildPalette(img image.Image) color.Palette {
	// Start with some basic colors
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},       // Black
		color.RGBA{30, 30, 40, 255},    // Dark background
		color.RGBA{255, 255, 255, 255}, // White
		color.RGBA{255, 0, 0, 255},     // Red (anchor)
		color.RGBA{100, 100, 100, 255}, // Gray (border)
	}

	// Sample colors from the image
	bounds := img.Bounds()
	colorMap := make(map[color.RGBA]int)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			r, g, b, a := img.At(x, y).RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			colorMap[c]++
		}
	}

	// Sort by frequency and take top colors
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

	// Add most common colors to palette
	for i := 0; i < len(counts) && len(palette) < 256; i++ {
		// Skip if too similar to existing
		c := counts[i].c
		duplicate := false
		for _, existing := range palette {
			er, eg, eb, _ := existing.RGBA()
			if abs(int(c.R)-int(er>>8)) < 8 &&
				abs(int(c.G)-int(eg>>8)) < 8 &&
				abs(int(c.B)-int(eb>>8)) < 8 {
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

func drawRegion(dst *image.Paletted, src image.Image, srcRect image.Rectangle, dstPt image.Point) {
	for y := 0; y < srcRect.Dy(); y++ {
		for x := 0; x < srcRect.Dx(); x++ {
			c := src.At(srcRect.Min.X+x, srcRect.Min.Y+y)
			_, _, _, a := c.RGBA()
			if a > 128 { // Only draw non-transparent pixels
				dst.Set(dstPt.X+x, dstPt.Y+y, c)
			}
		}
	}
}

func drawDot(dst *image.Paletted, x, y int, c color.Color) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if dx*dx+dy*dy <= 4 {
				dst.Set(x+dx, y+dy, c)
			}
		}
	}
}

func drawBorder(dst *image.Paletted, x, y, w, h int, c color.Color) {
	// Top and bottom
	for dx := 0; dx < w; dx++ {
		dst.Set(x+dx, y, c)
		dst.Set(x+dx, y+h-1, c)
	}
	// Left and right
	for dy := 0; dy < h; dy++ {
		dst.Set(x, y+dy, c)
		dst.Set(x+w-1, y+dy, c)
	}
}
