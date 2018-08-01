package dice

import "github.com/richardwilkes/toolbox/xmath/rand"

// Config holds a configuration for Dice.
type Config struct {
	// Randomizer holds the randomizer that will be used.
	Randomizer rand.Randomizer
	// GURPSFormat determines whether GURPS dice formatting should be used. A
	// value of true means the die count is always shown and the sides value
	// is suppressed if it is a '6', while a value of false means the die
	// count is suppressed if it is a '1' and the sides value is always shown.
	GURPSFormat bool
	// ExtraDiceFromModifiers determines if modifiers greater than or equal to
	// the average result of the base die should be converted to extra dice.
	// For example, 1d6+8 will become 3d6+1.
	ExtraDiceFromModifiers bool
}

// DefaultGURPSConfig creates a Dice config suited for GURPS.
func DefaultGURPSConfig() *Config {
	return &Config{
		Randomizer:  rand.NewCryptoRand(),
		GURPSFormat: true,
	}
}
