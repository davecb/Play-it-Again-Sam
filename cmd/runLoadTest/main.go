// Run a load test from a script in "perf" format.
// input looks like "01-Mar-2017 16:00:00 0 0 0 0 path 200 GET"
package main

import (
	"github.com/davecb/Play-it-Again-Sam/pkg/loadTesting"

	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/vharitonsky/iniflags"
)

const terminationTimeout = 35

func usage() {
	//nolint
	fmt.Fprint(os.Stderr, "Usage: runLoadTest --tps TPS [--progress "+
		"TPS][--from rec --for rec][-v] load-file.csv baseURL\n")
	flag.PrintDefaults()
	os.Exit(1)
}

// main interprets the options and args.
func main() {
	var tpsTarget, progressRate, stepDuration, startTps int
	var startFrom, runFor int
	var s3, ceph, rest bool
	var ro bool
	var rw, wo int64
	var bufSize int64
	var s3Bucket, s3Key, s3Secret string
	var verbose, debug, crash, akamaiDebug bool
	var sleepTime float64
	var cache, tail bool
	var strip, hostHeader, headers string
	var headerMap = make(map[string]string)
	var err error

	flag.IntVar(&runFor, "for", 0, "number of records to use, eg 1000 ")
	flag.IntVar(&startFrom, "from", 0, "number of records to skip, eg 100")
	flag.IntVar(&tpsTarget, "tps", 0, "TPS target")
	flag.IntVar(&progressRate, "progress", 0, "progress rate, in TPS steps")
	flag.IntVar(&progressRate, "start-tps", 0, "TPS to start from")
	flag.IntVar(&stepDuration, "duration", 10, "Duration of a step")

	flag.BoolVar(&s3, "s3", false, "use s3 protocol")
	flag.BoolVar(&rest, "rest", false, "use rest protocol")

	flag.BoolVar(&ro, "ro", false, "read-only test")
	flag.Int64Var(&rw, "rw", 0, "read-write test, w buffer size")
	flag.Int64Var(&wo, "wo", 0, "write-only test, w buffer size")

	flag.Float64Var(&sleepTime, "sleep", 1.0, "sleep time, seconds")

	flag.StringVar(&strip, "strip", "", "test to strip from paths")
	flag.StringVar(&hostHeader, "host-header", "", "add a Host: header")
	flag.StringVar(&headers, "headers", "", "add one or more key:value headers")

	flag.BoolVar(&cache, "cache", false, "allow caching")
	flag.BoolVar(&tail, "tail", false, "tail -f the input file")

	flag.BoolVar(&debug, "d", false, "add debugging messages")
	flag.BoolVar(&verbose, "v", false, "add verbose messages")
	flag.BoolVar(&crash, "crash", false, "exit on any error return")
	flag.BoolVar(&akamaiDebug, "akamai-debug", false, "add akamai debugging headers")

	flag.StringVar(&s3Bucket, "s3-bucket", "BUCKET NOT SET",
		"set bucket when using s3 protocol")
	flag.StringVar(&s3Key, "s3-key", "KEY NOT SET",
		"set key when using s3 protocol")
	flag.StringVar(&s3Secret, "s3-secret", "SECRET NOT SET",
		"set secret when using s3 protocol")
	iniflags.Parse()

	if flag.NArg() < 2 {
		fmt.Fprint(os.Stderr, "You must supply a load.csv file and a url\n") //nolint
		usage()
	}
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs

	setHeaders(headers, headerMap)
	if runFor == 0 {
		runFor = math.MaxInt64
	}

	if tpsTarget == 0 {
		log.Fatal("You must specify a --tps target, halting.")
	}

	// Interpret rw, ro and wo options
	r, w := setMode(ro, rw, wo)
	if wo != 0 {
		bufSize = wo
	} else if rw != 0 {
		bufSize = rw
	}

	proto := setProtocol(s3, ceph)
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
			Crash:        crash,
			AkamaiDebug:  akamaiDebug,
			SleepTime:    sleepTime,
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
			HeaderMap:    headerMap,
			R:            r,
			W:            w,
			BufSize:      bufSize,
		})
}

// setheaders creates a proper map of header:value pairs
func setHeaders(headers string, headerMap map[string]string) {
	if headers != "" {
		tokens := strings.Split(headers, " ")
		for _, t := range tokens {
			x := strings.Split(t, ":")
			if len(x) != 2 || x[0] == "" || x[1] == "" {
				log.Fatalf("headers must contain a key:value pair, found %q instead\n", t)
			}
			headerMap[x[0]] = x[1]
		}
	}
}

// setProtocol from s3 and ceph booleans
func setProtocol(s3, ceph bool) int {
	var proto int

	switch {
	case s3:
		proto = loadTesting.S3Protocol
	case ceph:
		proto = loadTesting.CephProtocol // unimplemented
	default: //REST
		proto = loadTesting.RESTProtocol
	}
	return proto
}

// setMode sets the r and w booleans based on the options set
// ro appears to default rue???
func setMode(ro bool, rw, wo int64) (bool, bool) {
	var r, w bool

	//log.Printf("ro=%t, rw=%d, wo=%d\n", ro, rw, wo)
	switch {
	case ro:
		r = true
		w = false
	case rw != 0:
		r = true
		w = true
	case wo != 0:
		r = false
		w = true
	default: // treat as ro if not set
		r = true
		w = false
	}
	//log.Printf("r=%t,w=%t\n", r, w)
	return r, w

}
