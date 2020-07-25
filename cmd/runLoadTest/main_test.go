package main

import (
	"fmt"
	"github.com/davecb/Play-it-Again-Sam/pkg/loadTesting"
	"log"
	"os"
	"testing"
	"time"
)

const (
	// Fail if the code takes more than internalTime:
	internalTime = time.Microsecond * 500
	budgetedTime = 100 * time.Millisecond + internalTime
)

func TestTimeBudget(t *testing.T) {
	var tests = []struct {
		debug bool
	}{
		// turning on debug will make it too slow
		{false},   // This will pass
		// {true}, ,   // This will fail
	}
	for _, test := range tests {
		budgetTest(test.debug, t)
	}
}

func budgetTest(debug bool, t *testing.T) {

	var tpsTarget, progressRate, stepDuration, startTps int
	var startFrom, runFor int
	var bufSize int64
	var s3Bucket, s3Key, s3Secret string
	var verbose, crash, akamaiDebug bool
	var serial, cache, tail bool
	var strip, hostHeader string
	var headerMap = make(map[string]string)
	var err error

	initial := time.Now()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs
	runFor = 1                                           // 1 record
	tpsTarget = 1                                        // at 1 TPS
	debug = false
	verbose = false
	filename := "timeBudget.csv"
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening %s: %s, halting.", filename, err)
	}
	defer f.Close() // nolint
	baseURL := ""

	loadTesting.RunLoadTest(f, filename, startFrom, runFor,
		tpsTarget, progressRate, startTps, baseURL,
		loadTesting.Config{
			Verbose:      verbose,
			Debug:        debug,
			Crash:        crash,
			AkamaiDebug:  akamaiDebug,
			Serialize:    serial,
			Cache:        cache,
			Tail:         tail,
			Protocol:     loadTesting.TimeBudgetProtocol,
			S3Key:        s3Key,
			S3Secret:     s3Secret,
			S3Bucket:     s3Bucket,
			Strip:        strip,
			Timeout:      terminationTimeout,
			StepDuration: stepDuration,
			HostHeader:   hostHeader,
			HeaderMap:    headerMap,
			R:            true,
			W:            false,
			BufSize:      bufSize,
		})

	totalTime := time.Since(initial)
	if totalTime >= budgetedTime {
		t.Error(fmt.Sprintf("Get took %f seconds, more than %v, error\n",
			totalTime.Seconds(), budgetedTime))
	} else {
		log.Printf("Get took %f seconds, within %v\n",
			totalTime.Seconds(), budgetedTime)
	}
}

