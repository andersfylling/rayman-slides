//go:build gio

// Command sprite-editor is an interactive tool for defining sprite regions.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type SpriteRegion struct {
	X       int `json:"x"`
	Y       int `json:"y"`
	W       int `json:"w"`
	H       int `json:"h"`
	AnchorX int `json:"anchorX"`
	AnchorY int `json:"anchorY"`
	// Hitbox relative to sprite origin (top-left of visual box)
	HitX int `json:"hitX,omitempty"`
	HitY int `json:"hitY,omitempty"`
	HitW int `json:"hitW,omitempty"`
	HitH int `json:"hitH,omitempty"`
}

type AtlasData struct {
	Image   string                  `json:"image"`
	Sprites map[string]SpriteRegion `json:"sprites"`
}

type Box struct {
	Name string
	// Visual sprite bounds (in atlas coordinates)
	X, Y, W, H int
	// Anchor point (relative to sprite top-left)
	AnchorX, AnchorY int
	// Hitbox (relative to sprite top-left)
	HitX, HitY, HitW, HitH int
}

// Edit mode determines what arrow keys modify
type EditMode int

const (
	ModeBox    EditMode = iota // Edit visual sprite box
	ModeAnchor                 // Edit anchor point
	ModeHitbox                 // Edit hitbox
)

var (
	atlasImg    image.Image
	atlasOp     paint.ImageOp
	boxes       []Box
	selectedIdx int = -1 // -1 = no selection

	editMode EditMode = ModeBox

	zoom       float32 = 1.0
	panX, panY float32 = 0, 0

	// Dragging state
	dragging      bool
	dragStartX    float32
	dragStartY    float32
	dragMode      string // "pan", "draw", "move"
	drawingBox    *Box
	dragOffsetX   int
	dragOffsetY   int

	tag        struct{}
	nextBoxNum int = 1
)

func main() {
	// Load atlas image
	imgFile, err := os.Open("assets/sprites/default/atlas.jpg")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening atlas: %v\n", err)
		os.Exit(1)
	}
	defer imgFile.Close()

	atlasImg, _, err = image.Decode(imgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding atlas: %v\n", err)
		os.Exit(1)
	}
	atlasOp = paint.NewImageOp(atlasImg)

	// Load existing sprites from atlas.json
	loadExistingSprites()

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  SPRITE EDITOR")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("  NAVIGATION:")
	fmt.Println("    Mouse wheel     = Zoom in/out")
	fmt.Println("    Middle-drag     = Pan around")
	fmt.Println("    +/-             = Zoom in/out")
	fmt.Println("")
	fmt.Println("  EDIT MODES (current shown in status):")
	fmt.Println("    1 or B          = BOX mode (visual sprite bounds)")
	fmt.Println("    2 or A          = ANCHOR mode (anchor point)")
	fmt.Println("    3 or H          = HITBOX mode (collision box)")
	fmt.Println("")
	fmt.Println("  EDITING:")
	fmt.Println("    Left-drag       = Draw new box")
	fmt.Println("    Left-click box  = Select box")
	fmt.Println("    Arrow keys      = Move/resize based on mode (1px)")
	fmt.Println("    Shift+Arrows    = Move/resize (10px)")
	fmt.Println("    Ctrl+Arrows     = Resize (in box/hitbox mode)")
	fmt.Println("    Delete/Backsp   = Delete selected box")
	fmt.Println("    Escape          = Deselect")
	fmt.Println("")
	fmt.Println("  FILE:")
	fmt.Println("    S               = Save to atlas.json")
	fmt.Println("    D               = Dump state to console")
	fmt.Println("    C               = Clear all boxes")
	fmt.Println("")
	fmt.Println("  COLORS:")
	fmt.Println("    Green           = Visual sprite box")
	fmt.Println("    Yellow          = Selected box")
	fmt.Println("    Red             = Hitbox")
	fmt.Println("    Cyan cross      = Anchor point")
	fmt.Println("")
	fmt.Printf("  Atlas size: %dx%d\n", atlasImg.Bounds().Dx(), atlasImg.Bounds().Dy())
	fmt.Printf("  Loaded %d existing sprites\n", len(boxes))
	fmt.Println("═══════════════════════════════════════════════════════")

	go func() {
		w := new(app.Window)
		w.Option(app.Title("Sprite Editor"))
		w.Option(app.Size(unit.Dp(1400), unit.Dp(900)))
		if err := run(w); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loadExistingSprites() {
	data, err := os.ReadFile("assets/sprites/default/atlas.json")
	if err != nil {
		fmt.Printf("No existing atlas.json found, starting fresh\n")
		return
	}

	var atlas AtlasData
	if err := json.Unmarshal(data, &atlas); err != nil {
		fmt.Printf("Error parsing atlas.json: %v\n", err)
		return
	}

	for name, region := range atlas.Sprites {
		box := Box{
			Name:    name,
			X:       region.X,
			Y:       region.Y,
			W:       region.W,
			H:       region.H,
			AnchorX: region.AnchorX,
			AnchorY: region.AnchorY,
			HitX:    region.HitX,
			HitY:    region.HitY,
			HitW:    region.HitW,
			HitH:    region.HitH,
		}
		// If no hitbox defined, default to sprite bounds
		if box.HitW == 0 && box.HitH == 0 {
			box.HitX = 0
			box.HitY = 0
			box.HitW = region.W
			box.HitH = region.H
		}
		boxes = append(boxes, box)
		nextBoxNum++
	}
}

func run(w *app.Window) error {
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Create clickable area for the whole window
			area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
			event.Op(gtx.Ops, &tag)
			area.Pop()

			// Request focus every frame
			gtx.Execute(key.FocusCmd{Tag: &tag})

			handleKeyboard(gtx)
			handlePointer(gtx)
			render(gtx)

			w.Invalidate()
			e.Frame(gtx.Ops)
		}
	}
}

