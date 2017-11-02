package main

import (
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/richardwilkes/rpgtools/dice"
)

func main() {
	for _, arg := range os.Args[1:] {
		d := dice.New(arg)
		fmt.Printf("%v = %s\n", d, humanize.Comma(int64(d.Roll())))
	}
}
