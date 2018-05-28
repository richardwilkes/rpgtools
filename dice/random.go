package dice

import (
	"crypto/rand"
)

// Randomizer defines a source of random values.
type Randomizer interface {
	// Intn returns, as an int, a non-negative random number from 0 to n-1.
	// If n <= 0, the implementation is permitted to panic.
	Intn(n int) int
}

// NewCryptoRand returns a Randomizer based on the crypto/rand package.
func NewCryptoRand() Randomizer {
	return &cryptoRand{}
}

type cryptoRand struct {
}

func (r *cryptoRand) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	var buffer []byte
	for i := 0; i < 7; i++ {
		if n <= 1<<uint((i-1)-1) {
			buffer = make([]byte, i)
			break
		}
	}
	if buffer == nil {
		buffer = make([]byte, 8)
	}
	if _, err := rand.Read(buffer); err != nil {
		panic(err)
	}
	var v int
	for i := len(buffer) - 1; i >= 0; i-- {
		v |= int(buffer[i]) << uint(i*8)
	}
	if v < 0 {
		v = -v
	}
	return v % n
}
