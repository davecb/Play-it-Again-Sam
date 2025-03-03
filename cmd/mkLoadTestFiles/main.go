// Create files from a load script in "perf" format. This supports a program
// that GETs, not PUTs, POSTs or DELETEs. PUTs are easy, as are DELETEs,
// but POSTs are ambiguous.
// input looks like "01-Mar-2017 16:00:00 0 0 0 0 path 200 GET"
package main

import (
	"github.com/davecb/Play-it-Again-Sam/pkg/loadTesting"

	"flag"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/vharitonsky/iniflags"
)

// See runLoadTest.go for any package vars.

// main interprets the options and args.
func main() {
	var startFrom, runFor int
	var verbose, zero bool
	var err error

	flag.IntVar(&runFor, "for", 0, "number of records to use, eg 1000 ")
	flag.IntVar(&startFrom, "from", 0, "number of records to skip, eg 100")
	flag.BoolVar(&zero, "zero", false, "create zero-size files")
	flag.BoolVar(&verbose, "v", false, "verbose")

	iniflags.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs

	if flag.NArg() < 1 {
		fmt.Fprint(os.Stderr, "Usage: mkLoadTestFiles [-v][--from N --for N] load-file.csv [url]\n") //nolint
		flag.PrintDefaults()
		os.Exit(1)
	}
	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("No load-test csv file provided, halting.\n")
	}
	if runFor == 0 {
		runFor = math.MaxInt64
	}
	baseURL := flag.Arg(1)
	if baseURL == "" {
		log.Printf("No url provided, writing to current directory")
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening %s: %s, halting.", filename, err)
	}
	defer f.Close() // nolint

	loadTesting.MkLoadTestFiles(f, filename, baseURL, startFrom, runFor,
		loadTesting.Config{
			Verbose:  verbose,
			Protocol: loadTesting.FilesystemProtocol,
			Strip:    "",
			Zero:     zero,
			// TerminationTimeout is 0
		})

}