func handleKeyboard(gtx layout.Context) {
	// Check for focus events first
	for {
		ev, ok := gtx.Event(key.FocusFilter{Target: &tag})
		if !ok {
			break
		}
		if _, ok := ev.(key.FocusEvent); ok {
			// Focus event received
		}
	}

	// Single filter to catch all key events
	for {
		ev, ok := gtx.Event(key.Filter{Focus: &tag})
		if !ok {
			break
		}
		if ke, ok := ev.(key.Event); ok {
			if ke.State == key.Press {
				step := 1
				if ke.Modifiers.Contain(key.ModShift) {
					step = 10
				}
				resize := ke.Modifiers.Contain(key.ModCtrl)

				switch ke.Name {
				// Edit mode switches
				case "1", "B":
					editMode = ModeBox
					fmt.Println("Mode: BOX (visual bounds)")
				case "2", "A":
					editMode = ModeAnchor
					fmt.Println("Mode: ANCHOR")
				case "3", "H":
					editMode = ModeHitbox
					fmt.Println("Mode: HITBOX")

				case "S":
					saveAtlas()
				case "D":
					dumpToConsole()
				case "C":
					boxes = nil
					selectedIdx = -1
					nextBoxNum = 1
					fmt.Println("Cleared all boxes")
				case "+", "=":
					zoom *= 1.2
					if zoom > 8 {
						zoom = 8
					}
					fmt.Printf("Zoom: %.0f%%\n", zoom*100)
				case "-":
					zoom /= 1.2
					if zoom < 0.1 {
						zoom = 0.1
					}
					fmt.Printf("Zoom: %.0f%%\n", zoom*100)
				case key.NameEscape:
					selectedIdx = -1
					fmt.Println("Deselected")
				case key.NameDeleteBackward, key.NameDeleteForward:
					if selectedIdx >= 0 && selectedIdx < len(boxes) {
						fmt.Printf("Deleted: %s\n", boxes[selectedIdx].Name)
						boxes = append(boxes[:selectedIdx], boxes[selectedIdx+1:]...)
						selectedIdx = -1
					}
				case key.NameLeftArrow:
					handleArrow(-step, 0, resize)
				case key.NameRightArrow:
					handleArrow(step, 0, resize)
				case key.NameUpArrow:
					handleArrow(0, -step, resize)
				case key.NameDownArrow:
					handleArrow(0, step, resize)
				}
			}
		}
	}
}

