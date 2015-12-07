package main

//+build !test
// go test -v -cover -tags test

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/codegangsta/cli"
	"github.com/zeroed/markovianomatic"
	"github.com/zeroed/markovianomatic/model"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var numWords int
	var prefixLen int
	var file string
	var verbose bool

	app := cli.NewApp()
	app.Name = "Markovianomatic"
	app.Usage = "Build a random text with Markov-ish rules"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "words",
			Value:       100,
			Usage:       "maximum number of words to print",
			Destination: &numWords,
		},
		cli.IntFlag{
			Name:        "prefix",
			Value:       2,
			Usage:       "prefix Length in words",
			Destination: &prefixLen,
		},
		cli.StringFlag{
			Name:        "file",
			Value:       "",
			Usage:       "text file to use as seed",
			Destination: &file,
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "I wanna read useless stuff",
			Destination: &verbose,
		},
	}

	app.Action = func(cc *cli.Context) {
		var c *markovianomatic.Chain
		cns, _ := model.Collections(model.Database())

		if len(cns) == 0 {
			fmt.Fprintf(os.Stderr, "There are no available PrefixBase. Start from scratch\n")
			c = newEmptyChain(prefixLen, verbose)
		} else {
			fmt.Fprintf(os.Stdout, "Want to [use/append/delete] an existing DB? [no(new)] ")
			if askForConfirmation() {
				i := chooseCollection(cns)
				c = load(prefixLen, verbose, cns[i])
			} else {
				c = newEmptyChain(prefixLen, verbose)
			}
		}

		if len(file) == 0 {
			c.Build(os.Stdin)
		} else {
			c.Load(file)
		}

		b := new(bytes.Buffer)
		c.Generate(b, numWords)
		fmt.Println(b.String())
	}

	app.Run(os.Args)
}

// askForConfirmation uses Scanln to parse user input. A user must type
// in "yes" or "no" and then press enter. It has fuzzy matching, so "y",
// "Y", "yes", "YES", and "Yes" all count as confirmations. If the input
// is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should
// use fmt to print out a question before calling askForConfirmation.
// E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func askForConfirmation() bool {
	const dflt string = "no"
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil && err.Error() == "unexpected newline" {
		response = dflt
	}

	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO", "new"}
	exitResponses := []string{"exit", "quit", "Q", "q"}
	// deleteResponses := []string{"d", "D", "delete", "Delete", "del", "DELETE"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else if containsString(exitResponses, response) {
		fmt.Fprintf(os.Stderr, "Bye\n\n")
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "Please type yes or no and then press enter:\n")
		return askForConfirmation()
	}
	return false
}

func chooseCollection(cns []string) int {
	for i, cn := range cns {
		fmt.Fprintf(os.Stdout, "[%d] %s\n", i, cn)
	}
	fmt.Fprintf(os.Stdout, "----\n")
	msg := ""
	for {
		fmt.Fprintf(os.Stdout, "[0-%02d]: %s ", len(cns)-1, msg)
		var i int
		_, err := fmt.Scanf("%d", &i)
		if err == nil && i >= 0 && i < len(cns) {
			return i
		} else {
			msg = "(nope)"
		}
	}
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func newEmptyChain(prefixLen int, verbose bool) *markovianomatic.Chain {
	fmt.Fprintf(os.Stdout, "Collection name: ")
	var cn string
	fmt.Scanf("%s", &cn)
	return markovianomatic.NewChain(prefixLen, verbose, cn)
}

func load(prefixLen int, verbose bool, cn string) *markovianomatic.Chain {
	_, dbc := model.Connect(cn)
	lc, _ := dbc.Count()
	fmt.Fprintf(os.Stdout, "Using %s with %d prefixes\n", cn, lc)
	var c *markovianomatic.Chain
	c = markovianomatic.NewChain(prefixLen, verbose, cn)
	loadChain(c, dbc)

	return c
}

func loadChain(c *markovianomatic.Chain, dbc *mgo.Collection) {
	iter := dbc.Find(bson.M{}).Iter()
	var node model.Node
	for iter.Next(&node) {
		c.Set(node.Key, node.Choices)
	}
	if err := iter.Close(); err != nil {
		fmt.Fprint(os.Stderr, "Error iterating the collection: %s\n", err.Error())
		os.Exit(1)
	}
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for i, e := range slice {
		if e == element {
			return i
		}
	}
	return -1
}
