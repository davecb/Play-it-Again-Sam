package loadTesting

import (
	"log"
	"net/http"
	"time"
)

// timeBudgetProto satisfies operation by doing timed no-ops.
type timeBudgetProto struct {
	prefix string
}

// Init does nothing
func (p timeBudgetProto) Init() {
	if conf.Debug {
		log.Printf("in timeBudgetProto.Init()\n")
	}
}

// Get does a GET that should take one tenth of a second
func (p timeBudgetProto) Get(path string, oldRc string) {
	if conf.Debug {
		log.Printf("in timeBudgetProto.Get(%s)\n", path)
	}

	initial := time.Now() // Response time starts
	// wait a tenth of a second
	time.Sleep(100 * time.Millisecond)
	latency := time.Since(initial) // Latency ends
	totalTime := time.Since(initial)
	transferTime := totalTime - latency // Transfer time ends

	reportPerformance(initial, latency, transferTime, []byte(""), path, http.StatusOK, oldRc)
}

// Put does a PUT that should take one tenth of a second
func (p timeBudgetProto) Put(path, size, oldRc string) {

	if conf.Debug {
		log.Printf("in timeBudgetProto.Put(%s, %s)\n", path, size)
	}
	initial := time.Now() // Response time starts
	// wait a tenth of a second
	time.Sleep(1000 * time.Millisecond)
	latency := time.Since(initial) // Latency ends
	totalTime := time.Since(initial)
	transferTime := totalTime - latency // Transfer time ends

	reportPerformance(initial, latency, transferTime, []byte(""), path, http.StatusOK, oldRc)
}

func (p timeBudgetProto) Post(path, size, oldRC, body string) {
	log.Fatalf("POST is unimplemented\n")
}