func handleArrow(dx, dy int, resize bool) {
	if selectedIdx < 0 || selectedIdx >= len(boxes) {
		// No selection - pan instead
		panX -= float32(dx * 10)
		panY -= float32(dy * 10)
		return
	}

	b := &boxes[selectedIdx]

	switch editMode {
	case ModeBox:
		if resize {
			b.W += dx
			b.H += dy
			if b.W < 1 {
				b.W = 1
			}
			if b.H < 1 {
				b.H = 1
			}
		} else {
			b.X += dx
			b.Y += dy
		}

	case ModeAnchor:
		// Move anchor point within sprite
		b.AnchorX += dx
		b.AnchorY += dy
		// Clamp to sprite bounds
		if b.AnchorX < 0 {
			b.AnchorX = 0
		}
		if b.AnchorX > b.W {
			b.AnchorX = b.W
		}
		if b.AnchorY < 0 {
			b.AnchorY = 0
		}
		if b.AnchorY > b.H {
			b.AnchorY = b.H
		}

	case ModeHitbox:
		if resize {
			b.HitW += dx
			b.HitH += dy
			if b.HitW < 1 {
				b.HitW = 1
			}
			if b.HitH < 1 {
				b.HitH = 1
			}
		} else {
			b.HitX += dx
			b.HitY += dy
		}
	}

	printSelected()
}

func printSelected() {
	if selectedIdx >= 0 && selectedIdx < len(boxes) {
		b := boxes[selectedIdx]
		modeStr := "BOX"
		switch editMode {
		case ModeAnchor:
			modeStr = "ANCHOR"
		case ModeHitbox:
			modeStr = "HITBOX"
		}
		fmt.Printf("  [%s] %s: box=(%d,%d,%d,%d) anchor=(%d,%d) hit=(%d,%d,%d,%d)\n",
			modeStr, b.Name, b.X, b.Y, b.W, b.H, b.AnchorX, b.AnchorY,
			b.HitX, b.HitY, b.HitW, b.HitH)
	}
}

func handlePointer(gtx layout.Context) {
	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: &tag,
			Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Scroll,
		})
		if !ok {
			break
		}
		pe, ok := ev.(pointer.Event)
		if !ok {
			continue
		}

		// Convert screen coords to image coords
		imgX := (pe.Position.X - panX) / zoom
		imgY := (pe.Position.Y - panY) / zoom

		switch pe.Kind {
		case pointer.Scroll:
			// Use whichever scroll axis has movement
			scrollAmt := pe.Scroll.Y
			if scrollAmt == 0 {
				scrollAmt = pe.Scroll.X
			}

			if scrollAmt > 0 {
				zoom /= 1.2
			} else if scrollAmt < 0 {
				zoom *= 1.2
			}
			if zoom > 8 {
				zoom = 8
			}
			if zoom < 0.1 {
				zoom = 0.1
			}
			// Adjust pan to keep mouse position fixed
			panX = pe.Position.X - imgX*zoom
			panY = pe.Position.Y - imgY*zoom

		case pointer.Press:
			// Re-request keyboard focus on any click
			gtx.Execute(key.FocusCmd{Tag: &tag})

			if pe.Buttons.Contain(pointer.ButtonTertiary) {
				// Middle mouse = pan
				dragging = true
				dragMode = "pan"
				dragStartX = pe.Position.X - panX
				dragStartY = pe.Position.Y - panY
			} else if pe.Buttons.Contain(pointer.ButtonSecondary) {
				// Right click = delete box at position
				deleteBoxAt(int(imgX), int(imgY))
			} else if pe.Buttons.Contain(pointer.ButtonPrimary) {
				// Left click = select or start drawing
				idx := boxAt(int(imgX), int(imgY))
				if idx >= 0 {
					selectedIdx = idx
					dragging = true
					dragMode = "move"
					dragOffsetX = int(imgX) - boxes[idx].X
					dragOffsetY = int(imgY) - boxes[idx].Y
					fmt.Printf("Selected: %s\n", boxes[idx].Name)
					printSelected()
				} else {
					selectedIdx = -1
					dragging = true
					dragMode = "draw"
					dragStartX = imgX
					dragStartY = imgY
					drawingBox = &Box{
						Name: fmt.Sprintf("sprite_%d", nextBoxNum),
						X:    int(imgX),
						Y:    int(imgY),
						W:    1,
						H:    1,
					}
				}
			}

		case pointer.Drag:
			if dragging {
				switch dragMode {
				case "pan":
					panX = pe.Position.X - dragStartX
					panY = pe.Position.Y - dragStartY
				case "move":
					if selectedIdx >= 0 && selectedIdx < len(boxes) {
						boxes[selectedIdx].X = int(imgX) - dragOffsetX
						boxes[selectedIdx].Y = int(imgY) - dragOffsetY
					}
				case "draw":
					if drawingBox != nil {
						x1, x2 := dragStartX, imgX
						y1, y2 := dragStartY, imgY
						if x1 > x2 {
							x1, x2 = x2, x1
						}
						if y1 > y2 {
							y1, y2 = y2, y1
						}
						drawingBox.X = int(x1)
						drawingBox.Y = int(y1)
						drawingBox.W = int(x2 - x1)
						drawingBox.H = int(y2 - y1)
					}
				}
			}

		case pointer.Release:
			if dragging {
				if dragMode == "draw" && drawingBox != nil && drawingBox.W > 5 && drawingBox.H > 5 {
					// Set defaults for new box
					drawingBox.AnchorX = drawingBox.W / 2
					drawingBox.AnchorY = drawingBox.H // Bottom center
					drawingBox.HitX = 0
					drawingBox.HitY = 0
					drawingBox.HitW = drawingBox.W
					drawingBox.HitH = drawingBox.H
					boxes = append(boxes, *drawingBox)
					selectedIdx = len(boxes) - 1
					fmt.Printf("Created: %s\n", drawingBox.Name)
					printSelected()
					nextBoxNum++
				} else if dragMode == "move" && selectedIdx >= 0 {
					b := boxes[selectedIdx]
					fmt.Printf("Moved: %s to x=%d y=%d\n", b.Name, b.X, b.Y)
				}
				dragging = false
				dragMode = ""
				drawingBox = nil
			}
		}
	}
}

