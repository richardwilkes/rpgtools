// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package dice simulates dice using standard roleplaying game notation.
package dice

import (
	"bytes"
	"math"
	"regexp"
	"strconv"

	"github.com/richardwilkes/toolbox/xmath/rand"
)

const (
	diceRegexStr = `(((\d+)?[dD](\d+))|((\d+)[dD](\d+)?)){1}([-+]\d+)?([xX](\d+))?`
	defaultSides = 6
)

var (
	// DefaultConfig is used if nil is passed in for a configuration. It is
	// also used when unmarshaling.
	DefaultConfig = &Config{Randomizer: rand.NewCryptoRand()}
	diceRegex     = regexp.MustCompile(diceRegexStr)
	diceRegexOnly = regexp.MustCompile(`^\s*` + diceRegexStr + `\s*$`)
)

// Dice holds the dice information.
type Dice struct {
	Config     *Config
	Count      int
	Sides      int
	Modifier   int
	Multiplier int
}

// Roll the dice. The spec will be used to create a new Dice object and the
// result of a single roll will be returned. May pass nil for cfg to get a
// default configuration.
func Roll(cfg *Config, spec string) int {
	return New(cfg, spec).Roll()
}

// New creates a new Dice based on the given configuration and spec. May pass
// nil for cfg to get a default configuration.
func New(cfg *Config, spec string) *Dice {
	if cfg == nil {
		cfg = DefaultConfig
	} else if cfg.Randomizer == nil {
		panic("cfg.Randomizer must be specified") // @allow
	}
	dice := &Dice{Config: cfg}
	match := diceRegexOnly.FindStringSubmatch(spec)
	if match != nil {
		if match[2] != "" {
			dice.Count = atoi(match[3])
			dice.Sides = atoi(match[4])
		} else {
			dice.Count = atoi(match[6])
			dice.Sides = atoi(match[7])
		}
		dice.Modifier = atoi(match[8])
		dice.Multiplier = atoi(match[10])
	}
	dice.Normalize()
	return dice
}

func atoi(text string) int {
	if value, err := strconv.Atoi(text); err == nil {
		return value
	}
	return 0
}

// Minimum returns the minimum result.
func (dice *Dice) Minimum() int {
	count, result := dice.adjustedCountAndModifier(dice.Config.ExtraDiceFromModifiers)
	result += count
	return result * dice.Multiplier
}

// Average returns the average result.
func (dice *Dice) Average() int {
	count, result := dice.adjustedCountAndModifier(dice.Config.ExtraDiceFromModifiers)
	result += count * (dice.Sides + 1) / 2
	return result * dice.Multiplier
}

// Maximum returns the maximum result.
func (dice *Dice) Maximum() int {
	count, result := dice.adjustedCountAndModifier(dice.Config.ExtraDiceFromModifiers)
	result += count * dice.Sides
	return result * dice.Multiplier
}

// Roll returns the result of rolling the dice.
func (dice *Dice) Roll() int {
	count, result := dice.adjustedCountAndModifier(dice.Config.ExtraDiceFromModifiers)
	for i := 0; i < count; i++ {
		result += 1 + dice.Config.Randomizer.Intn(dice.Sides)
	}
	return result * dice.Multiplier
}

func (dice *Dice) String() string {
	count, modifier := dice.adjustedCountAndModifier(dice.Config.ExtraDiceFromModifiers)
	var buffer bytes.Buffer
	if count > 0 {
		if dice.Config.GURPSFormat || count > 1 {
			buffer.WriteString(strconv.Itoa(count))
		}
		buffer.WriteString("d")
		if !dice.Config.GURPSFormat || dice.Sides != defaultSides {
			buffer.WriteString(strconv.Itoa(dice.Sides))
		}
	}
	if modifier > 0 {
		buffer.WriteString("+")
		buffer.WriteString(strconv.Itoa(modifier))
	} else if modifier < 0 {
		buffer.WriteString(strconv.Itoa(modifier))
	}
	if dice.Multiplier != 1 {
		buffer.WriteString("x")
		buffer.WriteString(strconv.Itoa(dice.Multiplier))
	}
	if buffer.Len() == 0 {
		buffer.WriteString("0")
	}
	return buffer.String()
}

// ApplyExtraDiceFromModifiers alters the Dice to reflect any adjustment that
// would be made if the ExtraDiceFromModifiers configuration flag was enabled.
func (dice *Dice) ApplyExtraDiceFromModifiers() {
	dice.Count, dice.Modifier = dice.adjustedCountAndModifier(true)
}

func (dice *Dice) adjustedCountAndModifier(applyExtractDiceFromModifiers bool) (count, modifier int) {
	dice.Normalize()
	count = dice.Count
	modifier = dice.Modifier
	if applyExtractDiceFromModifiers && modifier > 0 {
		average := (dice.Sides + 1) / 2
		if dice.Sides&1 == 1 {
			// Odd number of sides, so average is a whole number
			count += modifier / average
			modifier %= average
		} else {
			// Even number of sides, so average has an extra half, which means
			// we alternate
			for modifier > average {
				if modifier > 2*average {
					modifier -= 2*average + 1
					count += 2
				} else {
					modifier -= average + 1
					count++
				}
			}
		}
	}
	if count < 0 {
		count = 0
	}
	return
}

// Normalize the internal state.
func (dice *Dice) Normalize() {
	if dice.Count < 1 {
		dice.Count = 1
	}
	if dice.Sides < 1 {
		dice.Sides = defaultSides
	}
	if dice.Multiplier < 1 {
		dice.Multiplier = 1
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (dice Dice) MarshalText() (text []byte, err error) {
	return []byte(dice.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (dice *Dice) UnmarshalText(text []byte) error {
	*dice = *New(nil, string(text))
	return nil
}

// IsEquivalent returns true if this Dice is equivalent to another Dice.
// Normalizes both Dice.
func (dice *Dice) IsEquivalent(other *Dice) bool {
	dice.Normalize()
	other.Normalize()
	return dice.Count == other.Count && dice.Sides == other.Sides && dice.Modifier == other.Modifier && dice.Multiplier == other.Multiplier
}

// PoolProbability return the probability that at least one die will be equal
// to or greater than the target value.
func (dice *Dice) PoolProbability(target int) float64 {
	dice.Normalize()
	if dice.Count < 1 || dice.Sides < target {
		return 0
	}
	return 1 - math.Pow(1-float64(1+dice.Sides-target)/float64(dice.Sides), float64(dice.Count))
}

// ExtractFirstPosition returns a two-element slice of integers defining the
// location of the first Dice specification in the input text it finds. The
// match itself is at text[loc[0]:loc[1]]. A return value of nil indicates no
// match was found.
func ExtractFirstPosition(text string) []int {
	return diceRegex.FindStringIndex(text)
}

// ExtractAllPositions is similar to ExtractFirstPosition, except it returns a
// slice of up to max matches. If max is less than 1, then all matches will be
// returned. A return value of nil indicates no matches were found.
func ExtractAllPositions(text string, max int) [][]int {
	if max < 1 {
		max = -1
	}
	return diceRegex.FindAllStringIndex(text, max)
}
