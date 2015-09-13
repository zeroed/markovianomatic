package main

//+build !test
// go test -v -cover -tags test

import (
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

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	c := markovianomatic.NewChain(*prefixLen)
	c.Build(os.Stdin)
	text := c.Generate(*numWords)
	fmt.Println(text)
}
