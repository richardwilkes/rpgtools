package names

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

// Constants for naming the segment array indexes.
const (
	Initial = iota
	Interior
	Ending
	InitialVowel      = Initial
	InteriorVowel     = InitialVowel + Interior
	EndingVowel       = InitialVowel + Ending
	InitialConsonant  = EndingVowel + 1
	InteriorConsonant = InitialConsonant + Interior
	EndingConsonant   = InitialConsonant + Ending
	VowelToConsonant  = InitialConsonant - InitialVowel
	ArraySize         = EndingConsonant + 1
)

// Segment holds string segment and its frequency of occurrence.
type Segment struct {
	Value string `json:"value" yaml:"value"`
	Freq  int    `json:"freq" yaml:"freq"`
}

// GeneratorData holds the data necessary to create a random name generator.
// If persistence is desired, this is the data that should be recorded. The
// codegen package can turn this into Go code.
type GeneratorData struct {
	StartsWithVowelFreq     int                  `json:"starts_with_vowel_freq" yaml:"starts_with_vowel_freq"`
	StartsWithConsonantFreq int                  `json:"starts_with_consonant_freq" yaml:"starts_with_consonant_freq"`
	CountFreq               []int                `json:"count_freq" yaml:"count_freq"`
	Segments                [ArraySize][]Segment `json:"segments" yaml:"segments"`
}

// Generator provides a random name generator.
type Generator struct {
	Randomizer        rand.Randomizer
	CountTotalFreq    int
	SegmentsTotalFreq [ArraySize]int
	GeneratorData
}

// NewFromData creates a new random name generator from configuration data.
func NewFromData(data *GeneratorData) *Generator {
	// Manually copy the data over, to ensure no shared pointers are retained
	g := &Generator{
		Randomizer: rand.NewCryptoRand(),
		GeneratorData: GeneratorData{
			StartsWithVowelFreq:     data.StartsWithVowelFreq,
			StartsWithConsonantFreq: data.StartsWithConsonantFreq,
			CountFreq:               make([]int, len(data.CountFreq)),
		},
	}
	copy(g.CountFreq, data.CountFreq)
	for _, one := range data.CountFreq {
		g.CountTotalFreq += one
	}
	for i, seg := range data.Segments {
		g.Segments[i] = make([]Segment, len(seg))
		copy(g.Segments[i], seg)
		for j := range seg {
			g.SegmentsTotalFreq[i] += seg[j].Freq
		}
	}
	return g
}

// NewFromSamples uses the provided sample names to produce a new random name
// generator. Each sample should be a single word. 'vowelChecker' may be nil,
// in which case IsVowely() will be used.
func NewFromSamples(samples []string, vowelChecker VowelChecker) *Generator {
	if vowelChecker == nil {
		vowelChecker = IsVowely
	}
	g := &Generator{Randomizer: rand.NewCryptoRand()}
	for _, name := range samples {
		if name != "" {
			var segment string
			var last rune
			var index int
			var count int
			for _, ch := range strings.ToLower(name) {
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
			if len(g.CountFreq) < segCount {
				countFreq := make([]int, segCount)
				copy(countFreq, g.CountFreq)
				g.CountFreq = countFreq
			}
			g.CountFreq[segCount-1]++
			g.CountTotalFreq++
		}
	}
	for i := range g.Segments {
		sort.Slice(g.Segments[i], func(j, k int) bool {
			return txt.NaturalLess(g.Segments[i][j].Value, g.Segments[i][k].Value, false)
		})
	}
	return g
}

func (g *Generator) process(index int, segment string, vowelChecker VowelChecker) {
	r, _ := utf8.DecodeRuneInString(segment)
	if r == utf8.RuneError {
		return
	}
	if vowelChecker(r) {
		if index == InitialVowel {
			g.StartsWithVowelFreq++
		}
	} else {
		index += VowelToConsonant
		if index == InitialConsonant {
			g.StartsWithConsonantFreq++
		}
	}
	g.SegmentsTotalFreq[index]++
	for i := range g.Segments[index] {
		if g.Segments[index][i].Value == segment {
			g.Segments[index][i].Freq++
			return
		}
	}
	g.Segments[index] = append(g.Segments[index], Segment{
		Value: segment,
		Freq:  1,
	})
}

// Prune edits the generator data to remove any segments that have a frequency
// less than the specified value.
func (g *Generator) Prune(minimumFrequency int) {
	for i := range g.Segments {
		segments := make([]Segment, 0, len(g.Segments[i]))
		for j := range g.Segments[i] {
			if g.Segments[i][j].Freq >= minimumFrequency {
				segments = append(segments, g.Segments[i][j])
			} else {
				g.SegmentsTotalFreq[i] -= g.Segments[i][j].Freq
			}
		}
		g.Segments[i] = make([]Segment, len(segments))
		copy(g.Segments[i], segments)
	}
}

// Generate a new random name.
func (g *Generator) Generate() string {
	r := g.Randomizer.Intn(g.CountTotalFreq)
	var index int
	for i, freq := range g.CountFreq {
		if r < freq {
			index = i
			break
		}
		r -= freq
	}
	useVowel := g.Randomizer.Intn(g.StartsWithVowelFreq+g.StartsWithConsonantFreq) < g.StartsWithVowelFreq
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
		total := g.SegmentsTotalFreq[which]
		if total > 0 {
			var target int
			r = g.Randomizer.Intn(total)
			for j, one := range g.Segments[which] {
				if r < one.Freq {
					target = j
					break
				}
				r -= one.Freq
			}
			buffer.WriteString(g.Segments[which][target].Value)
		}
	}
	return strings.Title(buffer.String())
}
