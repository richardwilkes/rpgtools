package names

import "github.com/richardwilkes/toolbox/xmath/rand"

// Segment holds string segment and its frequency of occurrence.
type Segment struct {
	Value string `json:"value" yaml:"value"`
	Freq  int    `json:"freq" yaml:"freq"`
}

// PickSegmentValue picks a value from the segment slice.
func PickSegmentValue(rnd rand.Randomizer, total int, segments []Segment) string {
	r := rnd.Intn(total)
	for i := range segments {
		if r < segments[i].Freq {
			return segments[i].Value
		}
		r -= segments[i].Freq
	}
	return segments[0].Value // Should never reach this line
}
