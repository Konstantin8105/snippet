//go:build ingore

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Konstantin8105/snippet"
)

var (
	write  = flag.Bool("w", false, "write result to (source) file instead of stdout")
	help   = flag.Bool("h", false, "show help information")
	expect = flag.String("e", snippet.ExpectSnippets, "location of expected snippets")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: gosnippet [flags] [path ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		args = append(args, ".")
	}

	if *write {
		// replace into file
	}

	var errs []error
	for _, arg := range args {
		errs = append(errs, snippet.Compare(*expect, arg))
	}
	if err := errors.Join(errs...); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(2)
	}
}
