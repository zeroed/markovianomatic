package markovianomatic

import "strings"

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

func NewPrefix(l int) Prefix {
	if l < 2 {
		return make(Prefix, 2)
	}
	return make(Prefix, l)
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = strings.TrimSpace(word)
}

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}
