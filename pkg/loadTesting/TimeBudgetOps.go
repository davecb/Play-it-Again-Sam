package loadTesting

import (
	"log"
	"net/http"
	"time"
)


// TimeBudgetProto satisfies operation by doing timed no-ops.
type TimeBudgetProto struct {
	prefix string
}

// Init does nothing
func (p TimeBudgetProto) Init() {}

// Get does a GET that should take one tenth of a second
func (p TimeBudgetProto) Get(path string, oldRc string) {
	if conf.Debug {
		log.Printf("in rest.Get(%s)\n", path)
	}

	initial := time.Now() // Response time starts
	// wait a tenth of a second
	time.Sleep(100 * time.Millisecond)
	latency := time.Since(initial) // Latency ends
	totalTime := time.Since(initial)
	transferTime := totalTime - latency // Transfer time ends

	reportPerformance(initial, latency, transferTime, []byte(""), path, http.StatusOK, oldRc)
	close(alive) // This forces an immediate exit
}

// Put does a PUT that should take one tenth of a second
func (p TimeBudgetProto) Put(path, size, oldRc string) {

	if conf.Debug {
		log.Printf("in rest.Put(%s, %s)\n", path, size)
	}
	initial := time.Now() // Response time starts
	// wait a tenth of a second
	time.Sleep(1000 * time.Millisecond)
	latency := time.Since(initial) // Latency ends
	totalTime := time.Since(initial)
	transferTime := totalTime - latency // Transfer time ends


	reportPerformance(initial, latency, transferTime, []byte(""), path, http.StatusOK, oldRc)
	close(alive)
}

