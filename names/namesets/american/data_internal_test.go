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
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestEmbeddedCorporaLoadEveryLine(t *testing.T) {
	c := check.New(t)
	// Each embedded file holds one "name,count" line per distinct name, with no blank lines and no repeated or
	// non-positive entries, so the parsed map must hold exactly one entry per non-blank line. A loader regression that
	// silently dropped lines would slip past a bare non-empty check; this pins the corpora at their full size (tens of
	// thousands of first names and over a hundred thousand surnames).
	for name, tc := range map[string]struct {
		load func() map[string]int
		text string
		min  int
	}{
		"female": {femaleOnce, female, 60_000},
		"male":   {maleOnce, male, 40_000},
		"last":   {lastOnce, last, 150_000},
	} {
		lines := 0
		for line := range strings.Lines(tc.text) {
			if strings.TrimSpace(line) != "" {
				lines++
			}
		}
		c.True(lines >= tc.min, "%s: expected at least %d lines in the embedded file, found %d", name, tc.min, lines)
		c.Equal(lines, len(tc.load()), "%s: every line must load as its own name", name)
	}
}
