// Create files from a load script in "perf" format. This supports a program
// that GETs, not PUTs, POSTs or DELETEs. PUTs are easy, as are DELETEs,
// but POSTs are ambiguous.
// input looks like "01-Mar-2017 16:00:00 0 0 0 0 path 200 GET"
package main

import (
	"loadTesting/pkg/loadTesting"


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
	var ceph, s3, rest bool
	var configFile string
	var verbose bool
	var err error

	flag.IntVar(&runFor, "for", 0, "number of records to use, eg 1000 ")
	flag.IntVar(&startFrom, "from", 0, "number of records to skip, eg 100")
	flag.BoolVar(&s3, "s3", false, "use s3 protocol")
	flag.BoolVar(&rest, "rest", false, "use rest protocol")
	flag.BoolVar(&ceph, "ceph", false, "use ceph native protocol")
	flag.StringVar(&configFile, "config", "/home/davecb/vagrant/aoi1/src/RCDN/appsettings.txt", "config file")
	flag.BoolVar(&verbose, "v", false, "verbose")
	iniflags.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs

	if flag.NArg() < 1 {
		fmt.Fprint(os.Stderr, "Usage: mkLoadTestFiles [--s3][-v][--from N --for N] load-file.csv url\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("No load-test csv file provided, halting.\n")
	}
	if runFor == 0 {
		runFor =  math.MaxInt64
	}
	baseURL := flag.Arg(1)
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening %s: %s, halting.", filename, err)
	}
	defer f.Close() // nolint

	proto, err := setProtocol(s3, configFile, ceph)
	if err != nil {
		log.Fatalf("Error Serting protocol %v, halting.", err)
	}

	loadTesting.MkLoadTestFiles(f, filename, baseURL, startFrom, runFor,
		loadTesting.Config{
			Verbose:  verbose,
			Protocol: proto,
			Strip:    "",
			// Timeout is 0
		})

}

func setProtocol(s3 bool, configFile string, ceph bool) (int, error) {
	var err error

	var proto = loadTesting.RESTProtocol
	switch {
	case s3:
		proto = loadTesting.S3Protocol
	case ceph:
		proto = loadTesting.CephProtocol // unimplemented
	default: //REST
		proto = loadTesting.RESTProtocol
	}
	return proto, err
}
