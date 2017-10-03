package loadTesting

// Run a load test from a script in "perf" format.
// input looks like "01-Mar-2017 16:00:00 0 0 0 0 path 404 GET"

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// The protocols supported by the library
const (
	FilesystemProtocol = iota
	HTTPProtocol       // Eg, RCDN's http-based REST Protocol
	S3Protocol
	CephProtocol
)

// These are the field names in the csv file
const ( // nolint
	dateField         = iota // nolint
	timeField                // nolint
	latencyField             // nolint
	transferTimeField        // nolint
	thinkTimeField           // nolint
	bytesField
	pathField
	returnCodeField
	operatorField
)

// Config contains all the optional parameters.
type Config struct {
	Verbose      bool
	Debug        bool
	Serialize    bool
	Cache        bool
	RealTime	 bool
	Protocol     int
	Strip        string
	Timeout      time.Duration
	StepDuration int
	HostHeader   string
}

// Verbose turns on extra information for understanding the test
var verbose bool

// Debug turns on trace stuff
var debug bool

// cache turns off no-cache header
var cache bool

// realTime turns on pipe-like processing
var realTime bool

// Protocol indicates which of the above is in use
var protocol int

// Strip is the string to Strip from paths, or ""
var strip string

// Timeout is the delay before shutting dowsn
var timeout time.Duration

// stepDuration, typically 10 or 30
var stepDuration int

// optional host header
var hostHeader string

// serialize load tests, which changes the applied load: be carefull!
var serialize bool

var random = rand.New(rand.NewSource(42))
var pipe = make(chan []string, 100)
var alive = make(chan bool, 1000)

///var bucketName = "loadtest"
var junkDataFile = "/tmp/LoadTestJunkDataFile"

// nolint
const size = 396759652 // FIXME, this is a heuristic

// RunLoadTest does whatever main figured out that the caller wanted.
func RunLoadTest(f io.Reader, filename string, fromTime, forTime int,
	tpsTarget, progressRate int, baseURL string, conf Config) {
	var processed = 0

	// get settings from conf parameter
	setConf(conf)

	if debug {
		log.Printf("new runLoadTest(f, tpsTarget=%d, progressRate=%d, fromTime=%d, forTime=%d, baseURL=%s)\n",
			tpsTarget, progressRate, fromTime, forTime, baseURL)
	}

	doPrepWork(baseURL)           // Named "init" fucntion, creates junkDataFile
	defer os.Remove(junkDataFile) // nolint

	go workSelector(f, filename, fromTime, forTime, strip, pipe) // which pipes work to ...
	go generateLoad(pipe, tpsTarget, progressRate, baseURL)      // which then writes to "alive"
	for {
		select {
		case <-alive:
			processed++

		case <-time.After(time.Second * timeout):
			log.Printf("%d records processed\n", processed)
			log.Printf("No activity after %d seconds, halting normally.\n",
				timeout)
			os.Exit(0)
		}
	}
}

// setConf copies the conf parameters to globals
func setConf(conf Config) {
	verbose = conf.Verbose
	debug = conf.Debug
	cache = conf.Cache
	realTime = conf.RealTime
	serialize = conf.Serialize
	protocol = conf.Protocol
	strip = conf.Strip
	timeout = conf.Timeout
	stepDuration = conf.StepDuration
	hostHeader = conf.HostHeader
}

// workSelector pipes a selection from a file to the workers
func workSelector(f io.Reader, filename string, startFrom, runFor int, strip string, pipe chan []string) {

	if debug {
		log.Printf("in workSelector(r, %s, startFrom=%d runFor=%d, pipe)\n", filename, startFrom, runFor)
	}
	r := csv.NewReader(f)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1 // ignore differences

	// Skip forward to the starting point
	for i := 0; startFrom != 0; i++ {
		_, err := r.Read()
		if err == io.EOF {
			log.Printf("EOF at line %d, test will be empty\n", i)
			break
		}
		if err != nil {
			log.Fatalf("Fatal error skipping forward in %s: %s\n", filename, err)
		}
		if i >= startFrom {
			log.Printf("Skipped over %d lines, about to start reading data\n", i)
			break
		}
	}

	// From there, copy to pipe
	recNo := 0
	for ; recNo < runFor; recNo++ {
		record, err := r.Read()
		if err == io.EOF {
			if realTime {
				// just keep reading
				time.Sleep(time.Millisecond)
				continue
			}
			//log.Printf("At EOF on %s, no new work to queue\n", filename)
			break
		}
		if err != nil {
			log.Fatalf("Fatal error mid-way in %s: %s\n", filename, err)
		}
		if strip != "" {
			record[pathField] = strings.Replace(record[pathField], strip, "", 1)
		}
		//log.Printf("writing %v to pipe\n", record)
		pipe <- record
	}
	log.Printf("Loaded %d records, closing input\n", recNo)
	close(pipe)
}

