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
// namers lives in the markovStepper; the training and generation logic here is identical for both. The namers embed it
// by value so that their zero values are usable: a zero markov has no stepper and no training data, and generating from
// it yields "" rather than a nil pointer dereference, matching the zero SimpleNamer and CompoundNamer.
//
// mapping holds every successor of each key. endings holds, for each key, only the successors that lead to a final step
// soonest (see buildEndings): once a generated name has reached its target length the next step is drawn from endings
// whenever the key has an entry there, so the name wraps up at the first natural opportunity instead of wandering on
// until a final step happens to be picked.
type markov[S cmp.Ordered] struct {
	stepper      markovStepper[S]
	mapping      map[string][]weightedStep[S]
	endings      map[string][]weightedStep[S]
	final        map[S]struct{}
	lengths      []weightedStep[int]
	maxLength    int
	lowered      bool
	firstToUpper bool
}

func newMarkov[S cmp.Ordered](stepper markovStepper[S], data iter.Seq2[string, int], lowered, firstToUpper bool) markov[S] {
	n := markov[S]{
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
	makePair := func(step S, cumulative int64) weightedStep[S] {
		return weightedStep[S]{step: step, last: cumulative}
	}
	n.mapping = cumulativePairs(mapping, makePair)
	n.endings = cumulativePairs(n.buildEndings(mapping), makePair)
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

// buildEndings returns, for every key from which a final step is reachable, the successors that reach one soonest: the
// final successors themselves when the key has any, otherwise the successors whose next key is the fewest steps from
// a final one. The distances come from a breadth-first search run backwards from the keys with a final successor. Every
// key produced by training can reach a final step, because the remainder of the training name that created it is
// itself a path to one, so with real data every key gets an entry and a generated name that has reached its target
// length ends within at most one training name's length of further steps. Only a hand-built table whose graph never
// reaches a final step is left without entries. It must run after every training name has been added, since a step
// only becomes final once some name has ended with it.
func (n *markov[S]) buildEndings(mapping map[string]map[S]int) map[string]map[S]int {
	type edge struct {
		step S
		key  string
	}
	endings := make(map[string]map[S]int)
	record := func(key string, step S) {
		m, ok := endings[key]
		if !ok {
			m = make(map[S]int)
			endings[key] = m
		}
		m[step] = mapping[key][step]
	}
	// distance is the fewest steps from a key to emitting a final step; keys with a final successor seed the search at
	// 1. predecessors maps each key to the non-final edges that lead into it, which the search walks backwards.
	distance := make(map[string]int, len(mapping))
	predecessors := make(map[string][]edge)
	var queue []string
	for key, counts := range mapping {
		for step := range counts {
			if _, ok := n.final[step]; ok {
				if _, seen := distance[key]; !seen {
					distance[key] = 1
					queue = append(queue, key)
				}
				record(key, step)
				continue
			}
			next := n.stepper.advance(key, step)
			predecessors[next] = append(predecessors[next], edge{key: key, step: step})
		}
	}
	// The queue is processed in non-decreasing distance order, so the first time an edge out of a key is examined it
	// sets that key's true distance, and any later edge whose target sits one step closer is another shortest route.
	// The resulting table is therefore the same whatever order the maps above were iterated in.
	for i := 0; i < len(queue); i++ {
		key := queue[i]
		d := distance[key] + 1
		for _, e := range predecessors[key] {
			existing, seen := distance[e.key]
			if !seen {
				distance[e.key] = d
				queue = append(queue, e.key)
			}
			if !seen || existing == d {
				record(e.key, e.step)
			}
		}
	}
	return endings
}

// GenerateName generates a new random name.
func (n *markov[S]) GenerateName() string {
	return generateName(n)
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *markov[S]) GenerateNameWithRandomizer(rnd xrand.Randomizer) string {
	// A zero-value namer (one not built by a constructor) has no stepper and no training data, so there is nothing to
	// generate. Return "" as SimpleNamer and CompoundNamer do rather than dereferencing the nil stepper.
	if n.stepper == nil {
		return ""
	}
	var buffer strings.Builder
	maximum := selectMax(n.lengths, rnd)
	// Past 'maximum' the loop keeps going only to end on a natural (final) step, drawing from the endings table so that
	// it heads for the nearest one: with real training data that is never more than one training name's length away,
	// so the result stays within twice the longest training name. Training data whose transition graph cycles without a
	// reachable final step (which no trained table has, but a hand-built one can) would otherwise loop forever, so the
	// length is also capped there as a safety valve. Letter steps are single runes, so a letter namer never reaches the
	// cap with real data; a run namer takes whole runs, so in principle a contrived corpus could still be cut off at it.
	hardCap := 2 * n.maxLength
	key := n.stepper.initialKey()
	count := 0
	for {
		choices, ok := n.mapping[key]
		if !ok {
			break
		}
		if count >= maximum {
			if endings, found := n.endings[key]; found {
				choices = endings
			}
		}
		picked, ok := pickWeighted(choices, rnd, func(ws weightedStep[S]) int64 { return ws.last })
		if !ok {
			break
		}
		// Check the cap before appending rather than after so that a multi-rune step (a whole vowel or consonant run)
		// can never carry the name past it.
		if count += n.stepper.length(picked.step); count > hardCap {
			break
		}
		key = n.stepper.advance(key, picked.step)
		n.stepper.write(&buffer, picked.step)
		if count >= maximum {
			if _, final := n.final[picked.step]; final {
				break
			}
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
