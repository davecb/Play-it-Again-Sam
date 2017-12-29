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
	"gopkg.in/fsnotify.v1"
	//"google.golang.org/genproto/googleapis/watcher/v1"
	"strconv"
	"syscall"
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
	Put(path, size, oldRc string) error
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
	Verbose      bool   // Extra info about requests
	Debug        bool   // Extra infor about program
	Crash        bool   // Halt on any error
	Serialize    bool   // FIXME semi-evil hack
	Cache        bool   // allow caching
	Tail         bool   // tail a log
	AkamaiDebug  bool   // add Akamai debug headers
	Protocol     int    // rest, 23 or filesystem FIXME?
	S3Bucket     string // s3-specific options
	S3Key        string
	S3Secret     string
	Strip        string
	Timeout      time.Duration     // time to wait at end
	StepDuration int               // duration of a test step
	HostHeader   string            // add a Host: header
	HeaderMap    map[string]string // one or more key:value headers
	R            bool              // read tests allowed
	W            bool              // write tests allowed
	BufSize      int64             // max size of written file
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
	default:
		log.Fatalf("protocol %d not implemented yet", conf.Protocol)
	}
	// FIXME for write: defer os.Remove(junkDataFile)

	go workSelector(f, filename, fromTime, forTime, pipe)             // which pipes work to ...
	go generateLoad(pipe, tpsTarget, progressRate, startTps, baseURL) // which then writes to "alive"
waiter:
	for {
		select {
		case <-alive:
			processed++

		case <-time.After(time.Second * conf.Timeout):
			log.Printf("%d records processed\n", processed)
			log.Printf("No activity after %d seconds, halting normally.\n",
				conf.Timeout)
			break waiter
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
	recNo, pipe := copyToPipe(runFor, r, filename, pipe, watcher)
	log.Printf("EOF: loaded %d records, closing input pipe\n", recNo)
	close(pipe)
}

// copyToPipe sends work to the workers
func copyToPipe(runFor int, r *csv.Reader, filename string, pipe chan []string, watcher *fsnotify.Watcher) (int, chan []string) {
	// From there, copy to pipe

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
			log.Fatalf("Fatal error mid-way in %s: %s\n", filename, err)
		}
		if len(record) != 9 {
			// Warning: this discards real-time part-records
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
	if conf.BufSize > 0 {
		// create a buffer full of random bytes
		log.Printf("-rw and -wo tests are not yet supported, buffer size of %d ignored\n", conf.BufSize)
	}

	fmt.Print("#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc op\n")
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
		done := doWork()
		if done {
			return
		}
	}
}

// work is the thing that happens each second.
func doWork() bool {
	var r []string

	select {
	case <-closed:
		if conf.Debug {
			log.Print("pipe closed, no more requests to process.\n")
		}
		return true
	case r = <-pipe:
		//log.Printf("got %v\n", r)
	}

	switch {
	case r == nil:
		//log.Print("worker reached EOF, no more requests to send.\n")
		return true
	case len(r) != 9:
		// bad input data, crash please.
		log.Fatalf("number of fields != 9 in %v", r)
	case r[operatorField] == "GET" && conf.R:
		if conf.Serialize {
			// force this NOT to be asynchronous, for long-running load tests only
			op.Get(r[pathField], r[returnCodeField]) // nolint, ignore return value
		} else {
			go op.Get(r[pathField], r[returnCodeField]) // nolint
		}
	//case r[operatorField] == "PUT" && conf.W:
	//	go op.Put(r[pathField], r[bytesField], r[returnCodeField]) // nolint
	//case r[operatorField] == "DELE":
	//	go op.Dele(r[pathField], r[bytesField], r[returnCodeField]) // nolint
	//case r[operatorField] == "HEAD":
	//	go op.Head(r[pathField], r[bytesField], r[returnCodeField]) // nolint
	default:
		log.Printf("unimplemented operation %s in %v, ignored\n", r[operatorField], r)
	}
	return false
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
	fmt.Printf("%s %f %f 0 %d %s %d GET%s\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), len(body), path,
		rc, annotation)
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
	fmt.Fprint(os.Stderr, "#date      time         name        pid  utime stime maxrss inblock outblock\n")
	fmt.Fprintf(os.Stderr, "%s %s %d %f %f %d %d %d\n", start.Format("2006-01-02 15:04:05.000"),
		name, os.Getpid(), seconds(r.Utime), seconds(r.Stime), r.Maxrss*1024, r.Inblock, r.Oublock)
}

// seconds converts a syscall.Timeval to seconds
func seconds(t syscall.Timeval) float64 {
	return float64(time.Duration(t.Sec)*time.Second+time.Duration(t.Usec)*time.Microsecond) / float64(time.Second)
}
