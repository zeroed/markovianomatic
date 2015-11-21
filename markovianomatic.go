package markovianomatic

// TODO: save and load initial dictionary
// DONE: keep list on values (do not lose the weight)
// DONE: tabulation in printing
// TODO: more filters
// TODO: check correctness

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uitable"
	"github.com/zeroed/markovianomatic/model"
)

type StringMap map[string][]string

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     StringMap
	starters  []string
	prefixLen int
	verbose   bool
	lock      *sync.RWMutex
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int, verbose bool) *Chain {
	c := new(Chain)
	c.chain = make(StringMap)
	c.prefixLen = prefixLen
	c.verbose = verbose
	c.lock = &sync.RWMutex{}
	return c
}

// Keys return the list of keys in the chain
func (c *Chain) Keys() []string {
	var keys []string
	for k := range c.chain {
		keys = append(keys, k)
	}
	return keys
}

// Rkey return a random key from the chain starters
func (c *Chain) RKey() string {
	if l := len(c.starters); l > 0 {
		return c.starters[rand.Intn(l)]
	} else {
		return ""
	}
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

// Pretty is a pretty print of a chain
func (c *Chain) Pretty() {
	ks := c.Keys()
	sort.Strings(ks)
	fmt.Fprintf(os.Stdout, "--------------\n")

	table := uitable.New()
	table.MaxColWidth = 50

	table.AddRow("Index", "Prefix key", "Available choices")
	for i, x := range ks {
		table.AddRow(
			fmt.Sprintf("\033[0;37m %03d \033[0m", i),
			fmt.Sprintf("\033[0;32m %s \033[0m", x),
			fmt.Sprintf("\033[0;33m %s \033[0m", c.chain[x]))
	}

	fmt.Fprintf(os.Stdout, table.String())
	fmt.Fprintf(os.Stdout, "\n--------------\n")
}

// Prefix returns a value corresponding to a given prefix.
func (c *Chain) Prefix(k string) []string {
	return c.chain[k]
}

// Prefixes return the list of all the prefixes in the chain
func (c *Chain) Prefixes() (pxs []string) {
	k := c.Keys()
	sort.Strings(k)
	return k
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r io.Reader) {
	fmt.Fprintf(os.Stdout, "-- Markovianomatic live -- \n\ntype your text (Enter x2 to stop)...\n\n")

	pf := make(Prefix, c.prefixLen)
	qc := make(chan bool, 1)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	go func() {
		for _ = range sc {
			if c.verbose {
				fmt.Fprintf(os.Stdout, "\nReceived an interrupt, stopping services...\n")
			}
			qc <- true
			break
		}
	}()

	go func() {
		i := 0
		var s string
		br := bufio.NewReader(r)

		for {
			if n, err := fmt.Fscanf(br, "%s\n", &s); n > 0 {
				i = 0
				c.insert(s, &pf)
			} else {
				i += 1
				if err != nil && err.Error() != "unexpected newline" {
					fmt.Fprintf(os.Stderr, "Scan error: %s\n", err.Error())
				} else {
					if c.verbose {
						fmt.Fprintf(os.Stderr, "Write empty (%d)\n", i)
					}
				}
			}

			if i == 2 {
				qc <- true
				break
			}
		}
	}()

	<-qc
	return
}

func (c *Chain) insert(s string, p *Prefix) {
	key := p.String()
	s = sanitise(s)

	c.lock.Lock()
	if c.verbose {
		fmt.Fprintf(os.Stdout, "Association: |%s| -> [%s]\n", key, s)
	}
	if key != " " {
		c.chain[key] = append(c.chain[key], s)
	}
	if strings.HasPrefix(key, " ") && regexp.MustCompile(`\s\w`).MatchString(key) == true {
		if c.verbose {
			fmt.Fprintf(os.Stdout, "New starter: [%s]\n", key)
		}
		c.starters = append(c.starters, key)
	}
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
		for _, w := range strings.Fields(scanner.Text()) {
			c.insert(w, &p)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func downcase(s *string) {
	*s = strings.ToLower(*s)
}

func isCapital(s string) bool {
	if string(s[0]) == strings.ToLower(string(s[0])) {
		return false
	}
	return true
}

func hasEnding(s *string) bool {
	return strings.Contains(*s, ".")
}

func sanitise(s string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9éèàìòù]+")
	if err != nil {
		log.Fatal(err)
	}

	s = reg.ReplaceAllString(s, "")
	s = strings.ToLower(strings.Trim(s, "-"))
	return s
}

// Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(w io.Writer, n int) io.Writer {
	c.Pretty()

	if n < 1 {
		panic("Prefix too short")
	} else {
		fmt.Fprintf(os.Stdout, "%d prefixes, prefixes %d long. generating text ...\n", c.Length(), c.prefixLen)
	}

	p := NewPrefix(c.prefixLen)

	var k string
	var choices []string
	k = c.RKey()
	for i := 0; i < n; i++ {
		choices = c.chain[k]

		if c.verbose {
			fmt.Fprintf(os.Stdout, "Current key: [%s], choices: %s\n", k, choices)
			fmt.Fprintf(os.Stdout, "%02d iteration: prefix [%s]\n", i, p.String())
		}

		if len(choices) == 0 {
			fmt.Printf("No more choices!\n")
			break
		}

		next := choices[rand.Intn(len(choices))]

		if c.verbose {
			fmt.Printf("Next: [%s]\n", next)
		}
		w.Write([]byte(next))
		w.Write([]byte(" "))

		if i == 0 {
			p.Shift(k)
		}
		p.Shift(next)
		k = p.String()
	}
	return w
}
