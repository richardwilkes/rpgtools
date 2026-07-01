// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package names

import (
	"cmp"
	"iter"
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/xrand"
)

// weightedStep pairs a step (for SimpleNamer, a whole name) with the running cumulative weight through it (the prior
// steps' weights plus its own), which is the form pickWeighted consumes; the last entry in a slice therefore holds that
// slice's grand total.
type weightedStep[S cmp.Ordered] struct {
	step S
	last int64
}

// markovStepper supplies the per-step behavior that distinguishes the Markov namers: how a training name is broken into
// steps, how the lookup key evolves as steps are taken, and how a step contributes to a generated name.
type markovStepper[S cmp.Ordered] interface {
	// initialKey is the lookup key before any step has been taken.
	initialKey() string
	// steps decomposes a training name into the ordered sequence of steps it contributes.
	steps(name string) []S
	// advance returns the lookup key that follows taking step from key.
	advance(key string, step S) string
	// length is how many runes step adds to a generated name.
	length(step S) int
	// write appends step's text to the builder.
	write(b *strings.Builder, step S)
}

// markov is the shared core of the Markov-chain namers. S is the unit a name is generated from one step at a time: a
// rune for MarkovLetterNamer or a vowel/consonant run for MarkovRunNamer. Everything that differs between the two
// namers lives in the markovStepper; the training and generation logic here is identical for both.
type markov[S cmp.Ordered] struct {
	stepper      markovStepper[S]
	mapping      map[string][]weightedStep[S]
	final        map[S]struct{}
	lengths      []weightedStep[int]
	maxLength    int
	lowered      bool
	firstToUpper bool
}

func newMarkov[S cmp.Ordered](stepper markovStepper[S], data iter.Seq2[string, int], lowered, firstToUpper bool) *markov[S] {
	n := &markov[S]{
		stepper:      stepper,
		final:        make(map[S]struct{}),
		lowered:      lowered,
		firstToUpper: firstToUpper,
	}
	mapping := make(map[string]map[S]int)
	lengths := make(map[int]int)
	for name, count := range data {
		if count > 0 {
			if name = strings.TrimSpace(name); name != "" {
				n.add(name, count, mapping, lengths)
			}
		}
	}
	n.lengths, n.maxLength = computeLengths(lengths)
	n.mapping = cumulativePairs(mapping, func(step S, cumulative int64) weightedStep[S] {
		return weightedStep[S]{step: step, last: cumulative}
	})
	return n
}

func (n *markov[S]) add(name string, count int, mapping map[string]map[S]int, lengths map[int]int) {
	steps := n.stepper.steps(name)
	if len(steps) == 0 {
		return
	}
	key := n.stepper.initialKey()
	for _, step := range steps {
		m, ok := mapping[key]
		if !ok {
			m = make(map[S]int)
			mapping[key] = m
		}
		m[step] = addWeight(m[step], count)
		key = n.stepper.advance(key, step)
	}
	n.final[steps[len(steps)-1]] = struct{}{}
	nameLen := utf8.RuneCountInString(name)
	lengths[nameLen] = addWeight(lengths[nameLen], count)
}

// GenerateName generates a new random name.
func (n *markov[S]) GenerateName() string {
	return generateName(n)
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *markov[S]) GenerateNameWithRandomizer(rnd xrand.Randomizer) string {
	var buffer strings.Builder
	maximum := selectMax(n.lengths, rnd)
	// Past 'maximum' the loop keeps going only to end on a natural (final) step. Training data whose transition graph
	// cycles without a reachable final step would otherwise loop forever, so cap the length at twice the longest
	// training name as a safety valve; legitimate data never reaches it.
	hardCap := 2 * n.maxLength
	key := n.stepper.initialKey()
	count := 0
	for {
		choices, ok := n.mapping[key]
		if !ok {
			break
		}
		picked, ok := pickWeighted(choices, rnd, func(ws weightedStep[S]) int64 { return ws.last })
		if !ok {
			break
		}
		key = n.stepper.advance(key, picked.step)
		n.stepper.write(&buffer, picked.step)
		count += n.stepper.length(picked.step)
		if count >= maximum {
			if _, final := n.final[picked.step]; final {
				break
			}
		}
		if count >= hardCap {
			break
		}
	}
	return applyCase(buffer.String(), n.lowered, n.firstToUpper)
}

func computeLengths(lengths map[int]int) (result []weightedStep[int], maxLength int) {
	// Reuse the shared cumulative-weight builder (which accumulates in int64 and in sorted key order, so a seeded
	// randomizer reproduces the same length selection across process runs) rather than duplicating that arithmetic.
	result = cumulativeWeights(lengths, func(length int, cumulative int64) weightedStep[int] {
		return weightedStep[int]{step: length, last: cumulative}
	})
	if n := len(result); n != 0 {
		maxLength = result[n-1].step // keys accumulate in ascending order, so the last entry holds the longest length
	}
	return result, maxLength
}

func selectMax(lengths []weightedStep[int], rnd xrand.Randomizer) int {
	if p, ok := pickWeighted(lengths, rnd, func(ws weightedStep[int]) int64 { return ws.last }); ok {
		return p.step
	}
	return 0
}