func boxAt(x, y int) int {
	// Return topmost (last) box at position
	for i := len(boxes) - 1; i >= 0; i-- {
		b := boxes[i]
		if x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H {
			return i
		}
	}
	return -1
}

func deleteBoxAt(x, y int) {
	idx := boxAt(x, y)
	if idx >= 0 {
		fmt.Printf("Deleted: %s\n", boxes[idx].Name)
		boxes = append(boxes[:idx], boxes[idx+1:]...)
		if selectedIdx == idx {
			selectedIdx = -1
		} else if selectedIdx > idx {
			selectedIdx--
		}
	}
}

func render(gtx layout.Context) layout.Dimensions {
	// Dark background
	paint.Fill(gtx.Ops, color.NRGBA{30, 30, 40, 255})

	// Apply zoom and pan
	offset := op.Offset(image.Pt(int(panX), int(panY))).Push(gtx.Ops)
	scale := op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(zoom, zoom))).Push(gtx.Ops)

	// Draw atlas image
	atlasOp.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Draw existing boxes
	for i, b := range boxes {
		// Draw visual box
		var boxColor color.NRGBA
		if i == selectedIdx {
			boxColor = color.NRGBA{255, 255, 0, 255} // Yellow for selected
		} else {
			boxColor = color.NRGBA{0, 255, 0, 200} // Green for others
		}
		drawBoxOutline(gtx.Ops, b.X, b.Y, b.W, b.H, boxColor, 2)

		// Draw hitbox (red, slightly inset visually)
		if b.HitW > 0 && b.HitH > 0 {
			hitColor := color.NRGBA{255, 80, 80, 200}
			if i == selectedIdx && editMode == ModeHitbox {
				hitColor = color.NRGBA{255, 100, 100, 255}
			}
			drawBoxOutline(gtx.Ops, b.X+b.HitX, b.Y+b.HitY, b.HitW, b.HitH, hitColor, 1)
		}

		// Draw anchor point as a cross
		anchorColor := color.NRGBA{0, 255, 255, 255} // Cyan
		if i == selectedIdx && editMode == ModeAnchor {
			anchorColor = color.NRGBA{255, 255, 0, 255} // Yellow when editing
		}
		drawCross(gtx.Ops, b.X+b.AnchorX, b.Y+b.AnchorY, anchorColor, 6)
	}

	// Draw box being created
	if drawingBox != nil {
		drawBoxOutline(gtx.Ops, drawingBox.X, drawingBox.Y, drawingBox.W, drawingBox.H,
			color.NRGBA{255, 100, 100, 255}, 2)
	}

	scale.Pop()
	offset.Pop()

	// Draw HUD (outside zoom/pan)
	drawHUD(gtx)

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func drawHUD(gtx layout.Context) {
	// Status bar at bottom
	statusY := gtx.Constraints.Max.Y - 30

	// Background for status bar
	statusRect := clip.Rect{
		Min: image.Pt(0, statusY-5),
		Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
	}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{0, 0, 0, 220})
	statusRect.Pop()

	// Mode indicator boxes at top
	modeY := 10
	modes := []struct {
		mode  EditMode
		label string
		key   string
	}{
		{ModeBox, "BOX", "1/B"},
		{ModeAnchor, "ANCHOR", "2/A"},
		{ModeHitbox, "HITBOX", "3/H"},
	}

	modeX := 10
	for _, m := range modes {
		bgColor := color.NRGBA{60, 60, 70, 200}
		if editMode == m.mode {
			switch m.mode {
			case ModeBox:
				bgColor = color.NRGBA{0, 150, 0, 255}
			case ModeAnchor:
				bgColor = color.NRGBA{0, 150, 150, 255}
			case ModeHitbox:
				bgColor = color.NRGBA{150, 50, 50, 255}
			}
		}

		modeRect := clip.Rect{
			Min: image.Pt(modeX, modeY),
			Max: image.Pt(modeX+80, modeY+25),
		}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, bgColor)
		modeRect.Pop()

		modeX += 90
		_ = m.label // Would need text rendering
	}
}

