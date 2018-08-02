package names

// VowelChecker defines a function that returns true if the specified rune is
// to be considered a vowel.
type VowelChecker func(rune) bool

// IsVowel is a concrete implementation of VowelChecker.
func IsVowel(ch rune) bool {
	return ch == 'a' ||
		ch == 'e' ||
		ch == 'i' ||
		ch == 'o' ||
		ch == 'u'
}

// IsVowely is a concrete implementation of VowelChecker that includes 'y'.
func IsVowely(ch rune) bool {
	return ch == 'y' || IsVowel(ch)
}
