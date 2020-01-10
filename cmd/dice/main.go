// Copyright ©2017-2020 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/richardwilkes/rpgtools/dice"
)

func main() {
	for _, arg := range os.Args[1:] {
		d := dice.New(nil, arg)
		fmt.Printf("%v = %s\n", d, humanize.Comma(int64(d.Roll())))
	}
}