func drawBoxOutline(ops *op.Ops, x, y, w, h int, c color.NRGBA, t int) {
	// Top
	top := clip.Rect{Min: image.Pt(x, y), Max: image.Pt(x+w, y+t)}.Push(ops)
	paint.Fill(ops, c)
	top.Pop()
	// Bottom
	bot := clip.Rect{Min: image.Pt(x, y+h-t), Max: image.Pt(x+w, y+h)}.Push(ops)
	paint.Fill(ops, c)
	bot.Pop()
	// Left
	left := clip.Rect{Min: image.Pt(x, y), Max: image.Pt(x+t, y+h)}.Push(ops)
	paint.Fill(ops, c)
	left.Pop()
	// Right
	right := clip.Rect{Min: image.Pt(x+w-t, y), Max: image.Pt(x+w, y+h)}.Push(ops)
	paint.Fill(ops, c)
	right.Pop()
}

func drawCross(ops *op.Ops, x, y int, c color.NRGBA, size int) {
	// Horizontal line
	hLine := clip.Rect{
		Min: image.Pt(x-size, y-1),
		Max: image.Pt(x+size, y+1),
	}.Push(ops)
	paint.Fill(ops, c)
	hLine.Pop()

	// Vertical line
	vLine := clip.Rect{
		Min: image.Pt(x-1, y-size),
		Max: image.Pt(x+1, y+size),
	}.Push(ops)
	paint.Fill(ops, c)
	vLine.Pop()
}

func saveAtlas() {
	data := AtlasData{
		Image:   "atlas.jpg",
		Sprites: make(map[string]SpriteRegion),
	}

	for _, b := range boxes {
		region := SpriteRegion{
			X:       b.X,
			Y:       b.Y,
			W:       b.W,
			H:       b.H,
			AnchorX: b.AnchorX,
			AnchorY: b.AnchorY,
		}
		// Only include hitbox if it differs from visual bounds
		if b.HitX != 0 || b.HitY != 0 || b.HitW != b.W || b.HitH != b.H {
			region.HitX = b.HitX
			region.HitY = b.HitY
			region.HitW = b.HitW
			region.HitH = b.HitH
		}
		data.Sprites[b.Name] = region
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	if err := os.WriteFile("assets/sprites/default/atlas.json", jsonData, 0644); err != nil {
		fmt.Printf("Error writing atlas.json: %v\n", err)
		return
	}

	fmt.Printf("══════════════════════════════════════\n")
	fmt.Printf("  SAVED %d sprites to atlas.json\n", len(boxes))
	fmt.Printf("══════════════════════════════════════\n")
}

func dumpToConsole() {
	fmt.Println("\n══════════════════════════════════════")
	fmt.Println("  CURRENT STATE")
	fmt.Println("══════════════════════════════════════")
	for _, b := range boxes {
		fmt.Printf("  %s:\n", b.Name)
		fmt.Printf("    box:    x=%d y=%d w=%d h=%d\n", b.X, b.Y, b.W, b.H)
		fmt.Printf("    anchor: x=%d y=%d\n", b.AnchorX, b.AnchorY)
		fmt.Printf("    hitbox: x=%d y=%d w=%d h=%d\n", b.HitX, b.HitY, b.HitW, b.HitH)
	}
	fmt.Println("══════════════════════════════════════\n")
}
