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
	"strconv"
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
	dateField = iota // nolint
	timeField
	latencyField      // nolint
	transferTimeField // nolint
	thinkTimeField    // nolint
	bytesField
	pathField
	returnCodeField
	operatorField
)

// Config contains all the optional parameters.
type Config struct {
	Verbose      bool
	Protocol     int
	Strip        string
	Timeout      time.Duration
	StepDuration int
}

// Verbose turns on extra information for understanding the test
var verbose bool

// Protocol indicates which of the above is in use
var protocol int

// Strip is the string to Strip from paths, or ""
var strip string

// Timeout is the delay before shutting dowsn
var timeout time.Duration

// stepDurayion, typically 10 or 30
var stepDuration int

var random = rand.New(rand.NewSource(42))
var pipe = make(chan []string, 100)
var alive = make(chan bool, 1000)

///var bucketName = "loadtest"
var junkDataFile = "/tmp/LoadTestJunkDataFile"

const size = 396759652 // FIXME, this is a heuristic

// RunLoadTest does whatever main figured out that the caller wanted.
func RunLoadTest(f io.Reader, filename string, fromTime, forTime int,
	tpsTarget, progressRate int, baseURL string, conf Config) {
	var processed = 0

	//log.Printf("new runLoadTest(f, tpsTarget=%d, progressRate=%d, fromTime=%d, forTime=%d, baseURL=%s)\n",
	//	tpsTarget, progressRate, fromTime, forTime, baseURL)
	// get settings from conf parameter
	verbose = conf.Verbose
	protocol = conf.Protocol
	strip = conf.Strip
	timeout = conf.Timeout

	doPrepWork(baseURL)           // Named init(), creates junkDataFile
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

// workSelector pipes a selection from a file to the workers
func workSelector(f io.Reader, filename string, startFrom, runFor int, strip string, pipe chan []string) {
	var firstTime, thisSecond time.Time

	// log.Printf("in workSelector(r, %s, startFrom=%d runFor=%d, pipe)\n", filename, startFrom, runFor)
	r := csv.NewReader(f)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1 // ignore differences

	// Skip forward to the staring point
	for i := 0; i < startFrom; i++ {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Fatal error skipping forward in %s: %s\n", filename, err)
		}
		timeString := rec[timeField]
		thisSecond, err = time.Parse("15:04:05", timeString)
		if err != nil {
			log.Fatalf("Fatal error parsing time field in %s: %s\n", rec, err)
		}
		if firstTime.IsZero() {
			// initialize first Time on seeing the first record
			firstTime = thisSecond.Add(time.Second * time.Duration(startFrom))
		}
		if !thisSecond.Before(firstTime) {
			// it's not before the first second, AKA >= first second
			break
		}

	}

	// From there, copy to pipe
	for i := 0; i < runFor; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Fatal error mid-way in %s: %s\n", filename, err)
		}
		if strip != "" {
			record[pathField] = strings.Replace(record[pathField], strip, "", 1)
		}
		// log.Printf("writing %v to pipe\n", record)
		pipe <- record
	}
	//log.Print("closing pipe\n")
	close(pipe)
}

// generateLoad starts progressRate new threads every 10 seconds until we hit progressRate
func generateLoad(pipe chan []string, tpsTarget, progressRate int, urlPrefix string) {
	//log.Printf("generateLoad(pipe, tpsTarget=%d, progressRate=%d, from, for, file=%s\n",
	//	tpsTarget, progressRate, filename)

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
		// FIXME allow passed-in stepDuration
		for range time.Tick(10 * time.Second) { // nolint
			//start another progressRate of workers
			rate += progressRate
			if rate > tpsTarget {
				// OK, we're past the range, quit.
				log.Print("ok, all done\n")
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
		for i := 0; i < tpsTarget; i++ {
			// start a worker
			go worker(pipe, closed, urlPrefix)
		}
	case tpsTarget <= 0:
		log.Fatal("A zero or negative tps target is not meaningfull, halting\n")
	}
}

// worker reads and executes a task every second until it hits eof
func worker(pipe chan []string, closed chan bool, urlPrefix string) {
	//log.Print("started a worker\n")
	// wait a random fraction of one second before looping, for randomness.
	time.Sleep(time.Duration(random.Float64() * float64(time.Second)))

	for range time.Tick(1 * time.Second) { // nolint
		var r []string

		select {
		case <-closed:
			return
		case r = <-pipe:
			//log.Printf("got %v\n", r)
		}

		switch {
		case r == nil:
			//log.Print("worker at EOF\n")
			return
		case len(r) != 9:
			// bad input data, crash please.
			log.Fatalf("number of fields != 9 in %v", r)
		case r[operatorField] == "GET":
			go getJunkFile(urlPrefix, r[pathField])
		case r[operatorField] == "PUT":
			fileSize, err := strconv.ParseInt(r[bytesField], 10, 64)
			if err != nil {
				// fail the run if given bad input data here, too
				log.Fatalf("could not parse size field in %v\n", r)
			}
			go putJunkFile(urlPrefix, r[pathField], fileSize)
		default:
			log.Fatal("operations other than GET and PUT are not implemented yet\n")
		}
	}
}

// putJunkFile sends a specified number of bytes from /dev/urandom via a PUT
func putJunkFile(baseURL, path string, size int64) {
	//log.Printf("in putJunkFile(%s, %s, %d)\n", baseURL, path, size)

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
	//log.Printf("in getJunkFile(%s, %s), Protocol=%v\n", baseURL, path, Protocol)

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
	MustCreateFilesystemFile(junkDataFile, size)
	switch protocol {
	case S3Protocol:
		AmazonS3Prep(baseURL)
	}
}
