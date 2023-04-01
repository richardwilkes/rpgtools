// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package names

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

// Generator provides a random name generator.
type Generator struct {
	data              *Data
	countTotalFreq    int
	segmentsTotalFreq [ArraySize]int
}

// NewFromSamples uses the provided sample names to produce a new random name generator. Each sample should be a single
// word. 'vowelChecker' may be nil, in which case IsVowely() will be used. After processing all of the samples, any
// segments that don't meet the 'minimumFrequency' parameter will be pruned.
func NewFromSamples(samples []string, minimumFrequency int, vowelChecker txt.VowelChecker) *Generator {
	if vowelChecker == nil {
		vowelChecker = txt.IsVowely
	}
	g := &Generator{data: &Data{}}
	for _, name := range samples {
		name = strings.TrimSpace(strings.ToLower(name))
		if name != "" {
			var segment string
			var last rune
			var index, count int
			for _, ch := range name {
				if segment != "" {
					if vowelChecker(ch) != vowelChecker(last) {
						g.process(index, segment, vowelChecker)
						segment = ""
						if index == Initial {
							index = Interior
						}
						count++
					}
				}
				segment += string(ch)
				last = ch
			}
			if index != Initial {
				index = Ending
			}
			g.process(index, segment, vowelChecker)
			count++
			segCount := count
			if segCount < 2 {
				segCount = 2 // Force a minimum of 2 segments
			}
			if len(g.data.CountFreq) < segCount {
				countFreq := make([]int, segCount)
				copy(countFreq, g.data.CountFreq)
				g.data.CountFreq = countFreq
			}
			g.data.CountFreq[segCount-1]++
			g.countTotalFreq++
		}
	}
	for i := range g.data.Segments {
		segments := make([]Segment, 0, len(g.data.Segments[i]))
		for j := range g.data.Segments[i] {
			if g.data.Segments[i][j].Freq >= minimumFrequency {
				segments = append(segments, g.data.Segments[i][j])
			} else {
				g.segmentsTotalFreq[i] -= g.data.Segments[i][j].Freq
			}
		}
		sort.Slice(segments, func(j, k int) bool {
			return txt.NaturalLess(segments[j].Value, segments[k].Value, false)
		})
		g.data.Segments[i] = make([]Segment, len(segments))
		copy(g.data.Segments[i], segments)
	}
	return g
}

func (g *Generator) process(index int, segment string, vowelChecker txt.VowelChecker) {
	r, _ := utf8.DecodeRuneInString(segment)
	if r == utf8.RuneError {
		return
	}
	if vowelChecker(r) {
		if index == InitialVowel {
			g.data.StartsWithVowelFreq++
		}
	} else {
		index += VowelToConsonant
		if index == InitialConsonant {
			g.data.StartsWithConsonantFreq++
		}
	}
	g.segmentsTotalFreq[index]++
	for i := range g.data.Segments[index] {
		if g.data.Segments[index][i].Value == segment {
			g.data.Segments[index][i].Freq++
			return
		}
	}
	g.data.Segments[index] = append(g.data.Segments[index], Segment{
		Value: segment,
		Freq:  1,
	})
}

// Generate a new random name.
func (g *Generator) Generate() string {
	return g.GenerateWith(rand.NewCryptoRand())
}

// GenerateWith generates a new random name using the specified randomizer.
func (g *Generator) GenerateWith(rnd rand.Randomizer) string {
	r := rnd.Intn(g.countTotalFreq)
	var index int
	for i, freq := range g.data.CountFreq {
		if r < freq {
			index = i
			break
		}
		r -= freq
	}
	useVowel := rnd.Intn(g.data.StartsWithVowelFreq+g.data.StartsWithConsonantFreq) < g.data.StartsWithVowelFreq
	var buffer strings.Builder
	for i := 0; i <= index; i++ {
		var which int
		switch i {
		case 0:
			which = Initial
		case index:
			which = Ending
		default:
			which = Interior
		}
		if !useVowel {
			which += VowelToConsonant
		}
		useVowel = !useVowel
		total := g.segmentsTotalFreq[which]
		if total > 0 {
			buffer.WriteString(PickSegmentValue(rnd, total, g.data.Segments[which]))
		}
	}
	return txt.FirstToUpper(buffer.String())
}

// CloneData returns a clone of the data being used for this generator.
func (g *Generator) CloneData() *Data {
	return g.data.Clone()
}
