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
		ess, err := snippet.Get(*expect)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		gofiles, err := snippet.Paths(args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		for _, file := range gofiles {
			err := snippet.Update(file, ess)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}
		}
		return
	}

	var errs []error
	for _, arg := range args {
		errs = append(errs, snippet.Compare(*expect, arg))
	}
	if err := errors.Join(errs...); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
}
