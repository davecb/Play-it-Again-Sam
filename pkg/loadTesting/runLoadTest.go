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
	//"google.golang.org/genproto/googleapis/watcher/v1"
	"strconv"
	"strings"
	"syscall"
	"time"

	//"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/fsnotify.v1"
)

// The protocols supported by the library
const (
	FilesystemProtocol = iota
	RESTProtocol       // Simple http-based REST protocols
	S3Protocol         // Amazon s3 protocol or compatable
	CephProtocol       // reserved for native ceph protocol
	TimeBudgetProtocol // see if we're inside our time budget
)

// operations are the things a protocol must support
type operation interface {
	Init()
	Get(path, oldRc string)
	Put(path, size, oldRc string)
	Post(path, size, oldRc, body string)
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
	bodyField
)

// Config contains all the optional parameters.
type Config struct {
	Verbose        bool   // Extra info about requests
	Debug          bool   // Extra info about program
	Crash          bool   // Halt on any error
	Serialize      bool   // FIXME semi-evil hack
	Cache          bool   // allow caching
	Tail           bool   // tail a log
	ThunderingHerd bool   // don't randomize request times
	AkamaiDebug    bool   // add Akamai debug headers
	Protocol       int    // rest, etc
	S3Bucket       string // s3-specific options
	S3Key          string
	S3Secret       string
	Strip          string
	Timeout        time.Duration     // time to wait at end
	StepDuration   int               // duration of a test step
	HostHeader     string            // add a Host: header
	HeaderMap      map[string]string // one or more key:value headers
	R              bool              // read tests allowed
	W              bool              // write tests allowed
	BufSize        int64             // max size of written file
}

var OfferedRate int // Log offered rate in TPS

var conf Config
var op operation
var random = rand.New(rand.NewSource(42))
var pipe = make(chan []string, 100)
var alive = make(chan bool, 1000)
var closed = make(chan bool)
var junkDataFile = "/tmp/LoadTestJunkDataFile"

const size = 396759652 // nolint // FIXME, this is a heuristic

// RunLoadTest does whatever main figured out that the caller wanted.
func RunLoadTest(f *os.File, filename string, fromTime, forTime int,
	tpsTarget, progressRate, startTps int, baseURL string, cfg Config) {
	var processed = 0
	conf = cfg
	defer reportRUsage("RunLoadTest", time.Now())

	if conf.Debug {
		log.Printf("new runLoadTest(f, tpsTarget=%d, progressRate=%d, "+
			"startTps=%d, fromTime=%d, forTime=%d, baseURL=%s)\n",
			tpsTarget, progressRate, startTps, fromTime, forTime, baseURL)
	}

	// Figure out which set of operations to use
	switch conf.Protocol {
	case RESTProtocol:
		op = RestProto{prefix: baseURL}
		op.Init()
	case S3Protocol:
		op = S3Proto{prefix: baseURL}
		op.Init()
	case TimeBudgetProtocol:
		op = timeBudgetProto{prefix: baseURL}
		op.Init()
	default:
		log.Fatalf("protocol %d not implemented yet", conf.Protocol)
	}

	// Create data for rw and wo tests
	if conf.BufSize > 0 {
		log.Printf("Creating %d-byte data file %q\n", conf.BufSize,
			junkDataFile)
		mustCreateFilesystemFile(junkDataFile, conf.BufSize)
		defer os.Remove(junkDataFile) // nolint
	} else if conf.BufSize < 0 {
		log.Fatalf("A negative size for data files (%d) is meaningless, halting\n", conf.BufSize)
	}

	// select some work to do from the input file
	go workSelector(f, filename, fromTime, forTime, pipe)
	// which pipes work to ...
	go generateLoad(pipe, tpsTarget, progressRate, startTps, baseURL, conf.ThunderingHerd)
	// which then writes to "alive", ...
	for {
		select {
		case _, ok := <-alive:
			if !ok {
				// if alive was closed, we're done
				return
			}
			processed++
		case <-time.After(time.Second * conf.Timeout):
			// FIXME, this is memory-intensive
			log.Printf("%d records processed\n", processed)
			log.Printf("No activity after %d seconds, halting normally.\n",
				conf.Timeout)
			return
		}
	}
}