// generateLoad starts progressRate new threads every 10 seconds until we hit progressRate
func generateLoad(pipe chan []string, tpsTarget, progressRate int, urlPrefix string) {
	if debug {
		log.Printf("generateLoad(pipe, tpsTarget=%d, progressRate=%d, from, for, prefix\n",
			tpsTarget, progressRate)
	}

	fmt.Print("#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc\n")
	var closed = make(chan bool)
	switch {
	case progressRate != 0:
		// start the first workers
		rate := progressRate
		for i := 0; i < progressRate; i++ {
			go worker(pipe, closed, urlPrefix)
		}
		// add to the workers until we have enough
		log.Printf("now at %d requests/second\n", rate)
		for range time.Tick(time.Duration(stepDuration) * time.Second) { // nolint
			//start another progressRate of workers
			rate += progressRate
			if rate > tpsTarget {
				// OK, we're past the range, quit.
				log.Printf("completed maximum rate, starting %d sec cleanup timer\n", timeout)
				break
			}
			for i := 0; i < progressRate; i++ {
				// start a worker
				go worker(pipe, closed, urlPrefix)
			}
			log.Printf("now at %d requests/second\n", rate)
		}
		// let them run for a cycle and shut down
		time.Sleep(time.Duration(10 * float64(time.Second)))
		close(closed) // We're done
	case tpsTarget != 0:
		// start tpsTarget workers, to run until out of data.
		log.Printf("starting, at %d requests/second\n", tpsTarget)
		for i := 0; i < tpsTarget; i++ {
			// start a worker
			go worker(pipe, closed, urlPrefix)
		}
	case tpsTarget < 0:
		log.Fatal("A zero or negative tps target is not meaningfull, halting\n")
	case realTime:
		log.Print("starting to read the input file continuously, ^C to stop\n")
		for i := 0; i < 3; i++ {
			// The "3" is a heuristic
			go worker(pipe, closed, urlPrefix)
		}
	}
}

// worker reads and executes a task every second until it hits eof
func worker(pipe chan []string, closed chan bool, urlPrefix string) {
	if debug {
		log.Print("started a worker\n")
	}
	// wait a random fraction of one second before looping, for randomness.
	time.Sleep(time.Duration(random.Float64() * float64(time.Second)))

	for range time.Tick(1 * time.Second) { // nolint
		var r []string

		select {
		case <-closed:
			//log.Print("pipe closed, no more requests to send.\n")
			return
		case r = <-pipe:
			//log.Printf("got %v\n", r)
		}

		switch {
		case r == nil:
			//log.Print("worker reached EOF, no more requests to send.\n")
			return
		case len(r) != 9:
			// bad input data, crash please.
			log.Fatalf("number of fields != 9 in %v", r)
		case r[operatorField] == "GET":
			if serialize {
				// force this NOT to be asynchronous, for load tests only
				getJunkFile(urlPrefix, r[pathField])
			} else {
				go getJunkFile(urlPrefix, r[pathField])
			}
		case r[operatorField] == "PUT":
			// FIXME: treat PUT as GET, probably a 404
			if serialize {
				getJunkFile(urlPrefix, r[pathField])
			} else {
				go getJunkFile(urlPrefix, r[pathField])
			}
			//fileSize, err := strconv.ParseInt(r[bytesField], 10, 64)
			//if err != nil {
			//	// fail the run if given bad input data here, too
			//	log.Fatalf("could not parse size field in %v\n", r)
			//}
			//go putJunkFile(urlPrefix, r[pathField], fileSize)
		default:
			log.Fatal("operations other than GET and PUT are not implemented yet\n")
		}
	}
}

// putJunkFile sends a specified number of bytes from /dev/urandom via a PUT
func putJunkFile(baseURL, path string, size int64) { // nolint
	if debug {
		log.Printf("in putJunkFile(%s, %s, %d)\n", baseURL, path, size)
	}

	switch protocol {
	case HTTPProtocol:
		RestPut(baseURL, path, size)
	case S3Protocol:
		AmazonS3Put(baseURL, path, size) // nolint
	default:
		log.Fatalf("Protocol %s not implemented yet\n", string(protocol))
	}
	// alive <- true
}

// get a url and do nothing with it.
func getJunkFile(baseURL, path string) {
	if debug {
		log.Printf("in getJunkFile(%s, %s), protocol=%v\n", baseURL, path, protocol)
	}

	switch protocol {
	case HTTPProtocol:
		RestGet(baseURL, path)
	case S3Protocol:
		//MinioS3Get(baseURL, path)
		AmazonS3Get(baseURL, path)
	default:
		log.Fatalf("Protocol %s not implemented yet\n", string(protocol))
	}
	// alive <- true
}

// doPrepWork makes sure we have the prerequisites by protocol
func doPrepWork(baseURL string) {
	//MustCreateFilesystemFile(junkDataFile, size)  FXIME. needed for PUT
	switch protocol {
	case S3Protocol:
		AmazonS3Prep(baseURL)
	}
}
