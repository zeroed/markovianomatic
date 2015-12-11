package markovianomatic

// DONE: save and load initial dictionary
// TODO: parallelization on Save
// DONE: keep list on values (do not lose the weight)
// DONE: tabulation in printing
// TODO: more filters
// TODO: check correctness
// TODO: table menu: load/new/destroy

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

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uitable"
	"github.com/zeroed/markovianomatic/model"
)

// StringMap is a short name to address a type of "node" where a string
// is a key for a list of choices
type StringMap map[string][]string

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain      StringMap
	starters   []string
	prefixLen  int
	verbose    bool
	collection string
	lock       *sync.RWMutex
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int, verbose bool, cn string) *Chain {
	c := new(Chain)
	c.chain = make(StringMap)
	c.prefixLen = prefixLen
	c.verbose = verbose
	c.collection = sanitise(cn, true)
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

// RKey return a random key from the chain starters
func (c *Chain) RKey() string {
	if l := len(c.starters); l > 0 {
		return c.starters[rand.Intn(l)]
	}
	return ""
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

// Pretty is a pretty print of a chain
func (c *Chain) Pretty() {
	ks := c.Keys()
	sort.Strings(ks)
	fmt.Fprintf(os.Stdout, "--------------\n")

	table := uitable.New()
	table.MaxColWidth = 50
	table.Separator = "|"

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

// func (c *Chain) Restore(sm StringMap) {
// }

func (c *Chain) append(k string, s string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if k != " " {
		c.chain[k] = append(c.chain[k], s)
	}
	if strings.HasPrefix(k, " ") && regexp.MustCompile(`\s\w`).MatchString(k) == true {
		if c.verbose {
			fmt.Fprintf(os.Stdout, "New starter: [%s]\n", k)
		}
		c.starters = append(c.starters, k)
	}
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
				i++
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

func timeName() string {
	t := time.Now().UTC()
	s := fmt.Sprintf("%d%02d%02dT%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return fmt.Sprintf("dict_%s", s)
}

// Save method persist (save or update) the chain on the disk
func (c *Chain) Save() {
	var cn = c.collection
	if c.collection == "" {
		cn = timeName()
		c.collection = cn
	}

	sess, coll := model.Connect(cn)
	defer sess.Close()

	ks := c.Keys()
	sort.Strings(ks)

	var wg sync.WaitGroup
	var workForce = 5
	ch := make(chan model.NewNodeInfo, workForce)

	for i := 0; i < workForce; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var ni model.NewNodeInfo
			var more bool

			for {
				ni, more = <-ch
				if more {
					model.NewNode(ni).Save(coll)
				} else {
					return
				}
			}
		}()
	}

	count := len(ks)
	bar := uiprogress.AddBar(count)
	bar.AppendCompleted()
	bar.PrependElapsed()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Node (%d/%d)", b.Current(), count)
	})

	uiprogress.Start()
	for _, x := range ks {
		ch <- model.NewNodeInfo{x, c.chain[x]}

		bar.Incr()
	}

	uiprogress.Stop()
	close(ch)
	wg.Wait()
}

func (c *Chain) insert(s string, p *Prefix) {
	k := p.String()
	s = sanitise(s, false)

	if c.verbose {
		fmt.Fprintf(os.Stdout, "\033[0;34m Association: |%s| -> [%s]\033[0m\n", k, s)
	}

	c.append(k, s)
	p.Shift(s)
}

// Length returns the length of the chain.
func (c *Chain) Length() int {
	return len(c.chain)
}

// Load reads the content from a file and build a Chain
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

func sanitise(s string, strict bool) string {
	var reg *regexp.Regexp
	var err error

	if strict {
		reg, err = regexp.Compile("[^A-Za-z0-9]+")
	} else {
		reg, err = regexp.Compile("[^A-Za-z0-9éèàìòù]+")
	}
	if err != nil {
		log.Fatal(err)
	}

	s = reg.ReplaceAllString(s, "")
	s = strings.ToLower(strings.Trim(s, "-"))
	return s
}

// Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(w io.Writer, n int) io.Writer {
	if len(c.Keys()) < 0 {
		fmt.Fprintf(os.Stdout, "Empty text map. Cannot generate text\n")
		return w
	}

	if n < 1 {
		panic("Prefix too short")
	} else {
		fmt.Fprintf(os.Stdout, "%d prefixes, prefixes %d long. generating text ...\n\n", c.Length(), c.prefixLen)
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

// Set adds a pair "key choices" into the chain
func (c *Chain) Set(k string, v []string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if k != " " {
		c.chain[k] = v
	}
	// TODO: centralize starters
	if strings.HasPrefix(k, " ") && regexp.MustCompile(`\s\w`).MatchString(k) == true {
		if c.verbose {
			fmt.Fprintf(os.Stdout, "New starter: [%s]\n", k)
		}
		c.starters = append(c.starters, k)
	}
}
