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
	//"github.com/aws/aws-sdk-go/service/s3"
)

// The protocols supported by the library
const (
	FilesystemProtocol = iota
	RESTProtocol       // Eg, RCDN's http-based REST Protocol
	S3Protocol
	CephProtocol
)

// operations are the things a protocol must support
type operation interface {
	Init()
	Get(path, oldRc string) error
	Put(path string, size int64) error // FIXME add oldRc
}

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
	Crash        bool
	Serialize    bool
	Cache        bool
	Tail         bool
	Protocol     int
	S3Bucket     string
	S3Key        string
	S3Secret     string
	Strip        string
	Timeout      time.Duration
	StepDuration int
	HostHeader   string
}

var conf Config
var op operation
var random = rand.New(rand.NewSource(42))
var pipe = make(chan []string, 100)
var alive = make(chan bool, 1000)
var closed = make(chan bool)

var junkDataFile = "/tmp/LoadTestJunkDataFile" // FIXME for write and r/w tests
const size = 396759652                         // nolint // FIXME, this is a heuristic

// RunLoadTest does whatever main figured out that the caller wanted.
func RunLoadTest(f *os.File, filename string, fromTime, forTime int,
	tpsTarget, progressRate, startTps int, baseURL string, cfg Config) {
	var processed = 0
	conf = cfg

	if conf.Debug {
		log.Printf("new runLoadTest(f, tpsTarget=%d, progressRate=%d, "+
			"startTps=%d, fromTime=%d, forTime=%d, baseURL=%s)\n",
			tpsTarget, progressRate, startTps, fromTime, forTime, baseURL)
	}

	// Figure out set of operations to use
	switch conf.Protocol {
	case RESTProtocol:
		op = RestProto{prefix: baseURL}
		op.Init()
	case S3Protocol:
		op = S3Proto{prefix: baseURL}
		op.Init()
	default:
		log.Fatalf("protocol %d not implemented yet", conf.Protocol)
	}
	// FIXME defer os.Remove(junkDataFile) // nolint

	go workSelector(f, filename, fromTime, forTime, pipe)             // which pipes work to ...
	go generateLoad(pipe, tpsTarget, progressRate, startTps, baseURL) // which then writes to "alive"
	for {
		select {
		case <-alive:
			processed++

		case <-time.After(time.Second * conf.Timeout):
			log.Printf("%d records processed\n", processed)
			log.Printf("No activity after %d seconds, halting normally.\n",
				conf.Timeout)
			os.Exit(0)
		}
	}
}

// workSelector pipes a selection from a file to the workers
func workSelector(f *os.File, filename string, startFrom, runFor int, pipe chan []string) { // nolint

	if conf.Debug {
		log.Printf("in workSelector(r, %s, startFrom=%d runFor=%d, pipe)\n", filename, startFrom, runFor)
	}
	if conf.Tail {
		// if we're tailing, start at the end
		_, err := f.Seek(0, io.SeekEnd)
		if err != nil {
			log.Fatalf("Fatal error seeking to the end of %s: %s\n", filename, err)
		}
		log.Printf("seeked to the end of %s, doing a tail -f with normal timeouts\n",
			filename)
	}
	r := csv.NewReader(f)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1 // ignore differences

	skipForward(startFrom, r, filename)
	recNo, pipe := copyToPipe(runFor, r, filename, pipe)
	log.Printf("Loaded %d records, closing input\n", recNo)
	close(pipe)
}

// copyToPipe sends work to the workers
func copyToPipe(runFor int, r *csv.Reader, filename string, pipe chan []string) (int, chan []string) {
	// From there, copy to pipe

	recNo := 0
forloop:
	for ; recNo < runFor; recNo++ {
		record, err := r.Read()
		switch {
		case err == io.EOF && conf.Tail:
			// just keep reading
			time.Sleep(100 * time.Millisecond)
			continue
		case err == io.EOF:
			log.Printf("At EOF on %s, no new work to queue\n", filename)
			break forloop
		case err != nil:
			log.Fatalf("Fatal error mid-way in %s: %s\n", filename, err)
		}
		if len(record) != 9 {
			// FIXME? this discards real-time part-records
			continue
		}

		if conf.Strip != "" {
			record[pathField] = strings.Replace(record[pathField], conf.Strip, "", 1)
		}
		// log.Printf("writing %v to pipe\n", record)

		pipe <- record
	}
	return recNo, pipe
}

// generateLoad starts progressRate new threads every 10 seconds until we hit progressRate
func generateLoad(pipe chan []string, tpsTarget, progressRate, startTps int, urlPrefix string) {
	if conf.Debug {
		log.Printf("generateLoad(pipe, tpsTarget=%d, progressRate=%d, from, for, prefix\n",
			tpsTarget, progressRate)
	}

	fmt.Print("#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc\n")
	switch {
	case progressRate != 0:
		runProgressivelyIncreasingLoad(progressRate, tpsTarget, startTps, pipe)
	case tpsTarget != 0:
		runSteadyLoad(tpsTarget, pipe)
	case tpsTarget <= 0:
		log.Fatal("A zero or negative tps target is not meaningfull, halting\n")
	}
}

// run at a steady tps until the end of the data
func runSteadyLoad(tpsTarget int, pipe chan []string) {
	log.Printf("starting, at %d requests/second\n", tpsTarget)
	// start tpsTarget workers
	for i := 0; i < tpsTarget; i++ {
		go worker(pipe)
	}
}

// runProgressivelyIncreasingLoad, the classic load test
func runProgressivelyIncreasingLoad(progressRate, tpsTarget, startTps int, pipe chan []string) {

	// start the first workers
	if startTps == 0 {
		startTps = progressRate
	}
	rate := startTps
	for i := 0; i < startTps; i++ {
		go worker(pipe)
	}
	// add to the workers until we have enough
	log.Printf("now at %d requests/second\n", rate)
	for range time.Tick(time.Duration(conf.StepDuration) * time.Second) { // nolint
		//start another progressRate of workers
		rate += progressRate
		if rate > tpsTarget {
			// OK, we're past the range, quit.
			log.Printf("completed maximum rate, starting %d sec cleanup timer\n", conf.Timeout)
			break
		}
		for i := 0; i < progressRate; i++ {
			go worker(pipe)
		}
		log.Printf("now at %d requests/second\n", rate)
	}
	// let them run for a cycle and shut down
	time.Sleep(time.Duration(10 * float64(time.Second)))
	close(closed) // We're done
}

// worker reads and executes a task every second until it hits eof
func worker(pipe chan []string) {
	if conf.Debug {
		log.Print("started a worker\n")
	}
	// wait a random fraction of one second before looping, for randomness.
	time.Sleep(time.Duration(random.Float64() * float64(time.Second)))

	for range time.Tick(1 * time.Second) { // nolint
		doWork()
	}
}

// work is the thing that happens each second.
func doWork() {
	var r []string

	select {
	case <-closed:
		if conf.Debug {
			log.Print("pipe closed, no more requests to process.\n")
		}
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
		if conf.Serialize {
			// force this NOT to be asynchronous, for long-running load tests only
			op.Get(r[pathField], r[returnCodeField]) // nolint, ignore return value
		} else {
			go op.Get(r[pathField], r[returnCodeField]) // nolint, ignore return value
		}
	default:
		log.Printf("got unimplemented operation %s in %v, ignored\n", r[operatorField], r)
	}
}
