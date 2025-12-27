package render

// SpriteData holds renderer-specific sprite information
type SpriteData struct {
	Char  rune  // ASCII/Unicode character
	FG    Color // Foreground color
	BG    Color // Background color (usually transparent/black)
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

	// Player sprites
	atlas.Set("player", SpriteData{Char: '@', FG: ColorGreen, BG: ColorBlack})
	atlas.Set("player_idle", SpriteData{Char: '@', FG: ColorGreen, BG: ColorBlack})
	atlas.Set("player_walk", SpriteData{Char: '@', FG: ColorGreen, BG: ColorBlack})
	atlas.Set("player_jump", SpriteData{Char: '^', FG: ColorGreen, BG: ColorBlack})

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
