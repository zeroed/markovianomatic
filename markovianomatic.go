package markovianomatic

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
)

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     map[string][]string
	prefixLen int
	lock      *sync.RWMutex
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen, &sync.RWMutex{}}
}

// String return a printable representation of the Chain
func (c *Chain) String() string {
	var keys []string
	for k := range c.chain {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	s := []string{}
	for _, k := range keys {

		s = append(s, fmt.Sprintf("[%s]: %s", k, strings.Join(c.chain[k], ", ")))
	}
	return strings.Join(s, "\n")
}

// Length returns the length of the chain.
func (c *Chain) Length() int {
	return len(c.chain)
}

// Prefix returns a value corresponding to a given prefix.
func (c *Chain) Prefix(k string) []string {
	return c.chain[k]
}

// Prefixes return the list of all the prefixes in the chain
func (c *Chain) Prefixes() (pxs []string) {
	var keys []string
	for k := range c.chain {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	p := make(Prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}
		c.insert(s, &p)
	}
}

func (c *Chain) insert(s string, p *Prefix) {
	s = strings.ToLower(s)
	key := p.String()
	c.lock.Lock()
	c.chain[key] = append(c.chain[key], s)
	p.Shift(s)
	c.lock.Unlock()
}

//Load reads the content from a file and build a Chain
func (c *Chain) Load(name string) {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	p := make(Prefix, c.prefixLen)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		c.insert(scanner.Text(), &p)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// Generate returns a string of at most n words generated from Chain.
// The Generate function does a lot of allocations when it builds the words
// slice. As an exercise, modify it to take an io.Writer to which it
// incrementally writes the generated text with Fprint. Aside from being more
// efficient this makes Generate more symmetrical to Build.
func (c *Chain) Generate(n int) string {
	if n < 1 {
		panic("Refix too short")
	}
	p := make(Prefix, c.prefixLen)
	var words []string
	for i := 0; i < n; i++ {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			break
		}
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		p.Shift(next)
	}
	return strings.Join(words, " ")
}
