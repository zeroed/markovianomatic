package main

//+build !test
// go test -v -cover -tags test

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/zeroed/markovianomatic"
)

func main() {
	numWords := flag.Int("words", 100, "maximum number of words to print")
	prefixLen := flag.Int("prefix", 2, "prefix length in words")
	file := flag.String("file", "", "text file to use as seed")
	verbose := flag.Bool("verbose", false, "verbose")

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	c := markovianomatic.NewChain(*prefixLen, *verbose)
	if len(*file) == 0 {
		c.Build(os.Stdin)
	} else {
		c.Load(*file)
	}

	b := new(bytes.Buffer)
	c.Generate(b, *numWords)
	fmt.Println(b.String())
}
