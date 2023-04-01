// Copyright ©2017-2023 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package american

import (
	_ "embed"
	"strings"

	"github.com/richardwilkes/rpgtools/names/namesets"
)

// Data for American first names obtained from http://www.ssa.gov/oact/babynames/names.zip
// Data for American last names obtained from https://www2.census.gov/topics/genealogy/2010surnames/names.zip

//go:embed female.txt
var female string

// Female returns a map of American female first names to frequency of occurrence.
func Female() map[string]int {
	return namesets.MustLoadFromReader(strings.NewReader(female))
}

//go:embed male.txt
var male string

// Male returns a map of American male first names to frequency of occurrence.
func Male() map[string]int {
	return namesets.MustLoadFromReader(strings.NewReader(male))
}

//go:embed last.txt
var last string

// Last returns a map of American last names to frequency of occurrence.
func Last() map[string]int {
	return namesets.MustLoadFromReader(strings.NewReader(last))
}
