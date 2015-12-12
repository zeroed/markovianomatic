# Markovianomatic

Inspired by "The Go Authors"
source: [codewalk/markov](https://golang.org/doc/codewalk/markov/)

## Options

```
$ main/main -help
NAME:
   Markovianomatic - Build a random text with Markov-ish rules

USAGE:
   main/main [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --words "100"  maximum number of words to print
   --prefix "2"   prefix Length in words
   --file     text file to use as seed
   --verbose    I wanna read useless stuff
   --help, -h   show help
   --version, -v  print the version
```

Sample:

```
$ ./main -file ~/Downloads/wittgenstein.txt
no connection string provided: using Mongo localhost
Want to [u]se, [a]ppend, [d]elete an existing DB? [n]o(new): a
[0] witt01
----
[0-00]:  0

Using witt01 with 35101 prefixes
Node (35101/35101)   13s [====================================================================] 100%

35101 prefixes, prefixes 2 long. generating text ...

se vi è gusto il gradevole dal bello anche nel campo pratico egli semplicemente afferma che il bello ma anche il giudizio estetico non è nelle cose non ci simpegna per risolvere questo problema la critica del giudizio critica della ragion pratica che ci fosse una corrispondenza tra limperativo categorico e le inclinazioni se vive questa frustrazione dovuta allintenzionalità della morale kantiana ma della morale la buona volontà è una facoltà analitica che si chiama riflettente perché mentre nel giudizio riflettente è un giudizio conoscitivo il giudizio analitico nel secondo caso ci indica la forma del volere conforme a leggi

```

## Original documentation

     // Copyright 2011 The Go Authors.  All rights reserved.
     // Use of this source code is governed by a BSD-style
     // license that can be found in the LICENSE file.

     /*
     Generating random text: a Markov chain algorithm

     Based on the program presented in the "Design and Implementation" chapter
     of The Practice of Programming (Kernighan and Pike, Addison-Wesley 1999).
     See also Computer Recreations, Scientific American 260, 122 - 125 (1989).

     A Markov chain algorithm generates text by creating a statistical model of
     potential textual suffixes for a given prefix. Consider this text:

     	I am not a number! I am a free man!

     Our Markov chain algorithm would arrange this text into this set of prefixes
     and suffixes, or "chain": (This table assumes a prefix length of two words.)

     	Prefix       Suffix

     	"" ""        I
     	"" I         am
     	I am         a
     	I am         not
     	a free       man!
     	am a         free
     	am not       a
     	a number!    I
     	number! I    am
     	not a        number!

     To generate text using this table we select an initial prefix ("I am", for
     example), choose one of the suffixes associated with that prefix at random
     with probability determined by the input statistics ("a"),
     and then create a new prefix by removing the first word from the prefix
     and appending the suffix (making the new prefix is "am a"). Repeat this process
     until we can't find any suffixes for the current prefix or we exceed the word
     limit. (The word limit is necessary as the chain table may contain cycles.)

     Our version of this program reads text from standard input, parsing it into a
     Markov chain, and writes generated text to standard output.
     The prefix and output lengths can be specified using the -prefix and -words
     flags on the command-line.
     */
