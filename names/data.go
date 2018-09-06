package names

import (
	"bytes"
	"go/format"
	"sort"
	"text/template"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
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

// Data holds the data necessary to create a random name generator.
// If persistence is desired, this is the data that should be recorded. The
// codegen package can turn this into Go code.
type Data struct {
	StartsWithVowelFreq     int                  `json:"starts_with_vowel_freq" yaml:"starts_with_vowel_freq"`
	StartsWithConsonantFreq int                  `json:"starts_with_consonant_freq" yaml:"starts_with_consonant_freq"`
	CountFreq               []int                `json:"count_freq" yaml:"count_freq"`
	Segments                [ArraySize][]Segment `json:"segments"`
}

// Generator creates a new random name generator from this data.
func (data *Data) Generator() *Generator {
	g := &Generator{data: data.Clone()}
	for _, one := range data.CountFreq {
		g.countTotalFreq += one
	}
	for i, seg := range data.Segments {
		for j := range seg {
			g.segmentsTotalFreq[i] += seg[j].Freq
		}
	}
	return g
}

// Clone the data.
func (data *Data) Clone() *Data {
	g := &Data{
		StartsWithVowelFreq:     data.StartsWithVowelFreq,
		StartsWithConsonantFreq: data.StartsWithConsonantFreq,
		CountFreq:               make([]int, len(data.CountFreq)),
	}
	copy(g.CountFreq, data.CountFreq)
	for i, seg := range data.Segments {
		g.Segments[i] = make([]Segment, len(seg))
		copy(g.Segments[i], seg)
	}
	return g
}

// Code creates Go code for this data.
func (data *Data) Code(pkg, varName string) (string, error) {
	for i := range data.Segments {
		sort.Slice(data.Segments[i], func(j, k int) bool {
			return txt.NaturalLess(data.Segments[i][j].Value, data.Segments[i][k].Value, false)
		})
	}
	var buffer bytes.Buffer
	t, err := template.New("").Parse(`// Code generated - DO NOT EDIT.
package {{.Package}}

import "github.com/richardwilkes/rpgtools/names"

// {{.Name}} is a random name generator.
var {{.Name}} = (&names.Data{
	StartsWithVowelFreq: {{.Data.StartsWithVowelFreq}},
	StartsWithConsonantFreq: {{.Data.StartsWithConsonantFreq}},
	CountFreq: []int{ {{- range .Data.CountFreq}}{{.}},{{end}} },
	Segments: [names.ArraySize][]names.Segment{
		{{- range .Data.Segments}}
		{
			{{- range .}}
			{ Value: {{printf "%q" .Value}}, Freq: {{.Freq}} },
			{{- end}}
		},
		{{- end}}
	},
}).Generator()
`)
	if err != nil {
		return "", errs.Wrap(err)
	}
	if err = t.Execute(&buffer, &struct {
		Package string
		Name    string
		Data    *Data
	}{
		Package: pkg,
		Name:    varName,
		Data:    data,
	}); err != nil {
		return "", errs.Wrap(err)
	}
	src, err := format.Source(buffer.Bytes())
	if err != nil {
		return "", errs.Wrap(err)
	}
	return string(src), nil
}
