package render

// SpriteData holds renderer-specific sprite information
type SpriteData struct {
	Char   rune     // ASCII/Unicode character (used if Lines is empty)
	Lines  []string // Multi-line ASCII art (takes priority over Char)
	Anchor struct { // Position within sprite that maps to entity position
		X, Y int
	}
	FG Color // Foreground color
	BG Color // Background color (usually transparent/black)
}

// Width returns the width of the sprite in cells
func (s SpriteData) Width() int {
	if len(s.Lines) == 0 {
		return 1
	}
	maxW := 0
	for _, line := range s.Lines {
		if len(line) > maxW {
			maxW = len(line)
		}
	}
	return maxW
}

// Height returns the height of the sprite in cells
func (s SpriteData) Height() int {
	if len(s.Lines) == 0 {
		return 1
	}
	return len(s.Lines)
}

// SpriteAtlas maps sprite IDs to renderer-specific representations
type SpriteAtlas struct {
	sprites map[string]SpriteData
	fallback SpriteData
}

// NewSpriteAtlas creates a new sprite atlas with a fallback sprite
func NewSpriteAtlas() *SpriteAtlas {
	return &SpriteAtlas{
		sprites: make(map[string]SpriteData),
		fallback: SpriteData{
			Char: '?',
			FG:   ColorMagenta,
			BG:   ColorBlack,
		},
	}
}

// Set adds or updates a sprite mapping
func (a *SpriteAtlas) Set(id string, data SpriteData) {
	a.sprites[id] = data
}

// Get returns the sprite data for an ID, or fallback if not found
func (a *SpriteAtlas) Get(id string) SpriteData {
	if data, ok := a.sprites[id]; ok {
		return data
	}
	return a.fallback
}

// Has checks if a sprite ID exists in the atlas
func (a *SpriteAtlas) Has(id string) bool {
	_, ok := a.sprites[id]
	return ok
}

// DefaultASCIIAtlas returns a sprite atlas with default ASCII mappings
func DefaultASCIIAtlas() *SpriteAtlas {
	atlas := NewSpriteAtlas()

	// Player sprites - multi-line ASCII art
	// Anchor at feet position (center-bottom) for physics alignment
	playerSprite := SpriteData{
		Lines: []string{
			" O ",
			"/|\\",
			"/ \\",
		},
		FG: ColorGreen,
		BG: ColorBlack,
	}
	playerSprite.Anchor.X = 1 // Center horizontally
	playerSprite.Anchor.Y = 2 // Bottom row is the feet
	atlas.Set("player", playerSprite)
	atlas.Set("player_idle", playerSprite)
	atlas.Set("player_walk", playerSprite)

	// Player jumping sprite
	playerJumpSprite := SpriteData{
		Lines: []string{
			"\\O/",
			" | ",
			"/ \\",
		},
		FG: ColorGreen,
		BG: ColorBlack,
	}
	playerJumpSprite.Anchor.X = 1
	playerJumpSprite.Anchor.Y = 2
	atlas.Set("player_jump", playerJumpSprite)

	// Player punching right
	playerPunchRight := SpriteData{
		Lines: []string{
			" O ",
			"/|--*",
			"/ \\",
		},
		FG: ColorGreen,
		BG: ColorBlack,
	}
	playerPunchRight.Anchor.X = 1
	playerPunchRight.Anchor.Y = 2
	atlas.Set("player_punch_right", playerPunchRight)

	// Player punching left
	playerPunchLeft := SpriteData{
		Lines: []string{
			" O ",
			"*--|\\",
			"   / \\",
		},
		FG: ColorGreen,
		BG: ColorBlack,
	}
	playerPunchLeft.Anchor.X = 4 // Fist is at left, body center at 4
	playerPunchLeft.Anchor.Y = 2
	atlas.Set("player_punch_left", playerPunchLeft)

	// Enemy sprites
	atlas.Set("slime", SpriteData{Char: 's', FG: ColorGreen, BG: ColorBlack})
	atlas.Set("bat", SpriteData{Char: 'b', FG: ColorMagenta, BG: ColorBlack})
	atlas.Set("enemy", SpriteData{Char: 'E', FG: ColorRed, BG: ColorBlack})

	// Collectibles
	atlas.Set("ting", SpriteData{Char: 'o', FG: ColorYellow, BG: ColorBlack})
	atlas.Set("cage", SpriteData{Char: '#', FG: ColorCyan, BG: ColorBlack})
	atlas.Set("health", SpriteData{Char: '+', FG: ColorRed, BG: ColorBlack})

	// Environment
	atlas.Set("platform", SpriteData{Char: '=', FG: ColorWhite, BG: ColorBlack})
	atlas.Set("ladder", SpriteData{Char: 'H', FG: Color{139, 69, 19}, BG: ColorBlack})
	atlas.Set("water", SpriteData{Char: '~', FG: ColorBlue, BG: ColorBlack})
	atlas.Set("hazard", SpriteData{Char: '^', FG: ColorRed, BG: ColorBlack})

	return atlas
}

// DefaultHalfBlockAtlas returns a sprite atlas optimized for half-block rendering
func DefaultHalfBlockAtlas() *SpriteAtlas {
	atlas := NewSpriteAtlas()

	// Half-block rendering uses block characters with colors
	// The character matters less; color is primary

	atlas.Set("player", SpriteData{Char: '█', FG: ColorGreen, BG: ColorBlack})
	atlas.Set("slime", SpriteData{Char: '▄', FG: ColorGreen, BG: ColorBlack})
	atlas.Set("bat", SpriteData{Char: '▀', FG: ColorMagenta, BG: ColorBlack})

	// Add more as needed...

	return atlas
}
