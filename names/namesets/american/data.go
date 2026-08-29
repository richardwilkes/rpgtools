// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
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
	"maps"
	"strings"
	"sync"

	"github.com/richardwilkes/rpgtools/names/namesets"
)

// Data for American first names obtained from http://www.ssa.gov/oact/babynames/names.zip
// Data for American last names obtained from https://www2.census.gov/topics/genealogy/2010surnames/names.zip

//go:embed female.txt
var female string

// The embedded corpora are large, so each is parsed at most once and the parsed map is cached. Callers receive a fresh
// clone of that map so they retain the prior contract of owning a map they may freely mutate without affecting other
// callers or the cache. That clone is not free: it copies tens of thousands of entries on every call (several
// megabytes and a fraction of a millisecond), which each accessor's doc warns about.
var (
	femaleOnce = sync.OnceValue(func() map[string]int {
		return namesets.MustLoadFromReader(strings.NewReader(female))
	})
	maleOnce = sync.OnceValue(func() map[string]int {
		return namesets.MustLoadFromReader(strings.NewReader(male))
	})
	lastOnce = sync.OnceValue(func() map[string]int {
		return namesets.MustLoadFromReader(strings.NewReader(last))
	})
)

// Female returns a map of American female first names to frequency of occurrence. Each call clones the cached corpus
// (tens of thousands of entries), so call it once and reuse the result rather than calling it inside a loop.
func Female() map[string]int {
	return maps.Clone(femaleOnce())
}

//go:embed male.txt
var male string

// Male returns a map of American male first names to frequency of occurrence. Each call clones the cached corpus (tens
// of thousands of entries), so call it once and reuse the result rather than calling it inside a loop.
func Male() map[string]int {
	return maps.Clone(maleOnce())
}

//go:embed last.txt
var last string

// Last returns a map of American last names to frequency of occurrence. Each call clones the cached corpus (over a
// hundred thousand entries), so call it once and reuse the result rather than calling it inside a loop.
func Last() map[string]int {
	return maps.Clone(lastOnce())
}
