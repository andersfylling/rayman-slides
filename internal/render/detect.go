// Package render implements terminal rendering with multiple backends.
package render

import (
	"os"
	"strings"
)

// Capability represents detected terminal capabilities
type Capability struct {
	Truecolor bool
	Color256  bool
	Unicode   bool
}

// Detect probes the terminal for capabilities
func Detect() Capability {
	cap := Capability{}

	// Check COLORTERM for truecolor
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		cap.Truecolor = true
		cap.Color256 = true
	}

	// Check TERM for 256 color
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") {
		cap.Color256 = true
	}

	// Assume unicode support if not explicitly disabled
	// Most modern terminals support it
	lang := os.Getenv("LANG")
	cap.Unicode = strings.Contains(strings.ToLower(lang), "utf")
	if cap.Unicode == false {
		// Default to true for modern terminals
		cap.Unicode = true
	}

	return cap
}

// SelectRenderer picks the best renderer for the terminal
// Currently returns TcellRenderer which supports all modes internally
func SelectRenderer(cap Capability, override Mode) GameRenderer {
	// TcellRenderer handles all terminal modes via tcell
	// In the future, we can return different implementations:
	// - SDLRenderer for graphical output
	// - VulkanRenderer for GPU-accelerated graphics
	renderer := NewTcellRenderer()

	// Configure atlas based on mode
	switch override {
	case ModeASCII:
		renderer.SetAtlas(DefaultASCIIAtlas())
	case ModeHalfBlock:
		renderer.SetAtlas(DefaultHalfBlockAtlas())
	case ModeBraille:
		// TODO: BrailleAtlas when implemented
		renderer.SetAtlas(DefaultASCIIAtlas())
	default:
		// Auto: use ASCII for now, could be smarter based on capabilities
		renderer.SetAtlas(DefaultASCIIAtlas())
	}

	return renderer
}

// Mode selects the rendering mode
type Mode int

const (
	ModeAuto Mode = iota
	ModeASCII
	ModeHalfBlock
	ModeBraille
)
