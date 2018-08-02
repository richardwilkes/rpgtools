package codegen

import (
	"bytes"
	"go/format"
	"sort"
	"text/template"

	"github.com/richardwilkes/rpgtools/names"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
)

type info struct {
	Package string
	Name    string
	Data    *names.GeneratorData
}

// CreateGoCode creates Go code for the provided names.GeneratorData.
func CreateGoCode(pkg, varName string, data *names.GeneratorData) (string, error) {
	for i := range data.Segments {
		sort.Slice(data.Segments[i], func(j, k int) bool {
			return txt.NaturalLess(data.Segments[i][j].Value, data.Segments[i][k].Value, false)
		})
	}
	var buffer bytes.Buffer
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", errs.Wrap(err)
	}
	if err = t.Execute(&buffer, &info{
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

var tmpl = `// Code generated - DO NOT EDIT.
package {{.Package}}

import "github.com/richardwilkes/rpgtools/names"

var {{.Name}} = names.NewFromData(&names.GeneratorData{
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
})
`
