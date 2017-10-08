package loadTesting

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

// Tuning for large loads. With this, when we start reporting network
// errors, then we've overloaded somebody. Previously we bottlenecked
// on closed but not recycled sockets'
// For calvin this doesn't change the results up to 240, but
// then the former setup failed.
const (
	MaxIdleConnections int = 100
	RequestTimeout     int = 0
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: MaxIdleConnections,
	},
	Timeout: time.Duration(RequestTimeout) * time.Second,
}

// RestGet does a GET from an http target and times it
func RestGet(baseURL, path string) {
	if conf.Debug {
		log.Printf("in RestGet(%s,%s)\n", baseURL, path)
	}
	req, err := http.NewRequest("GET", baseURL+"/"+path, nil)
	if err != nil {
		// try running right through this
		// log.Fatalf("error creating http request, %v: halting.\n", err)
		log.Printf("error creating http request, %v: continuing.\n", err)
		dumpRequest(req)
		fmt.Printf("%s 0 0 0 0 %s %d GET\n",
			time.Now().Format("2006-01-02 15:04:05.000"), path, -1)
		alive <- true
		return
	}
	if !conf.Cache {
		req.Header.Add("cache-control", "no-cache")
	}
	if conf.HostHeader != "" {
		req.Host = conf.HostHeader
		// Go disfeature: host is special,
		// See also https://github.com/golang/go/issues/7682
		req.Header.Add("Host", conf.HostHeader)
	}

	initial := time.Now() // Response time starts
	resp, err := httpClient.Do(req)
	latency := time.Since(initial) // Latency ends
	if err != nil {
		//log.Fatalf("error getting http response, %v: halting.\n", err)
		// try running right through this
		log.Printf("error getting http response, %v: continuing.\n", err)
		dumpRequest(req)
		dumpResponse(resp)
		fmt.Printf("%s %f %f 0 %d %s %d GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			latency.Seconds(), 0.0, 0, path, -2)
		alive <- true
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	transferTime := time.Since(initial) - latency // Transfer time ends
	defer resp.Body.Close()                       // nolint
	if err != nil {
		//log.Fatalf(`error reading http response, "%v": halting.\n`, err)
		// try running right through this
		log.Printf("error reading http response, %v: continuing.\n", err)
		dumpRequest(req)
		dumpResponse(resp)
		fmt.Printf("%s %f %f 0 %d %s %d GET\n",
			time.Now().Format("2006-01-02 15:04:05.000"),
			latency.Seconds(), transferTime.Seconds(), 0, path, -3)
		alive <- true
		return
	}
	if conf.Verbose || firstDigit(resp.StatusCode) == 5 {
		// dump if its not a 200 OK, etc.
		dumpRequest(req)
		dumpResponse(resp)
	}

	fmt.Printf("%s %f %f 0 %d %s %d GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), len(body), path, resp.StatusCode)
	alive <- true
}

// RestPut does an ordinary REST (not ceph or s3) put operation.
func RestPut(baseURL, path string, size int64) error {

	if conf.Debug {
		log.Printf("putting %s\n", baseURL+"/"+path)
	}
	if size <= 0 {
		fmt.Printf("%s 0 0 0 %d %s %d PUT\n",
			time.Now().Format("2006-01-02 15:04:05.000"),
			size, path, 411) // 411 means "length required"
		alive <- true
		return nil
	}
	// make sure we have a dummy file
	fp, err := os.Open(junkDataFile)
	if err != nil {
		return fmt.Errorf("can't open %q", junkDataFile)
	}
	defer fp.Close() // nolint

	initial := time.Now() // Response time starts
	// do put
	req, err := http.NewRequest("PUT", baseURL+"/"+path, io.LimitReader(fp, size))
	if err != nil {
		dumpRequest(req)
		log.Fatalf("error creating http request, %v: halting.\n", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		// Timeouts and bad parameters will trigger this case. FIXME, continue?
		dumpRequest(req)
		log.Fatalf("error getting http response, %v: halting.\n", err)
	}
	latency := time.Since(initial) // Response time ends
	contents, err := ioutil.ReadAll(resp.Body)
	transferTime := time.Since(initial) - latency // Transfer time ends
	defer resp.Body.Close()                       // nolint
	if err != nil {
		dumpRequest(req)
		dumpResponse(resp)
		// FIXME, continue?
		log.Fatalf(`error reading http response, "%v": halting.\n`, err)
	}
	if conf.Verbose || firstDigit(resp.StatusCode) == 5 {
		// dump if its not a 200 OK, etc.
		dumpRequest(req)
		dumpResponse(resp)
	}

	fmt.Printf("%s %f %f 0 %d %s %d PUT\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), len(contents), path, resp.StatusCode)
	alive <- true
	return nil
}

// return the first digit of a status code, where
// 1 - informational
// 2 - success
// 3 - partial success
// 4 - temporary failure
// 5 - permanent failure
func firstDigit(i int) int {
	return i / 100
}

// dumpRequest provides extra information about an http request if it can
func dumpRequest(req *http.Request) {
	var dump []byte
	var err error

	if req == nil {
		log.Print("Request: <nil>\n")
		return
	}
	dump, err = httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatalf("fatal error dumping http request, %v: halting.\n", err)
	}
	log.Printf("Request: \n%s\n", dump)
}

// dumpResponse provides extra information about an http response
func dumpResponse(resp *http.Response) {
	if resp == nil {
		log.Print("Response: <nil>\n")
		return
	}
	contents, err := httputil.DumpResponse(resp, true)
	if err != nil {
		// Error dumping http response, trying without body.
		// That avoids errors if length headers are wrong
		contents, err = httputil.DumpResponse(resp, false)
		if err != nil {
			log.Fatalf("error dumping http response, %v: halting.\n", err)
		}
	}
	log.Print("Response:\n")
	log.Printf("    Length: %d\n", len(string(contents)))
	shortDescr, _ := codeDescr(resp.StatusCode)
	log.Printf("    Status code: %d %s\n", resp.StatusCode, shortDescr)
	hdr := resp.Header
	for key, value := range hdr {
		log.Println("   ", key, ":", value)
	}
	log.Printf("    Contents: \"\n%s\"\n", string(contents))
}
