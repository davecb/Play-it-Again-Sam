// Run a load test from a script in "perf" format.
// input looks like "01-Mar-2017 16:00:00 0 0 0 0 path ERROR=400"
package main

import (
	"newLoadTesting/pkg/loadTesting"

	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

const terminationTimeout = 35

func usage() {
	fmt.Fprint(os.Stderr, "Usage: runLoadTest --tps TPS [--progress TPS][--from sec --for sec][-v] load-file.csv baseURL\n")
	flag.PrintDefaults()
	os.Exit(1)
}

// main interprets the options and args.
func main() {
	var tpsTarget, progressRate, stepDuration int
	var startFrom, runFor int
	var s3, ceph, rest bool
	var verbose, debug bool
	var serial, cache, realTime  bool
	var configFile, strip, hostHeader string
	var err error

	flag.IntVar(&runFor, "for", 0, "number of records to use, eg 1000 ")
	flag.IntVar(&startFrom, "from", 0, "number of records to skip, eg 100")
	flag.IntVar(&tpsTarget, "tps", 0, "TPS target")
	flag.IntVar(&stepDuration, "duration", 10, "Duration of a step")
	flag.IntVar(&progressRate, "progress", 0, "progress rate in TPS steps")
	flag.BoolVar(&s3, "s3", false, "use s3 protocol")
	flag.BoolVar(&ceph, "ceph", false, "use ceph native protocol")
	flag.BoolVar(&rest, "rest", false, "use rest protocol")
	flag.BoolVar(&serial, "serialize", false, "serialize load (only for load testing)")
	flag.StringVar(&configFile, "config", "/home/davecb/vagrant/aoi1/src/RCDN/appsettings.txt", "config file")
	flag.StringVar(&strip, "strip", "", "strip text from paths")
	flag.StringVar(&hostHeader, "host-header", "", "add a Host: header")
	flag.BoolVar(&cache, "cache", false, "allow caching")
	flag.BoolVar(&realTime, "real-time", false, "tail -f the input file")
	flag.BoolVar(&debug, "d", false, "add debugging")
	flag.BoolVar(&verbose, "v", false, "set verbose to true")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprint(os.Stderr, "You must supply a load.csv file and a url\n")
		usage()
	}
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs
	if runFor == 0 {
		runFor = math.MaxInt64
	}
	if tpsTarget == 0 && !realTime {
		log.Fatal("You must specify a --tps target, halting.")
	}

	var proto = loadTesting.HTTPProtocol
	switch {
	case s3:
		proto = loadTesting.S3Protocol
		err = loadTesting.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("Could not read config file %s, halting. %v", configFile, err)
		}
	case ceph:
		proto = loadTesting.CephProtocol
	default: //REST
		proto = loadTesting.HTTPProtocol
	}

	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("No load-test csv file provided, halting.\n")
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

	loadTesting.RunLoadTest(io.Reader(f), filename, startFrom, runFor, tpsTarget, progressRate, baseURL,
		loadTesting.Config{
			Verbose:      verbose,
			Debug:        debug,
			Serialize:    serial,
			Cache:        cache,
			RealTime:     realTime,
			Protocol:     proto,
			Strip:        strip,
			Timeout:      terminationTimeout,
			StepDuration: stepDuration,
			HostHeader:   hostHeader,
		})
}
