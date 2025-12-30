//go:build gio

package render

import (
	"encoding/json"
	"image"
	"io/fs"
	_ "image/jpeg"
	_ "image/png"
)

// SpriteRegion defines a rectangular region in the atlas
type SpriteRegion struct {
	X       int  `json:"x"`
	Y       int  `json:"y"`
	W       int  `json:"w"`
	H       int  `json:"h"`
	AnchorX int  `json:"anchorX"`
	AnchorY int  `json:"anchorY"`
	FlipX   bool `json:"flipX,omitempty"`
}

// AtlasData is the JSON structure for atlas metadata
type AtlasData struct {
	Image   string                  `json:"image"`
	Sprites map[string]SpriteRegion `json:"sprites"`
}

// Atlas holds the sprite sheet image and lookup table
type Atlas struct {
	Image   image.Image
	Sprites map[string]SpriteRegion
}

// LoadAtlas loads a sprite atlas from a filesystem using the default profile
func LoadAtlas(fsys fs.FS) (*Atlas, error) {
	return LoadAtlasProfile(fsys, "default")
}

// LoadAtlasProfile loads a sprite atlas from a specific profile folder
func LoadAtlasProfile(fsys fs.FS, profile string) (*Atlas, error) {
	basePath := "assets/sprites/" + profile

	// Load metadata
	jsonData, err := fs.ReadFile(fsys, basePath+"/atlas.json")
	if err != nil {
		return nil, err
	}

	var data AtlasData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	// Load image
	imgFile, err := fsys.Open(basePath + "/" + data.Image)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}

	return &Atlas{
		Image:   img,
		Sprites: data.Sprites,
	}, nil
}

// GetRegion returns the sprite region for an ID, with fallback
func (a *Atlas) GetRegion(id string) (SpriteRegion, bool) {
	if region, ok := a.Sprites[id]; ok {
		return region, true
	}
	// Fallback to default sprites
	if region, ok := a.Sprites["player"]; ok && len(id) >= 6 && id[:6] == "player" {
		return region, true
	}
	return SpriteRegion{}, false
}

// SubImage returns the image for a specific sprite region
func (a *Atlas) SubImage(region SpriteRegion) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	if si, ok := a.Image.(subImager); ok {
		return si.SubImage(image.Rect(region.X, region.Y, region.X+region.W, region.Y+region.H))
	}
	return a.Image
}
