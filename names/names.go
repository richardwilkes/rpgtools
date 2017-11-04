package names

import (
	"crypto/rand"
	"math/big"
	"time"
)

// First returns a random first name for the specified gender.
func First(forMale bool) string {
	if forMale {
		return choose(male)
	}
	return choose(female)
}

// Last returns a random last name.
func Last() string {
	return choose(last)
}

// Full returns a random first and last name for the specified gender.
func Full(forMale bool) string {
	return First(forMale) + " " + Last()
}

func choose(names []string) string {
	if n, err := rand.Int(rand.Reader, big.NewInt(int64(len(names)))); err == nil {
		return names[n.Int64()]
	}
	now := time.Now().UnixNano()
	if now < 0 {
		now = -now
	}
	return names[now%int64(len(names))]
}
