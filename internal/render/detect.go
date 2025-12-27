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
func SelectRenderer(cap Capability, override Mode) Renderer {
	mode := override

	// Auto-detect if no override
	if mode == ModeAuto {
		switch {
		case cap.Truecolor:
			mode = ModeHalfBlock
		case cap.Color256:
			mode = ModeHalfBlock
		default:
			mode = ModeASCII
		}
	}

	switch mode {
	case ModeASCII:
		return NewASCIIRenderer()
	case ModeHalfBlock:
		return NewHalfBlockRenderer(cap.Truecolor)
	case ModeBraille:
		return NewBrailleRenderer()
	default:
		return NewASCIIRenderer()
	}
}

// Mode selects the rendering mode
type Mode int

const (
	ModeAuto Mode = iota
	ModeASCII
	ModeHalfBlock
	ModeBraille
)
