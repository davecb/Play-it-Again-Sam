// TestTimeBudget tests that the system under test completes within a specified time budget.
// The test runs a set of subtests, each with a different debug configuration.
// The budgetTest function is responsible for running the system under test and verifying
// that the total execution time is within the budgeted time.
func TestTimeBudget(t *testing.T)

// budgetTest runs the system under test and verifies that the total execution time
// is within the budgeted time. If the execution time exceeds the budget, an error
// is reported.
func budgetTest(debug bool, t *testing.T)

// systemUnderTest is the function that contains the code being tested. It sets up
// the necessary configuration and then calls the RunLoadTest function from the
// loadTesting package to execute the load test.
func systemUnderTest(debug bool)
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
	internalTime = time.Millisecond
	budgetedTime = 100 * time.Millisecond + internalTime
)

// TestTimeBudget tests that the system under test completes within a specified time budget.
// The test runs a set of subtests, each with a different debug configuration.
// The budgetTest function is responsible for running the system under test and verifying
// that the total execution time is within the budgeted time.
func TestTimeBudget(t *testing.T) {
	var tests = []struct {
		debug bool
	}{
		// turning on debug will make it too slow
		{false},   // This will pass
		//{true},    // This will fail
		// turning on both will trigger a bug
	}
	for _, test := range tests {
		budgetTest(test.debug, t)
	}
}

// budgetTest runs the system under test and verifies that the total execution time
// is within the budgeted time. If the execution time exceeds the budget, an error
// is reported.
func budgetTest(debug bool, t *testing.T) {

	initial := time.Now()
    systemUnderTest(debug)
    totalTime := time.Since(initial)
	if totalTime >= budgetedTime {
		t.Error(fmt.Sprintf("Get took %f seconds, more than %v, error\n",
			totalTime.Seconds(), budgetedTime))
	} else {
		log.Printf("Get took %f seconds, within %v\n",
			totalTime.Seconds(), budgetedTime)
	}
	time.Sleep(10 * time.Second) // let pipes drain
}

// systemUnderTest is the function that contains the code being tested. It sets up
// the necessary configuration and then calls the RunLoadTest function from the
// loadTesting package to execute the load test.
func systemUnderTest(debug bool) {
	var tpsTarget, progressRate, stepDuration, startTps int
	var startFrom, runFor int
	var bufSize int64
	var s3Bucket, s3Key, s3Secret string
	var verbose, crash, akamaiDebug bool
	var serial, cache, tail bool
	var strip, hostHeader string
	var headerMap = make(map[string]string)
	var err error

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs
	runFor = 1                                           // 1 record
	tpsTarget = 1                                        // at 1 TPS
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
}

