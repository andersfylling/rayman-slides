package collision

// AABB is an axis-aligned bounding box
type AABB struct {
	X, Y          float64 // Top-left corner
	Width, Height float64
}

// NewAABB creates a bounding box
func NewAABB(x, y, w, h float64) AABB {
	return AABB{X: x, Y: y, Width: w, Height: h}
}

// Center returns the center point of the box
func (a AABB) Center() (float64, float64) {
	return a.X + a.Width/2, a.Y + a.Height/2
}

// Overlaps checks if two boxes overlap
func (a AABB) Overlaps(b AABB) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

// Contains checks if a point is inside the box
func (a AABB) Contains(x, y float64) bool {
	return x >= a.X && x < a.X+a.Width &&
		y >= a.Y && y < a.Y+a.Height
}

// Penetration returns how much b penetrates into a (for resolution)
func (a AABB) Penetration(b AABB) (float64, float64) {
	if !a.Overlaps(b) {
		return 0, 0
	}

	// Calculate overlap on each axis
	left := (b.X + b.Width) - a.X
	right := (a.X + a.Width) - b.X
	top := (b.Y + b.Height) - a.Y
	bottom := (a.Y + a.Height) - b.Y

	// Find minimum penetration axis
	dx, dy := 0.0, 0.0

	if left < right {
		dx = -left
	} else {
		dx = right
	}

	if top < bottom {
		dy = -top
	} else {
		dy = bottom
	}

	// Return smallest axis to resolve
	if abs(dx) < abs(dy) {
		return dx, 0
	}
	return 0, dy
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
