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

// weightedStep pairs a step (for SimpleNamer, a whole name) with the cumulative weight through it, the form
// pickWeighted consumes.
type weightedStep[S cmp.Ordered] struct {
	step S
	last int64
}

// markovStepper supplies the per-step behavior that distinguishes the Markov namers.
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

// markov is the shared core of the Markov-chain namers, generating a name one step (S) at a time: a rune for
// MarkovLetterNamer or a vowel/consonant run for MarkovRunNamer. The namers embed it by value so their zero values are
// usable and generate "".
//
// mapping holds every successor of each key. endings holds, for each key, only the successors that lead to a final step
// soonest (see buildEndings); once a generated name reaches its target length the next step is drawn from there.
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
// final successors themselves when the key has any, otherwise those whose next key is fewest steps from a final one,
// found by a breadth-first search run backwards from the keys with a final successor. Every key produced by training
// can reach a final step (the rest of the training name is such a path), so only a hand-built table can lack entries.
// It must run after every training name has been added, since a step only becomes final once some name ends with it.
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
	// distance is the fewest steps from a key to a final step; keys with a final successor seed the search at 1.
	// predecessors maps each key to the non-final edges leading into it.
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
	// The queue is in non-decreasing distance order, so the first edge examined out of a key sets its true distance and
	// any later edge at the same distance is another shortest route, whatever order the maps were iterated in.
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
	if n.stepper == nil { // zero-value namer
		return ""
	}
	var buffer strings.Builder
	maximum := selectMax(n.lengths, rnd)
	// Past 'maximum' the loop continues only to end on a final step, which the endings table keeps within one training
	// name's length. The cap is a safety valve for a hand-built table whose graph never reaches a final step.
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
		// Checked before appending so a multi-rune step cannot carry the name past the cap.
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
	result = cumulativeWeights(lengths, func(length int, cumulative int64) weightedStep[int] {
		return weightedStep[int]{step: length, last: cumulative}
	})
	if n := len(result); n != 0 {
		maxLength = result[n-1].step // sorted ascending
	}
	return result, maxLength
}

func selectMax(lengths []weightedStep[int], rnd xrand.Randomizer) int {
	if p, ok := pickWeighted(lengths, rnd, func(ws weightedStep[int]) int64 { return ws.last }); ok {
		return p.step
	}
	return 0
}