// workSelector pipes a selection from a file to the workers
func workSelector(f *os.File, filename string, startFrom, runFor int, pipe chan []string) { // nolint
	var watcher *fsnotify.Watcher

	if conf.Debug {
		log.Printf("in workSelector(r, %s, startFrom=%d runFor=%d, pipe)\n", filename, startFrom, runFor)
	}
	if conf.Tail {
		// if we're tailing, start at the end
		_, err := f.Seek(0, io.SeekEnd)
		if err != nil {
			log.Fatalf("Fatal error seeking to the end of %s: %s\n", filename, err)
		}
		watcher, _ = fsnotify.NewWatcher()
		if err != nil {
			log.Fatalf("Fatal error setting up fsnotify for tail of %s: %s\n", filename, err)
			// FIXME: to fall back to polling, set watcher to nil
		}
		defer watcher.Close() // nolint
		err = watcher.Add(filename)
		if err != nil {
			log.Fatalf("Fatal error addding %s to fsnotify: %s\n", filename, err)
		}
		log.Printf("seeked to the end of %s, doing a tail -f with normal timeouts\n",
			filename)
	}

	r := csv.NewReader(f)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1 // ignore differences

	skipForward(startFrom, r, filename)
	recNo := copyToPipe(runFor, r, filename, pipe, watcher)
	log.Printf("EOF: loaded %d records, closing input pipe\n", recNo)
	close(pipe)
}

// copyToPipe pipes work to the workers
func copyToPipe(runFor int, r *csv.Reader, filename string, pipe chan []string, watcher *fsnotify.Watcher) int {

	recNo := 0
forloop:
	for ; recNo < runFor; recNo++ {
		record, err := r.Read()
		switch {
		case err == io.EOF && conf.Tail:
			// just keep reading, even if we truncate...
			if watcher == nil {
				time.Sleep(100 * time.Millisecond)
			} else {
				//log.Print("waiting for fsnotify\n")
				if err = waitForChange(watcher); err != nil {
					log.Fatalf("Fatal error waiting for fsnotify on %s, %v\n", filename, err)
				}
			}
			continue
		case err == io.EOF:
			log.Printf("At EOF on %s, no new work to queue\n", filename)
			break forloop
		case err != nil:
			log.Printf("Fatal error mid-way reading %q from %s, stopping: %s\n", record, filename, err)
			break forloop
		}
		if len(record) < 9 {
			log.Printf("ill-formed record %q ignored\n",
				record)
			// Warning: this discards real-time part-records
			continue
		}

		if conf.Strip != "" {
			record[pathField] = strings.Replace(record[pathField], conf.Strip, "", 1)
		}
		//log.Printf("writing %v to pipe\n", record)

		pipe <- record
	}
	return recNo
}

// generateLoad starts progressRate new threads every 10 seconds until we hit progressRate
func generateLoad(pipe chan []string, tpsTarget, progressRate, startTps int, urlPrefix string, tHeard bool) {
	if conf.Debug {
		log.Printf("generateLoad(pipe, tpsTarget=%d, progressRate=%d, from, for, prefix\n",
			tpsTarget, progressRate)
	}

	fmt.Print("#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc op offered\n")
	switch {
	case progressRate != 0:
		runProgressivelyIncreasingLoad(progressRate, tpsTarget, startTps, pipe, tHeard)
	case tpsTarget != 0:
		runSteadyLoad(tpsTarget, pipe, tHeard)
	case tpsTarget <= 0:
		log.Fatal("A zero or negative tps target is not meaningful, halting\n")
	}
}

// run at a steady tps until the end of the data
func runSteadyLoad(tpsTarget int, pipe chan []string, tHeard bool) {
	log.Printf("starting, at %d requests/second\n", tpsTarget)
	OfferedRate = tpsTarget
	// start tpsTarget workers
	for i := 0; i < tpsTarget; i++ {
		go worker(pipe, tHeard)
	}
}

// runProgressivelyIncreasingLoad, the classic load test
func runProgressivelyIncreasingLoad(progressRate, tpsTarget, startTps int, pipe chan []string, tHeard bool) {

	// start the first workers
	if startTps == 0 {
		startTps = progressRate
	}
	rate := startTps
	OfferedRate = startTps
	for i := 0; i < startTps; i++ {
		go worker(pipe, tHeard)
	}
	// add to the workers until we have enough
	log.Printf("now at %d requests/second\n", rate)
	for range time.Tick(time.Duration(conf.StepDuration) * time.Second) { // nolint
		//start another progressRate of workers
		rate += progressRate
		OfferedRate = rate
		if rate > tpsTarget {
			// OK, we're past the range, quit.
			log.Printf("completed maximum rate, starting %d sec cleanup timer\n", conf.Timeout)
			break
		}
		for i := 0; i < progressRate; i++ {
			go worker(pipe, tHeard)
		}
		log.Printf("now at %d requests/second\n", rate)
		fmt.Printf("#TPS=%d\n", rate) // add as a column?
	}
	// let them run for a cycle and shut down
	time.Sleep(time.Duration(10 * float64(time.Second)))
	// this needs refactoring
	close(closed)
}

