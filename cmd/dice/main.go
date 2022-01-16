// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
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
	"time"

	"github.com/dustin/go-humanize"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/i18n"
)

func main() {
	if cmdline.License == "" {
		cmdline.License = "Mozilla Public License, version 2.0"
	}
	if cmdline.CopyrightYears == "" {
		cmdline.CopyrightYears = fmt.Sprintf("2017-%d", time.Now().Year())
	}
	if cmdline.CopyrightHolder == "" {
		cmdline.CopyrightHolder = "Richard A. Wilkes"
	}
	cl := cmdline.New(true)
	cl.UsageSuffix = i18n.Text("[dice expression]...")
	cl.Parse(os.Args[1:])
	for _, arg := range os.Args[1:] {
		d := dice.New(arg)
		fmt.Printf("%v = %s\n", d, humanize.Comma(int64(d.Roll(false))))
	}
}
