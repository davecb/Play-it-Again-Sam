// Run a load test from a script in "perf" format.
// input looks like "01-Mar-2017 16:00:00 0 0 0 0 path ERROR=400"
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

const terminationTimeout = 35

func usage() {
	fmt.Fprint(os.Stderr, "Usage: runLoadTest --tps TPS [--progress TPS][--from rec --for rec][-v] load-file.csv baseURL\n")
	flag.PrintDefaults()
	os.Exit(1)
}

// main interprets the options and args.
func main() {
	var tpsTarget, progressRate, stepDuration, startTps int
	var startFrom, runFor int
	var s3, ceph, rest bool
	var ro, rw, wo bool
	var s3Bucket, s3Key, s3Secret string
	var verbose, debug, crash bool
	var serial, cache, tail bool
	var strip, hostHeader string
	var err error

	flag.IntVar(&runFor, "for", 0, "number of records to use, eg 1000 ")
	flag.IntVar(&startFrom, "from", 0, "number of records to skip, eg 100")
	flag.IntVar(&tpsTarget, "tps", 0, "TPS target")
	flag.IntVar(&progressRate, "progress", 0, "progress rate, in TPS steps")
	flag.IntVar(&progressRate, "start-tps", 0, "TPS to start from")
	flag.IntVar(&stepDuration, "duration", 10, "Duration of a step")

	flag.BoolVar(&s3, "s3", false, "use s3 protocol")
	flag.BoolVar(&rest, "rest", false, "use rest protocol")

	flag.BoolVar(&ro, "ro", true, "read-only test")
	flag.BoolVar(&rw, "rw", false, "read-write test")
	flag.BoolVar(&wo, "wo", false, "write-only test")

	flag.BoolVar(&serial, "serialize", false, "serialize load (only for load testing)")
	flag.StringVar(&strip, "strip", "", "test to strip from paths")
	flag.StringVar(&hostHeader, "host-header", "", "add a Host: header")
	flag.BoolVar(&cache, "cache", false, "allow caching")
	flag.BoolVar(&tail, "tail", false, "tail -f the input file")

	flag.BoolVar(&debug, "d", false, "add debugging messages")
	flag.BoolVar(&verbose, "v", false, "add verbose messages")
	flag.BoolVar(&crash, "crash", false, "exit on any error return")

	flag.StringVar(&s3Bucket, "s3-bucket", "BUCKET NOT SET",
		"set bucket when using s3 protocol")
	flag.StringVar(&s3Key, "s3-key", "KEY NOT SET",
		"set key when using s3 protocol")
	flag.StringVar(&s3Secret, "s3-secret", "SECRET NOT SET",
		"set secret when using s3 protocol")
	iniflags.Parse()

	if flag.NArg() < 2 {
		fmt.Fprint(os.Stderr, "You must supply a load.csv file and a url\n")
		usage()
	}
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs

	if runFor == 0 {
		runFor = math.MaxInt64
	}

	if tpsTarget == 0 {
		log.Fatal("You must specify a --tps target, halting.")
	}
	if wo || rw {
		log.Fatal("Read-write and write-only tests are not yet implemented, use default or -ro.")
	}

	proto, err := setProtocol(s3, ceph)
	if err != nil {
		log.Fatalf("Error Setting protocol %v, halting.", err)
	}

	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("No load-test .csv file provided, halting.\n")
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening %s: %s, halting.", filename, err)
	}
	defer f.Close() // nolint

	baseURL := flag.Arg(1)
	if baseURL == "" {
		log.Fatalf("No base url provided, halting. \n")
	}

	loadTesting.RunLoadTest(f, filename, startFrom, runFor,
		tpsTarget, progressRate, startTps, baseURL,
		loadTesting.Config{
			Verbose:      verbose,
			Debug:        debug,
			Crash:	      crash,
			Serialize:    serial,
			Cache:        cache,
			Tail:         tail,
			Protocol:     proto,
			S3Key:        s3Key,
			S3Secret:     s3Secret,
			S3Bucket:     s3Bucket,
			Strip:        strip,
			Timeout:      terminationTimeout,
			StepDuration: stepDuration,
			HostHeader:   hostHeader,
		})
}

func setProtocol(s3, ceph bool) (int, error) {
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