// worker reads and executes a task every second until it hits eof
func worker(pipe chan []string, tHerd bool) {
	if conf.Debug {
		log.Print("started a worker\n")
	}
	if conf.Protocol == TimeBudgetProtocol {
		// Do the operation immediately, once, to measure its speed
		doWork()
		return
	}
	if !tHerd {
		// wait a random fraction of one second before looping, for randomness.
		time.Sleep(time.Duration(random.Float64() * float64(time.Second)))
	} // otherwise send a "thundering herd" of requests, for a stress tests

	for range time.Tick(1 * time.Second) { // nolint
		done := doWork()
		if done {
			return
		}
	}
}

// work is the thing that happens each second.
func doWork() bool {
	var r []string

	r, eof := getWork()
	if eof {
		return true
	}

	switch {
	case r == nil:
		log.Print("worker reached EOF, no more requests to send.\n")
		return true
	case len(r) < 9:
		// bad input data, crash
		log.Fatalf("number of fields < 9 in %v", r)
	case r[operatorField] == "GET" && conf.R:
		go op.Get(r[pathField], r[returnCodeField])
	case r[operatorField] == "PUT" && conf.W:
		go op.Put(r[pathField], r[bytesField], r[returnCodeField])
	case r[operatorField] == "POST" && conf.R:
		go op.Post(r[pathField], r[bytesField], r[returnCodeField], r[bodyField])
	//case r[operatorField] == "DELE":
	//	go op.Dele(r[pathField], r[bytesField], r[returnCodeField]) // nolint
	//case r[operatorField] == "HEAD":
	//	go op.Head(r[pathField], r[bytesField], r[returnCodeField]) // nolint
	default:
		log.Printf("unimplemented operation %s in %v, ignored\n", r[operatorField], r)
	}
	return false
}

// getWork gets stuff for worker to do
func getWork() ([]string, bool) {
	var r []string
	var ok bool

	select {
	case <-closed:
		// peculiar to increasing load test, refactor
		if conf.Debug {
			log.Print("pipe closed, no more requests to process.\n")
		}
		return nil, true
	case r, ok = <-pipe:
		if !ok {
			// We're at eof
			return nil, true
		}
		if conf.Debug {
			log.Printf("got %v\n", r)
		}
		return r, false
	}
}

// waitForChange waits for the tail of a file to be written to
// cargo courtesy Satyajit Ranjeev, http://satran.in/2017/11/15/Implementing_tails_follow_in_go.html
func waitForChange(w *fsnotify.Watcher) error {
	for {
		select {
		case event := <-w.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				return nil
			}
		case err := <-w.Errors:
			return err
		}
	}
}

// reportPerformance in standard format
func reportPerformance(initial time.Time, latency time.Duration,
	transferTime time.Duration, body []byte, path string,
	rc int, oldRc string) {
	var annotation = ""

	if oldRc != "" {
		old, _ := strconv.Atoi(oldRc)
		if rc != old && old != 0 {
			annotation = fmt.Sprintf(" expected=%d", old)
		}
	}
	fmt.Printf("%s %f %f 0 %d %s %d GET %d %s\n",
		initial.Format("2006-01-02 15:04:05.00000000"),
		latency.Seconds(), transferTime.Seconds(), len(body), path,
		rc, OfferedRate, annotation)
}

// reportRusage reports cpu-seconds, memory and IOPS used
func reportRUsage(name string, start time.Time) {
	var r syscall.Rusage

	err := syscall.Getrusage(syscall.RUSAGE_SELF, &r)
	if err != nil {
		log.Fatal(err)
		log.Printf("%s %s %d no resource usage available\n",
			start.Format("2006-01-02 15:04:05.000"), name, os.Getpid())
		return
	}
	log.Printf("#date      time         name        pid  utime stime maxrss inblock outblock\n")
	log.Printf("%s %s %d %f %f %d %d %d\n", start.Format("2006-01-02 15:04:05.000"),
		name, os.Getpid(), seconds(r.Utime), seconds(r.Stime), r.Maxrss*1024, r.Inblock, r.Oublock)
}

// seconds converts a syscall.Timeval to seconds
func seconds(t syscall.Timeval) float64 {
	return float64(time.Duration(t.Sec)*time.Second+time.Duration(t.Usec)*time.Microsecond) / float64(time.Second)
}
