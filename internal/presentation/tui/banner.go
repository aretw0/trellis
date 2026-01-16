package tui

import (
	"fmt"

	"github.com/muesli/termenv"
)

// PrintBanner outputs a professional ASCII art banner for Trellis.
func PrintBanner() {
	p := termenv.ColorProfile()
	// Using a subtle gradient-like color scheme (Indigo/Violet)
	s1 := termenv.String("  _____             _ _ _ ").Foreground(p.Color("#818cf8"))
	s2 := termenv.String(" |_   _|           | | | |").Foreground(p.Color("#a78bfa"))
	s3 := termenv.String("   | |  _ __ ___| | | (_)___").Foreground(p.Color("#c084fc"))
	s4 := termenv.String("   | | | '__/ _ \\ | | | / __|").Foreground(p.Color("#e879f9"))
	s5 := termenv.String("   | | | | |  __/ | | | \\__ \\").Foreground(p.Color("#f472b6"))
	s6 := termenv.String("   \\_/ |_|  \\___|_|_|_|_|___/").Foreground(p.Color("#fb7185"))

	fmt.Println()
	fmt.Println(s1)
	fmt.Println(s2)
	fmt.Println(s3)
	fmt.Println(s4)
	fmt.Println(s5)
	fmt.Println(s6)
	fmt.Println()
}
